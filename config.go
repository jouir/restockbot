package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// Config to store JSON configuration
type Config struct {
	TwitterConfig  `json:"twitter"`
	TelegramConfig `json:"telegram"`
	URLs           []string `json:"urls"`
	IncludeRegex   string   `json:"include_regex"`
	ExcludeRegex   string   `json:"exclude_regex"`
	BrowserAddress string   `json:"browser_address"`
}

// TwitterConfig to store Twitter API secrets
type TwitterConfig struct {
	ConsumerKey       string              `json:"consumer_key"`
	ConsumerSecret    string              `json:"consumer_secret"`
	AccessToken       string              `json:"access_token"`
	AccessTokenSecret string              `json:"access_token_secret"`
	Hashtags          []map[string]string `json:"hashtags"`
}

// TelegramConfig to store Telegram API key
type TelegramConfig struct {
	Token       string `json:"token"`
	ChatID      int64  `json:"chat_id"`
	ChannelName string `json:"channel_name"`
}

// NewConfig creates a Config struct
func NewConfig() *Config {
	return &Config{}
}

// Read Config from configuration file
func (c *Config) Read(file string) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	jsonFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return err
	}
	return nil
}

// HasTwitter returns true when Twitter has been configured
func (c *Config) HasTwitter() bool {
	return (c.TwitterConfig.AccessToken != "" && c.TwitterConfig.AccessTokenSecret != "" && c.TwitterConfig.ConsumerKey != "" && c.TwitterConfig.ConsumerSecret != "")
}

// HasTelegram returns true when Telegram has been configured
func (c *Config) HasTelegram() bool {
	return c.TelegramConfig.Token != "" && (c.TelegramConfig.ChatID != 0 || c.TelegramConfig.ChannelName != "")
}
