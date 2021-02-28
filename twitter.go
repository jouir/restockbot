package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Tweet to store relationship between a Product and a Twitter notification
type Tweet struct {
	gorm.Model
	TweetID    int64
	ProductURL string
	Product    Product `gorm:"foreignKey:ProductURL"`
}

// TwitterNotifier to manage notifications to Twitter
type TwitterNotifier struct {
	db          *gorm.DB
	client      *twitter.Client
	user        *twitter.User
	hashtagsMap []map[string]string
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

	return &TwitterNotifier{client: client, user: user, hashtagsMap: c.Hashtags, db: db}, nil
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

// formatPrice using internationalization rules
// euro sign is placed after the value
// default the currency, or symbol if applicable, is placed before the value
func formatPrice(value float64, currency string) string {
	switch {
	case currency == "EUR":
		return fmt.Sprintf("%.2f€", value)
	case currency == "USD":
		return fmt.Sprintf("$%.2f", value)
	default:
		return fmt.Sprintf("%s%.2f", currency, value)
	}
}

// NotifyWhenAvailable create a Twitter status for announcing that a product is available
// implements the Notifier interface
func (c *TwitterNotifier) NotifyWhenAvailable(shopName string, productName string, productPrice float64, productCurrency string, productURL string) error {
	// format message
	formattedPrice := formatPrice(productPrice, productCurrency)
	hashtags := c.buildHashtags(productName)
	message := fmt.Sprintf("%s: %s for %s is available at %s %s", shopName, productName, formattedPrice, productURL, hashtags)

	// create thread
	tweetID, err := c.createTweet(message)
	if err != nil {
		return fmt.Errorf("failed to create new twitter thread: %s", err)
	}
	log.Infof("tweet %d sent", tweetID)

	// save thread to database
	t := Tweet{TweetID: tweetID, ProductURL: productURL}
	trx := c.db.Create(&t)
	if trx.Error != nil {
		return fmt.Errorf("failed to save tweet %d to database: %s", t.TweetID, trx.Error)
	}
	log.Debugf("tweet %d saved to database", t.TweetID)
	return nil
}

// NotifyWhenNotAvailable create a Twitter status replying to the NotifyWhenAvailable status to say it's over
// implements the Notifier interface
func (c *TwitterNotifier) NotifyWhenNotAvailable(productURL string, duration time.Duration) error {
	// find Tweet in the database
	var tweet Tweet
	trx := c.db.Where(Tweet{ProductURL: productURL}).First(&tweet)
	if trx.Error != nil {
		return fmt.Errorf("failed to find tweet in database for product with url %s: %s", productURL, trx.Error)
	}
	if tweet.TweetID == 0 {
		log.Warnf("tweet for product with url %s not found, skipping close notification", productURL)
		return nil
	}

	// format message
	message := fmt.Sprintf("And it's over (%s)", duration)

	// close thread on twitter
	_, err := c.replyToTweet(tweet.TweetID, message)
	if err != nil {
		return fmt.Errorf("failed to create reply tweet: %s", err)
	}
	log.Infof("reply to tweet %d sent", tweet.TweetID)

	// remove tweet from database
	trx = c.db.Unscoped().Delete(&tweet)
	if trx.Error != nil {
		return fmt.Errorf("failed to remove tweet %d from database: %s", tweet.TweetID, trx.Error)
	}
	log.Debugf("tweet removed from database")

	return nil
}
