package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// NAGIOS_OK to return the Nagios OK code (see https://nagios-plugins.org/doc/guidelines.html#AEN78)
	NAGIOS_OK = 0
	// NAGIOS_WARNING to return the Nagios WARNING code
	NAGIOS_WARNING = 1
	// NAGIOS_CRITICAL to return the Nagios CRITICAL code
	NAGIOS_CRITICAL = 2
	// NAGIOS_UNKNOWN to return the Nagios UNKNOWN code
	NAGIOS_UNKNOWN = 3
)

// Monitor will check for last execution time for each shop and return either
// a warning or critical alert when the threshold has been reached
func Monitor(db *gorm.DB, warningTimeout int, criticalTimeout int) (rc int) {

	// Find date and time thresholds
	warningTime := time.Now().Add(-1 * time.Duration(warningTimeout) * time.Second)
	criticalTime := time.Now().Add(-1 * time.Duration(criticalTimeout) * time.Second)

	// shops needing attention
	// key is error level (warning, critical) and value is the shop name
	resultMap := make(map[string][]string)

	// List shops
	var shops []Shop
	trx := db.Find(&shops)
	if trx.Error != nil {
		fmt.Printf("%s\n", trx.Error)
		return NAGIOS_UNKNOWN
	}

	for _, shop := range shops {
		// Fetch last execution time
		var product Product
		trx := db.Where(Product{ShopID: shop.ID}).Order("updated_at asc").First(&product)
		if trx.Error == gorm.ErrRecordNotFound {
			fmt.Printf("%s\n", fmt.Errorf("No product found for shop %s", shop.Name))
			return NAGIOS_CRITICAL
		}
		if trx.Error != nil {
			fmt.Printf("%s\n", trx.Error)
			return NAGIOS_UNKNOWN
		}

		// Compare to thresholds and add to result map
		if product.UpdatedAt.Before(criticalTime) {
			log.Infof("%s has been updated at %s (before time of %s)", shop.Name, product.UpdatedAt, criticalTime)
			resultMap["critical"] = append(resultMap["critical"], shop.Name)
		} else if product.UpdatedAt.Before(warningTime) {
			log.Infof("%s has been updated at %s (before time of %s)", shop.Name, product.UpdatedAt, warningTime)
			resultMap["warning"] = append(resultMap["warning"], shop.Name)
		} else {
			log.Infof("%s has been updated at %s (after %s)", shop.Name, product.UpdatedAt, warningTime)
		}
	}

	var message, prefix string

	if len(resultMap["warning"]) > 0 {
		rc = NAGIOS_WARNING
		prefix = "WARN"
		message = strings.Join(resultMap["warning"], ", ")
	} else if len(resultMap["critical"]) > 0 {
		rc = NAGIOS_CRITICAL
		prefix = "CRIT"
		message = strings.Join(resultMap["critical"], ", ")
	} else {
		rc = NAGIOS_OK
		prefix = "OK"
		message = "All shops have been updated recently"
	}

	// Print output
	fmt.Printf("%s - %s\n", prefix, message)
	return rc
}
