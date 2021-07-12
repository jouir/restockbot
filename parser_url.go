package main

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/MontFerret/ferret/pkg/compiler"
	"github.com/MontFerret/ferret/pkg/drivers"
	"github.com/MontFerret/ferret/pkg/drivers/cdp"
	"github.com/MontFerret/ferret/pkg/drivers/http"
)

// URLParser structure to handle websites parsing logic
type URLParser struct {
	url string
	ctx context.Context
}

// String to print URLParser
// Implements the Parser interface
func (p *URLParser) String() string {
	return fmt.Sprintf("URLParser<%s>", p.url)
}

// ShopName returns shop name from URL
func (p *URLParser) ShopName() (string, error) {
	return ExtractShopName(p.url)
}

// NewURLParser to create a new URLParser instance
func NewURLParser(url string, browserAddress string) *URLParser {

	log.Debugf("creating context with headless browser drivers")
	ctx := context.Background()
	ctx = drivers.WithContext(ctx, cdp.NewDriver(cdp.WithAddress(browserAddress)))
	ctx = drivers.WithContext(ctx, http.NewDriver(), drivers.AsDefault())

	return &URLParser{
		url: url,
		ctx: ctx,
	}
}

// Parse a website to return list of products
// Implements Parser interface
// TODO: redirect output to logger
func (p *URLParser) Parse() ([]*Product, error) {
	shopName, err := ExtractShopName(p.url)
	if err != nil {
		return nil, err
	}

	query, err := createQuery(shopName, p.url)
	if err != nil {
		return nil, err
	}
	comp := compiler.New()
	program, err := comp.Compile(string(query))
	if err != nil {
		return nil, err
	}

	out, err := program.Run(p.ctx)
	if err != nil {
		return nil, err
	}
	var products []*Product
	err = json.Unmarshal(out, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func createQuery(shopName string, url string) (string, error) {
	switch shopName {
	case "cybertek.fr":
		return createQueryForCybertek(url), nil
	case "ldlc.com":
		return createQueryForLDLC(url), nil
	case "materiel.net":
		return createQueryForMaterielNet(url), nil
	case "microcenter.com":
		return createQueryForMicroCenter(url), nil
	case "mediamarkt.ch":
		return createQueryForMediamarktCh(url), nil
	case "newegg.com":
		return createQueryForNewegg(url), nil
	case "steg-electronics.ch":
		return createQueryForStegElectronics(url), nil
	case "topachat.com":
		return createQueryForTopachat(url), nil
	case "vsgamers.es":
		return createQueryForVersusGamers(url), nil
	default:
		return "", fmt.Errorf("shop %s not supported", shopName)
	}
}

func createQueryForLDLC(url string) string {
	q := `
// gather first page
LET first_page  = '` + url + `'
LET doc = DOCUMENT(first_page, {driver: "cdp"})

// discover next pages
LET pagination = ELEMENT(doc, ".pagination")
LET next_pages = (
	FOR url in ELEMENTS(pagination, "a")
		RETURN "https://www.ldlc.com" + url.attributes.href
)

// append first page to pagination and remove duplicates
LET pages = SORTED_UNIQUE(APPEND(next_pages, first_page))

// create a result array containing an one array of products per page
LET results = (
	FOR page IN pages
		NAVIGATE(doc, page)
		LET products = (
			FOR el IN ELEMENTS(doc, ".pdt-item")
				LET url = ELEMENT(el, "a")
				LET name = INNER_TEXT(ELEMENT(el, "h3"))
				LET price = TO_FLOAT(SUBSTITUTE(SUBSTITUTE(INNER_TEXT(ELEMENT(el, ".price")), "€", "."), " ", ""))
				LET available = !CONTAINS(INNER_TEXT(ELEMENT(el, ".stock-web"), 'span'), "RUPTURE")
				RETURN {
					name: name,
					url: "https://www.ldlc.com" + url.attributes.href,
					price: price,
					price_currency: "EUR",
					available: available,
				}
		)
		RETURN products
)

// combine all arrays to a single one
RETURN FLATTEN(results)
	`
	return q
}

func createQueryForMaterielNet(url string) string {
	q := `
// gather first page
LET first_page  = '` + url + `'
LET doc = DOCUMENT(first_page, {driver: "cdp"})

// discover next pages
LET pagination = ELEMENT(doc, ".pagination")
LET next_pages = (
    FOR url in ELEMENTS(pagination, "a")
        RETURN "https://www.materiel.net" + url.attributes.href
)

// append first page to pagination and remove duplicates
LET pages = SORTED_UNIQUE(APPEND(next_pages, first_page))

// create a result array containing an one array of products per page
LET results = (
    FOR page IN pages
        NAVIGATE(doc, page)
        WAIT_ELEMENT(doc, "div .o-product__price")
        LET products = (
            FOR el IN ELEMENTS(doc, "div .ajax-product-item")
                LET image = ELEMENT(el, "img")
                LET url = ELEMENT(el, "a")
                LET price = TO_FLOAT(SUBSTITUTE(SUBSTITUTE(INNER_TEXT(ELEMENT(el, "div .o-product__price")), "€", "."), " ", ""))
                LET available = !CONTAINS(ELEMENT(el, "div .o-availability__value"), "Rupture")
                RETURN {
                    name: image.attributes.alt,
                    url: "https://www.materiel.net" + url.attributes.href,
                    price: price,
                    price_currency: "EUR",
                    available: available,
                }
        )
        RETURN products
)

// combine all arrays to a single one
RETURN FLATTEN(results)
	`
	return q
}

func createQueryForTopachat(url string) string {
	q := `
LET page = '` + url + `'
LET doc = DOCUMENT(page, {driver: "cdp"})

FOR el IN ELEMENTS(doc, "article .grille-produit")
    LET url = ELEMENT(el, "a")
    LET name = INNER_TEXT(ELEMENT(el, "h3"))
    LET price = TO_FLOAT(ELEMENT(el, "div .prod_px_euro").attributes.content)
    LET available = !CONTAINS(ELEMENT(el, "link").attributes.href, "http://schema.org/OutOfStock")
    RETURN {
        url: "https://www.topachat.com" + url.attributes.href,
        name: name,
        price: price,
        price_currency: "EUR",
        available: available,
    }
	`
	return q
}

func createQueryForCybertek(url string) string {
	q := `
// gather first page
LET first_page  = '` + url + `'
LET doc = DOCUMENT(first_page, {driver: "cdp"})
LET home_page = 'https://www.cybertek.fr/boutique/index.aspx'

// discover next pages
LET pagination = ELEMENT(doc, "div .pagination-div")
LET next_pages = (
    FOR url in ELEMENTS(pagination, "a")
        RETURN url.attributes.href
)

// append first page to pagination, remove "#" link and remove duplicates
LET pages = SORTED_UNIQUE(APPEND(MINUS(next_pages, ["#"]), first_page))

// create a result array containing an one array of products per page
LET results = (
    FOR page in pages
        NAVIGATE(doc, page)
        LET products_available = (
            FOR el IN ELEMENTS(doc, "div .listing_dispo")
                LET url = ELEMENT(el, "a")
                FILTER url.attributes.href != home_page // exclude home page
                LET name = TRIM(FIRST(SPLIT(INNER_TEXT(ELEMENT(el, "div .height-txt-cat")), "-")))
                LET price = TO_FLOAT(SUBSTITUTE(INNER_TEXT(ELEMENT(el, "div .price_prod_resp")), "€", "."))
                RETURN {
                    name: name,
                    url: url.attributes.href,
                    available: true,
                    price: price,
                    price_currency: "EUR",
                }
        )
        LET products_not_available = (
            FOR el IN ELEMENTS(doc, "div .listing_nodispo")
                LET url = ELEMENT(el, "a")
                FILTER url.attributes.href != home_page // exclude home page
                LET name = TRIM(FIRST(SPLIT(INNER_TEXT(ELEMENT(el, "div .height-txt-cat")), "-")))
                LET price = TO_FLOAT(SUBSTITUTE(INNER_TEXT(ELEMENT(el, "div .price_prod_resp")), "€", "."))
                RETURN {
                    name: name,
                    url: url.attributes.href,
                    available: false,
                    price: price,
                    price_currency: "EUR",
                }
        )
        // combine available and not available list of products into a single array of products
        RETURN FLATTEN([products_available, products_not_available])
)

// combine all arrays to a single one
RETURN FLATTEN(results)
    `
	return q
}

func createQueryForMediamarktCh(url string) string {
	q := `
LET page = '` + url + `'
LET doc = DOCUMENT(page, {driver: "cdp"})

LET pagination = (
    FOR el IN ELEMENTS(doc, "div .pagination-wrapper a")
        RETURN "https://www.mediamarkt.ch" + el.attributes.href
)

LET pages = SORTED_UNIQUE(pagination)

LET results = (
    FOR page IN pages
        NAVIGATE(doc, page)
        LET products = (
            FOR el IN ELEMENTS(doc, "div .product-wrapper")
                LET name = TRIM(FIRST(SPLIT(INNER_TEXT(ELEMENT(el, "h2")), "-")))
                LET url = ELEMENT(el, "a").attributes.href
                LET price = TO_FLOAT(CONCAT(POP(ELEMENTS(el, "div .price span"))))
                LET available = !REGEX_TEST(INNER_TEXT(ELEMENT(el, "div .availability li")), "^Non disponible(.*)")
                RETURN {
                    name: name,
                    url: "https://www.mediamarkt.ch" + url,
                    price: price,
                    price_currency: "CHF",
                    available: available,
                }
        )
        RETURN products
)

RETURN FLATTEN(results)
	`
	return q
}

func createQueryForMicroCenter(url string) string {
	q := `
LET first_page = '` + url + `'
LET doc = DOCUMENT(first_page, {driver: "cdp"})
LET base_url = 'https://www.microcenter.com'

// determine store id or default to "shippable items" (online store)
LET detected_store_id = LAST(REGEX_MATCH(first_page, '&storeid=([0-9]*)', true))
LET store_id = detected_store_id ? detected_store_id : "029" // id of the online store if not found
COOKIE_SET(doc, {"name": "storeSelected", "value": store_id})

LET next_pages = (
    FOR a IN ELEMENTS(doc, "div .pagination .pages a")
        RETURN base_url + a.attributes.href
)

LET pages = SORTED_UNIQUE(APPEND(next_pages, first_page))

LET results = (
    FOR page IN pages
        NAVIGATE(doc, page)
        LET products = (
            FOR el IN ELEMENTS(doc, "div .products .product_wrapper")
                LET details = ELEMENT(el, "h2")
                LET name = INNER_TEXT(details)
                LET url = ELEMENT(details, "a").attributes.href
                LET price = TO_FLOAT(ELEMENT(details, "a").attributes."data-price")
                LET available = LENGTH(ELEMENTS(el, "div .price_wrapper form .STBTN"))>0
                RETURN {
                    name: name,
                    url: base_url + url + "?storeid=" + store_id,
                    price: price,
                    price_currency: "USD",
                    available: available,
                }
        )
        RETURN products
    )

RETURN FLATTEN(results)
	`
	return q
}

func createQueryForNewegg(url string) string {
	q := `
LET first_page = '` + url + `'
LET doc = DOCUMENT(first_page, {driver: "cdp"})

LET pagination = LAST(ELEMENTS(doc, 'div .list-tool-pagination'))
LET pages = (
    FOR button IN ELEMENTS(pagination, 'button')
        RETURN TO_INT(INNER_TEXT(button))
)

LET pages_count = MAX(pages)

LET results = (
    FOR page IN 1..pages_count
        LET current_page = first_page + '&page=' + page

        NAVIGATE(doc, current_page)

        LET products = (
            FOR el IN ELEMENTS(doc, "div .item-cell")
                LET a = ELEMENT(el, "div .item-title")
                LET price = TO_FLOAT(SUBSTITUTE(SUBSTITUTE(TRIM(INNER_TEXT(ELEMENT(el, "div .price-current"))), '$', ''), ',', ''))
                LET available = !ELEMENT_EXISTS(el, "div .item-promo")
                RETURN {
                    name: INNER_TEXT(a),
                    url: a.attributes.href,
                    price: price,
                    price_currency: "USD",
                    available: available,
                }
        )
        RETURN products
)

RETURN FLATTEN(results)
	`
	return q
}

func createQueryForStegElectronics(url string) string {
	q := `
LET first_page = '` + url + `'
LET doc = DOCUMENT(first_page, { driver: "cdp" })

LET next_pages = (
    FOR next_page IN ELEMENTS(doc, "div#pagination a.current-pages")
        RETURN next_page.attributes.href
)

LET pages = SORTED_UNIQUE(APPEND(REMOVE_VALUE(next_pages, '#'), first_page))

LET results = (
    FOR page IN pages
        NAVIGATE(doc, page)
        LET products = (
            FOR el IN ELEMENTS(doc, "article.product-list-container div.productGridElement")
                LET link = ELEMENT(el, "h2 a")
                LET shipment = ELEMENT(el, "div.availabilityInfo a .iconShipment").attributes.style.color != "#BFBFBF"
                LET pickup = ELEMENT(el, "div.availabilityInfo a .iconPickup").attributes.style.color != "#BFBFBF"
                LET price = TO_FLOAT(SUBSTITUTE(INNER_TEXT(ELEMENT(el, "div.generalPrice")), "'", ""))
                RETURN {
                    name: link.attributes.title,
                    url: link.attributes.href,
                    available: shipment OR pickup,
                    price: price,
                    price_currency: "CHF",
                }
        )
        RETURN products
)

RETURN FLATTEN(results)
	`
	return q
}

/*
* TODO:
*  - pagination (relies on scroll down move to feed the "div.vs-product-list" element with new items)
*  - remove products with "name: null"
 */
func createQueryForVersusGamers(url string) string {
	q := `
LET first_page = '` + url + `'
LET doc = DOCUMENT(first_page, { driver: "cdp" })
LET base_url = 'https://www.vsgamers.es'

FOR el IN ELEMENTS(doc, "div.vs-product-list div.vs-product-list-item")
    LET link = ELEMENT(el, "a")
    LET hasLink = link != null ? true : false
    LET title = ELEMENT(el, "div.vs-product-card-title")
    LET name = title != null ? INNER_TEXT(title) : null
    LET url = hasLink ? base_url + link.attributes.href : null

    LET available = ELEMENT(el, "div.vs-product-card-action button") != null

    LET price_element = ELEMENT(el, "div.vs-product-card-prices span")
    LET price = price_element != null ? TO_FLOAT(price_element.attributes."data-price") : 0
    LET price_currency = "EUR"
    RETURN {
        name: name,
        url: url,
        available: available,
        price: price,
        price_currency: price_currency,
    }
	`
	return q
}
