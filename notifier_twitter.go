package main

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// maximum number of characters a tweet can support
const tweetMaxSize = 280

// Tweet to store relationship between a Product and a Twitter notification
type Tweet struct {
	gorm.Model
	TweetID     int64   `gorm:"not null;unique"`
	Hash        string  `gorm:"unique"`
	LastTweetID int64   `gorm:"index"`
	Counter     int64   `gorm:"not null;default:1"`
	ProductURL  string  `gorm:"index"`
	Product     Product `gorm:"not null;references:URL"`
}

// TwitterNotifier to manage notifications to Twitter
type TwitterNotifier struct {
	db            *gorm.DB
	client        *twitter.Client
	user          *twitter.User
	hashtagsMap   []map[string]string
	enableReplies bool
	retentionDays int
}

// NewTwitterNotifier creates a TwitterNotifier
func NewTwitterNotifier(c *TwitterConfig, db *gorm.DB) (*TwitterNotifier, error) {
	// create table
	err := db.AutoMigrate(&Tweet{})
	if err != nil {
		return nil, err
	}

	// create twitter client
	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	// verify credentials at least once
	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}
	log.Debugf("connected to twitter as @%s", user.ScreenName)

	notifier := &TwitterNotifier{
		client:        client,
		user:          user,
		hashtagsMap:   c.Hashtags,
		db:            db,
		enableReplies: c.EnableReplies,
		retentionDays: c.Retention,
	}

	// delete old tweets
	if err = notifier.ensureRetention(); err != nil {
		return nil, err
	}

	return notifier, nil

}

// ensureRetention deletes tweets according to the defined retention
func (c *TwitterNotifier) ensureRetention() error {
	if c.retentionDays == 0 {
		log.Debugf("tweet retention not found, skipping database cleanup")
		return nil
	}

	var oldTweets []Tweet
	retentionDate := time.Now().Local().Add(-time.Hour * 24 * time.Duration(c.retentionDays))
	trx := c.db.Where("updated_at < ?", retentionDate).Find(&oldTweets)
	if trx.Error != nil {
		return fmt.Errorf("cannot find twitter old statuses: %s", trx.Error)
	}
	for _, t := range oldTweets {
		log.Debugf("twitter old status found with id %d", t.TweetID)
		if trx = c.db.Unscoped().Delete(&t); trx.Error != nil {
			log.Warnf("cannot remove old tweet %d: %s", t.TweetID, trx.Error)
		} else {
			log.Infof("twitter old status %d removed from database", t.TweetID)
		}
	}
	return nil
}

// create a brand new tweet
func (c *TwitterNotifier) createTweet(message string) (int64, error) {
	tweet, _, err := c.client.Statuses.Update(message, nil)
	if err != nil {
		return 0, err
	}
	log.Debugf("twitter status %d created: %s", tweet.ID, tweet.Text)
	return tweet.ID, nil
}

// reply to another tweet
func (c *TwitterNotifier) replyToTweet(tweetID int64, message string) (int64, error) {
	message = fmt.Sprintf("@%s %s", c.user.ScreenName, message)
	tweet, _, err := c.client.Statuses.Update(message, &twitter.StatusUpdateParams{InReplyToStatusID: tweetID})
	if err != nil {
		return 0, nil
	}
	log.Debugf("twitter status %d created: %s", tweet.ID, tweet.Text)
	return tweet.ID, nil
}

// parse product name to build a list of hashtags
func (c *TwitterNotifier) buildHashtags(productName string) string {
	productName = strings.ToLower(productName)
	for _, rule := range c.hashtagsMap {
		for pattern, value := range rule {
			if ok, _ := regexp.MatchString(pattern, productName); ok {
				return value
			}
		}
	}
	return ""
}

