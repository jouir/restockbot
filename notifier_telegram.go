package main

import (
	"fmt"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TelegramMessage to store relationship between a Product and a Telegram notification
type TelegramMessage struct {
	gorm.Model
	MessageID  int `gorm:"not null;unique"`
	ProductURL string
	Product    Product `gorm:"foreignKey:ProductURL"`
}

// TelegramNotifier to manage notifications to Twitter
type TelegramNotifier struct {
	db            *gorm.DB
	bot           *telegram.BotAPI
	chatID        int64
	channelName   string
	enableReplies bool
}

// NewTelegramNotifier to create a Notifier with Telegram capabilities
func NewTelegramNotifier(config *TelegramConfig, db *gorm.DB) (*TelegramNotifier, error) {
	// create table
	err := db.AutoMigrate(&TelegramMessage{})
	if err != nil {
		return nil, err
	}

	// create client
	bot, err := telegram.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}
	log.Debugf("connected to telegram as %s", bot.Self.UserName)

	return &TelegramNotifier{
		db:            db,
		bot:           bot,
		chatID:        config.ChatID,
		channelName:   config.ChannelName,
		enableReplies: config.EnableReplies,
	}, nil
}

// NotifyWhenAvailable create a Telegram message for announcing that a product is available
// implements the Notifier interface
func (n *TelegramNotifier) NotifyWhenAvailable(shopName string, productName string, productPrice float64, productCurrency string, productURL string) error {
	// TODO: check if message exists in the database to avoid flood

	// send message to telegram
	formattedPrice := formatPrice(productPrice, productCurrency)
	rawMessage := `*Name:* %s
*Retailer:* %s
*Price:* %s
*URL*: [go to website](%s)
*Date/Time:* %s`
	message := fmt.Sprintf(rawMessage, productName, shopName, formattedPrice, productURL, time.Now().UTC().Format("2006-01-02 15:04:05 (-0700)"))
	messageID, err := n.sendMessage(message, 0)
	if err != nil {
		return err
	}

	// save telegram message to database
	m := TelegramMessage{MessageID: messageID, ProductURL: productURL}
	trx := n.db.Create(&m)
	if trx.Error != nil {
		return fmt.Errorf("failed to save telegram message %d to database: %s", m.MessageID, trx.Error)
	}
	log.Debugf("telegram message %d saved to database", m.MessageID)

	return nil
}

// NotifyWhenNotAvailable create a Telegram message replying to the NotifyWhenAvailable message to say it's gone
// implements the Notifier interface
func (n *TelegramNotifier) NotifyWhenNotAvailable(productURL string, duration time.Duration) error {
	// find message in the database
	var m TelegramMessage
	trx := n.db.Where(TelegramMessage{ProductURL: productURL}).First(&m)
	if trx.Error != nil {
		return fmt.Errorf("failed to find telegram message in database for product with url %s: %s", productURL, trx.Error)
	}
	if m.MessageID == 0 {
		log.Warnf("telegram message for product with url %s not found, skipping close notification", productURL)
		return nil
	}

	if n.enableReplies {
		// format message
		text := fmt.Sprintf("And it's gone (%s)", duration)

		// send reply on telegram
		_, err := n.sendMessage(text, m.MessageID)
		if err != nil {
			return fmt.Errorf("failed to reply on telegram: %s", err)
		}
		log.Infof("reply to telegram message %d sent", m.MessageID)
	}

	// remove message from database
	trx = n.db.Unscoped().Delete(&m)
	if trx.Error != nil {
		return fmt.Errorf("failed to remove message %d from database: %s", m.MessageID, trx.Error)
	}
	log.Debugf("telegram message removed from database")
	return nil
}

func (n *TelegramNotifier) sendMessage(text string, reply int) (int, error) {
	log.Debugf("sending message %s to telegram", text)
	var request telegram.MessageConfig
	if n.chatID != 0 {
		request = telegram.NewMessage(n.chatID, text)
	} else {
		request = telegram.NewMessageToChannel(n.channelName, text)
	}
	request.DisableWebPagePreview = true
	request.ParseMode = telegram.ModeMarkdown

	if reply != 0 {
		request.ReplyToMessageID = reply
	}

	response, err := n.bot.Send(request)
	if err != nil {
		return 0, err
	}
	log.Infof("message %d sent to telegram", response.MessageID)
	return response.MessageID, nil
}
