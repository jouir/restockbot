package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	paapi5 "github.com/spiegel-im-spiegel/pa-api"
	"github.com/spiegel-im-spiegel/pa-api/entity"
	"github.com/spiegel-im-spiegel/pa-api/query"
)

// NewAmazonServer creates an Amazon Server function based on the Marketplace.
// The paapi5 marketplaceEnum is not exported, so this type cannot be used in simple map.
func NewAmazonServer(marketplace string) *paapi5.Server {
	switch marketplace {
	case "www.amazon.fr":
		return paapi5.New(paapi5.WithMarketplace(paapi5.LocaleFrance))
	case "www.amazon.com":
		return paapi5.New(paapi5.WithMarketplace(paapi5.LocaleUnitedStates))
	default:
		return paapi5.New() // default Marketplace
	}
}

// Map of messages to detect availability
var availabilityMessages = []string{"En stock."}

// AmazonParser structure to handle Amazon parsing logic
type AmazonParser struct {
	client          paapi5.Client
	searches        []string
	includeRegex    *regexp.Regexp
	excludeRegex    *regexp.Regexp
	amazonFulfilled bool
	amazonMerchant  bool
	affiliateLinks  bool
}

// NewAmazonParser to create a new AmazonParser instance
func NewAmazonParser(marketplace string, partnerTag string, accessKey string, secretKey string, searches []string, includeRegex string, excludeRegex string, amazonFulfilled bool, amazonMerchant bool, affiliateLinks bool) (*AmazonParser, error) {
	var err error
	var includeRegexCompiled, excludeRegexCompiled *regexp.Regexp

	log.Debugf("compiling include name regex")
	if includeRegex != "" {
		includeRegexCompiled, err = regexp.Compile(includeRegex)
		if err != nil {
			return nil, err
		}
	}

	log.Debugf("compiling exclude name regex")
	if excludeRegex != "" {
		excludeRegexCompiled, err = regexp.Compile(excludeRegex)
		if err != nil {
			return nil, err
		}
	}

	return &AmazonParser{
		client:          NewAmazonServer(marketplace).CreateClient(partnerTag, accessKey, secretKey),
		searches:        searches,
		includeRegex:    includeRegexCompiled,
		excludeRegex:    excludeRegexCompiled,
		amazonFulfilled: amazonFulfilled,
		amazonMerchant:  amazonMerchant,
		affiliateLinks:  affiliateLinks,
	}, nil
}

// Parse Amazon API to return list of products
// Implements Parser interface
func (p *AmazonParser) Parse() ([]*Product, error) {

	var products []*Product

	for _, search := range p.searches {

		log.Debugf("searching for '%s' on %s", search, p.client.Marketplace())

		// create search request on API
		q := query.NewSearchItems(
			p.client.Marketplace(),
			p.client.PartnerTag(),
			p.client.PartnerType(),
		).Search(query.Keywords, search).EnableItemInfo().EnableOffers()
		body, err := p.client.Request(q)
		if err != nil {
			return nil, err
		}

		// decode response
		res, err := entity.DecodeResponse(body)
		if err != nil {
			return nil, err
		}

		// decode products
		for _, item := range res.SearchResult.Items {

			product := &Product{}
			if !p.affiliateLinks {
				product.URL = fmt.Sprintf("https://%s/dp/%s", p.client.Marketplace(), item.ASIN)

			} else {
				product.URL = item.DetailPageURL // includes partner tag
			}
			product.Name = item.ItemInfo.Title.DisplayValue

			if item.Offers != nil && *item.Offers.Listings != nil {
				for _, offer := range *item.Offers.Listings {
					// detect if product is packaged by Amazon
					if p.amazonFulfilled && !offer.DeliveryInfo.IsAmazonFulfilled {
						log.Debugf("excluding offer by '%s' for product '%s' because not fulfilled by Amazon", offer.MerchantInfo.Name, product.Name)
						continue
					}

					// detect if product is sold by Amazon
					if p.amazonMerchant && !strings.HasPrefix(offer.MerchantInfo.Name, "Amazon") {
						log.Debugf("excluding offer by '%s' for product '%s' because not sold by Amazon", offer.MerchantInfo.Name, product.Name)
						continue
					}

					// detect price
					product.Price = offer.Price.Amount
					product.PriceCurrency = offer.Price.Currency

					// detect availability
					if ContainsString(availabilityMessages, offer.Availability.Message) {
						product.Available = true
						break
					}
				}
			}

			products = append(products, product)
		}
	}

	// apply filters
	products = filterInclusive(p.includeRegex, products)
	products = filterExclusive(p.excludeRegex, products)

	return products, nil
}

// String to print AmazonParser
// Implements the Parser interface
func (p *AmazonParser) String() string {
	return fmt.Sprintf("AmazonParser<%s@%s>", p.client.PartnerTag(), p.client.Marketplace())
}

// ShopName returns shop name from Amazon Marketplace
func (p *AmazonParser) ShopName() string {
	return strings.ReplaceAll(p.client.Marketplace(), "www.", "")
}
