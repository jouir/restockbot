package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// Mapping between GPU model and NVIDIA SKU on the API
var nvidiaSKUs = map[string]string{
	"RTX 3060 Ti": "NVGFT060T",
	"RTX 3070":    "NVGFT070",
	"RTX 3070 Ti": "NVGFT070T",
	"RTX 3080":    "NVGFT080",
	"RTX 3080 Ti": "NVGFT080T",
	"RTX 3090":    "NVGFT090",
	"RTX 3090 Ti": "NVGFT090T",
}

// Mapping between location and currency
var nvidiaCurrencies = map[string]string{
	"es": "EUR",
	"fr": "EUR",
	"it": "EUR",
}

var supportedNvidiaGpus = []string{"RTX 3060 Ti", "RTX 3070", "RTX 3070 Ti", "RTX 3080", "RTX 3080 Ti", "RTX 3090", "RTX 3090 Ti"}
var supportedNvidiaFELocations = []string{"es", "fr", "it"}

type NvidiaFEParser struct {
	location  string
	gpus      []string
	userAgent string
	client    *http.Client
}

// NewNvidiaFRParser creates a parser for NVIDIA Founders Edition website for a specific location
// Takes a location (ex: fr) and GPU (ex: RTX 3070, RTX 3090)
func NewNvidiaFRParser(location string, gpus []string, userAgent string, timeout int) (*NvidiaFEParser, error) {
	// Check supported locales
	if !ContainsString(supportedNvidiaFELocations, location) {
		return nil, fmt.Errorf("location %s not supported (expect one of %s", location, supportedNvidiaFELocations)
	}

	// Check supported GPU list
	for _, gpu := range gpus {
		if !ContainsString(supportedNvidiaGpus, gpu) {
			return nil, fmt.Errorf("GPU %s not supported, expected one of %s", gpu, supportedNvidiaGpus)
		}
	}

	// Empty user agent will return an error on the NVIDIA API
	// It's probably a security measure to avoid third-party robots
	if userAgent == "" {
		return nil, fmt.Errorf("user agent required (use the same one as your web browser)")
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	return &NvidiaFEParser{
		location:  location,
		gpus:      gpus,
		userAgent: userAgent,
		client:    client,
	}, nil
}

// ShopName returns a nice name for NVIDIA Founders Edition website
// Implements the Parser interface
func (p *NvidiaFEParser) ShopName() (string, error) {
	return fmt.Sprintf("nvidia.com/%s-%s/shop", p.location, p.location), nil
}

// String to print NvidiaFEParser
// Implements the Parser interface
func (p *NvidiaFEParser) String() string {
	return fmt.Sprintf("NvidiaFEParser<%s>", p.location)
}

// NvidiaFEResponse to store NVIDIA API response
type NvidiaFEResponse struct {
	Success bool `json:"success"`
	ListMap []struct {
		IsActive string `json:"is_active"`
		Price    string `json:"price"`
	}
}

// Parse NVIDIA store API to return list of products
// Implements Parser interface
func (p *NvidiaFEParser) Parse() ([]*Product, error) {
	var products []*Product
	for _, gpu := range p.gpus {
		sku := nvidiaSKUs[gpu]
		apiURL := fmt.Sprintf("https://api.store.nvidia.com/partner/v1/feinventory?status=1&skus=%s&locale=%s-%s", sku, p.location, p.location)

		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new request: %s", err)
		}

		req.Header.Set("User-Agent", p.userAgent)
		req.Header.Set("Accept", "application/json")

		log.Debugf("requesting NVIDIA API: %s", req)

		res, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to request NVIDIA API: %s", err)
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %s", err)
		}

		if res.StatusCode != http.StatusOK {
			log.Debugf("%s", body)
			return nil, fmt.Errorf("NVIDIA API returned %d", res.StatusCode)
		}

		response := NvidiaFEResponse{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %s", err)
		}

		if !response.Success {
			return nil, fmt.Errorf("NVIDIA API returned an applicative failure for GPU %s and location %s", gpu, p.location)
		}

		for _, element := range response.ListMap {

			var product = &Product{
				Name:          gpu,
				PriceCurrency: nvidiaCurrencies[p.location],
			}

			available, err := strconv.ParseBool(element.IsActive)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool from response: %s", err)
			}
			product.Available = available

			productPrice, err := strconv.ParseFloat(element.Price, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float from response: %s", err)
			}
			product.Price = productPrice

			productURL, err := createNvidiaFEProductURL(p.location, gpu)
			if err != nil {
				return nil, fmt.Errorf("failed to create product URL: %s", err)
			}
			product.URL = productURL
			products = append(products, product)
		}

	}
	return products, nil
}

// Create the product URL
// Ex: https://store.nvidia.com/fr-fr/geforce/store/gpu/?page=1&limit=100&locale=fr-fr&category=GPU&gpu=RTX%203080&manufacturer=NVIDIA
func createNvidiaFEProductURL(location string, gpu string) (string, error) {
	locale := fmt.Sprintf("%s-%s", location, location)

	productURL, err := url.Parse("https://store.nvidia.com")
	if err != nil {
		return "", err
	}

	productURL.Path += fmt.Sprintf("%s/geforce/store/gpu/", locale)

	params := url.Values{}
	params.Add("page", "1")
	params.Add("limit", "100")
	params.Add("locale", locale)
	params.Add("category", "GPU")
	params.Add("gpu", gpu)
	params.Add("manufacturer", "NVIDIA")

	productURL.RawQuery = params.Encode()
	return productURL.String(), nil
}
