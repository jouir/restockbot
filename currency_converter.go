package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CurrencyConverter to cache rates of different currency pairs
type CurrencyConverter struct {
	rates map[string]float64
}

// NewCurrencyConverter to create a CurrencyConverter
func NewCurrencyConverter() *CurrencyConverter {
	return &CurrencyConverter{
		rates: make(map[string]float64),
	}
}

// Convert an amount in a given currency to another currency
// Eventually fetch rate from a remote API then cache the result
func (c *CurrencyConverter) Convert(amount float64, fromCurrency string, toCurrency string) (float64, error) {
	var err error

	if fromCurrency == toCurrency {
		return amount, nil
	}

	// exclude invalid currencies
	if fromCurrency == "" || toCurrency == "" {
		return 0.0, fmt.Errorf("invalid currency pair used for convertion (from='%s', to='%s')", fromCurrency, toCurrency)
	}

	// searching currency pair rate in cache
	pair := fmt.Sprintf("%s%s", strings.ToLower(fromCurrency), strings.ToLower(toCurrency))
	rate, exists := c.rates[pair]
	if !exists {
		// fetching rate from api
		rate, err = c.getRate(fromCurrency, toCurrency)
		if err != nil {
			return 0.0, err
		}
		// store rate in cache
		c.rates[pair] = rate
	}

	return rate * amount, nil
}

// CurrencyResponse to unmarshall JSON response from API
type CurrencyResponse struct {
	Date string  `json:"date"`
	Rate float64 `json:"rate"`
}

// getRate retreives rate from a remote API
func (c *CurrencyConverter) getRate(fromCurrency string, toCurrency string) (float64, error) {
	if fromCurrency == "" || toCurrency == "" {
		return 0.0, fmt.Errorf("invalid currency pair used for convertion (from=%s, to=%s)", fromCurrency, toCurrency)
	}

	// lowering currency names to match the API route
	fromCurrency = strings.ToLower(fromCurrency)
	toCurrency = strings.ToLower(toCurrency)

	log.Debugf("fetching %s%s rate from currency api", fromCurrency, toCurrency)
	resp, err := http.Get("https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/" + fromCurrency + "/" + toCurrency + ".json")
	if err != nil {
		return 0.0, fmt.Errorf("could not retreive currency rates for pair %s%s: %s", fromCurrency, toCurrency, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("could not parse currency rates response for pair %s%s: %s", fromCurrency, toCurrency, err)
	}

	// response has a dynamic name for the rate
	//   -> {"date": "2021-05-22", "usd": 1.218125}
	// making it predictable
	//   -> {"date": "2021-05-22", "rate": 1.218125}}
	bodyParsed := strings.Replace(string(body), toCurrency, "rate", 1)

	var response CurrencyResponse
	err = json.Unmarshal([]byte(bodyParsed), &response)
	if err != nil {
		return 0.0, err
	}

	return response.Rate, nil
}