// NotifyWhenAvailable create a Twitter status for announcing that a product is available
// implements the Notifier interface
func (c *TwitterNotifier) NotifyWhenAvailable(shopName string, productName string, productPrice float64, productCurrency string, productURL string) error {
	// format message
	hashtags := c.buildHashtags(productName)
	message := formatAvailableTweet(shopName, productName, productPrice, productCurrency, productURL, hashtags, 0)

	// compute message checksum to avoid duplicates
	var tweet Tweet
	hash := fmt.Sprintf("%x", md5.Sum([]byte(message)))
	trx := c.db.Where(Tweet{Hash: hash}).First(&tweet)
	if trx.Error != nil && trx.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("could not search for tweet with hash %s for product '%s': %s", hash, productURL, trx.Error)
	}

	if trx.Error == gorm.ErrRecordNotFound {

		// tweet has not been sent in the past
		// create thread
		tweetID, err := c.createTweet(message)
		if err != nil {
			return fmt.Errorf("could not create new twitter thread for product '%s': %s", productURL, err)
		}
		log.Infof("tweet %d sent for product '%s'", tweetID, productURL)

		// save thread to database
		tweet = Tweet{TweetID: tweetID, ProductURL: productURL, Hash: hash, Counter: 1}
		trx = c.db.Create(&tweet)
		if trx.Error != nil {
			return fmt.Errorf("could not save tweet %d to database for product '%s': %s", tweet.TweetID, productURL, trx.Error)
		}
		log.Debugf("tweet %d saved to database", tweet.TweetID)

	} else {

		// tweet already has been sent in the past
		// creating new thread with a counter
		tweet.Counter++
		message = formatAvailableTweet(shopName, productName, productPrice, productCurrency, productURL, hashtags, tweet.Counter)
		tweetID, err := c.createTweet(message)
		if err != nil {
			return fmt.Errorf("could not create new twitter thread for product '%s': %s", productURL, err)
		}
		log.Infof("tweet %d sent for product '%s'", tweetID, productURL)

		// save thread to database
		tweet.LastTweetID = tweetID
		if trx = c.db.Save(&tweet); trx.Error != nil {
			return fmt.Errorf("could not save tweet %d to database for product '%s': %s", tweet.TweetID, productURL, trx.Error)
		}
		log.Debugf("tweet %d saved to database", tweet.TweetID)
	}

	return nil
}

// formatAvailableTweet creates a message based on product characteristics
func formatAvailableTweet(shopName string, productName string, productPrice float64, productCurrency string, productURL string, hashtags string, counter int64) string {
	// format message
	formattedPrice := formatPrice(productPrice, productCurrency)
	message := fmt.Sprintf("%s: %s for %s is available at %s %s", shopName, productName, formattedPrice, productURL, hashtags)
	if counter > 1 {
		message = fmt.Sprintf("%s (%d)", message, counter)
	}

	// truncate tweet if too big
	if utf8.RuneCountInString(message) > tweetMaxSize {
		messageWithoutProduct := fmt.Sprintf("%s:  for %s is available at %s %s", shopName, formattedPrice, productURL, hashtags)
		if counter > 1 {
			messageWithoutProduct = fmt.Sprintf("%s (%d)", messageWithoutProduct, counter)
		}
		// maximum tweet size - other characters - additional "…" to say product name has been truncated
		productNameSize := tweetMaxSize - utf8.RuneCountInString(messageWithoutProduct) - 1
		format := fmt.Sprintf("%%s: %%.%ds… for %%s is available at %%s %%s", productNameSize)
		message = fmt.Sprintf(format, shopName, productName, formattedPrice, productURL, hashtags)
		if counter > 1 {
			message = fmt.Sprintf("%s (%d)", message, counter)
		}
	}

	return message
}

// NotifyWhenNotAvailable create a Twitter status replying to the NotifyWhenAvailable status to say it's over
// implements the Notifier interface
func (c *TwitterNotifier) NotifyWhenNotAvailable(productURL string, duration time.Duration) error {
	// find Tweet in the database
	var tweet Tweet
	trx := c.db.Where(Tweet{ProductURL: productURL}).First(&tweet)

	if trx.Error != nil {
		return fmt.Errorf("could not find tweet for product '%s' in the database: %s", productURL, trx.Error)
	}

	if c.enableReplies {
		// format message
		message := fmt.Sprintf("And it's gone (%s)", duration)

		// select tweet to reply
		lastTweetID := CoalesceInt64(tweet.LastTweetID, tweet.TweetID)
		if lastTweetID == 0 {
			return fmt.Errorf("could not find original tweet ID to create reply for product '%s'", productURL)
		}

		// close thread on twitter
		tweetID, err := c.replyToTweet(lastTweetID, message)
		if err != nil {
			return fmt.Errorf("could not close thread on twitter for product '%s': %s", productURL, err)
		}
		log.Infof("reply to tweet %d sent with id %d for product '%s'", lastTweetID, tweetID, productURL)

		// save tweet id on database
		tweet.LastTweetID = tweetID
		if trx = c.db.Save(&tweet); trx.Error != nil {
			return fmt.Errorf("could not save tweet %d to database for product '%s': %s", tweet.TweetID, productURL, trx.Error)
		}
		log.Debugf("tweet %d saved in database", tweet.TweetID)
	} else {
		log.Debugf("twitter replies are disabled, skipping not available notification for '%s'", productURL)
	}

	return nil
}
