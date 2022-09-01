package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// NagiosOk return the Nagios OK code (see https://nagios-plugins.org/doc/guidelines.html#AEN78)
	NagiosOk = 0
	// NagiosWarning return the Nagios WARNING code (see https://nagios-plugins.org/doc/guidelines.html#AEN78)
	NagiosWarning = 1
	// NagiosCritical return the Nagios CRITICAL code (see https://nagios-plugins.org/doc/guidelines.html#AEN78)
	NagiosCritical = 2
	// NagiosUnknown return the Nagios UNKNOWN code (see https://nagios-plugins.org/doc/guidelines.html#AEN78)
	NagiosUnknown = 3
)

// MonitoringResult to store result of Nagios checks
type MonitoringResult struct {
	ShopName   string
	UpdatedAt  time.Time
	ReturnCode int
}

// String to print a MonitoringResult nicely
func (m MonitoringResult) String() string {
	diff := time.Now().Sub(m.UpdatedAt)

	var wording string
	if diff.Seconds() > 0 {
		wording = "seconds"
	} else {
		wording = "second"
	}

	return fmt.Sprintf("%s (%d %s ago)", m.ShopName, diff, wording)
}

// FormatMonitoringResults to print a list of MonitoringResult nicely
func FormatMonitoringResults(results []MonitoringResult) string {
	var s []string
	for _, result := range results {
		s = append(s, result.String())
	}
	return strings.Join(s, ", ")
}

// Monitor will check for last execution time for each shop and return either
// a warning or critical alert when the threshold has been reached
func Monitor(db *gorm.DB, warningTimeout int, criticalTimeout int) (rc int) {

	// Find date and time thresholds
	warningTime := time.Now().Add(-time.Duration(warningTimeout) * time.Second)
	criticalTime := time.Now().Add(-time.Duration(criticalTimeout) * time.Second)

	// Map to sort monitoring result by status code
	resultMap := make(map[int][]MonitoringResult)

	// List shops
	var shops []Shop
	trx := db.Find(&shops)
	if trx.Error != nil {
		fmt.Printf("%s\n", trx.Error)
		return NagiosUnknown
	}

	for _, shop := range shops {
		// Fetch last execution time
		var product Product
		trx := db.Where(Product{ShopID: shop.ID}).Order("updated_at asc").First(&product)
		if trx.Error == gorm.ErrRecordNotFound {
			fmt.Printf("%s\n", fmt.Errorf("No product found for shop %s", shop.Name))
			return NagiosCritical
		}
		if trx.Error != nil {
			fmt.Printf("%s\n", trx.Error)
			return NagiosUnknown
		}

		// Compare to thresholds and add to result map
		result := MonitoringResult{ShopName: shop.Name, UpdatedAt: product.UpdatedAt, ReturnCode: NagiosOk}
		if product.UpdatedAt.Before(criticalTime) {
			log.Infof("%s has been updated at %s (before time of %s) (crit)", shop.Name, product.UpdatedAt, criticalTime)
			result.ReturnCode = NagiosCritical
		} else if product.UpdatedAt.Before(warningTime) {
			log.Infof("%s has been updated at %s (before time of %s) (warn)", shop.Name, product.UpdatedAt, warningTime)
			result.ReturnCode = NagiosWarning
		} else {
			log.Infof("%s has been updated at %s (after %s) (ok)", shop.Name, product.UpdatedAt, warningTime)
		}
		resultMap[result.ReturnCode] = append(resultMap[result.ReturnCode], result)
	}

	var message, prefix string

	if len(resultMap[NagiosWarning]) > 0 {
		rc = NagiosWarning
		prefix = "WARN"
		message = FormatMonitoringResults(resultMap[NagiosWarning])
	} else if len(resultMap[NagiosCritical]) > 0 {
		rc = NagiosCritical
		prefix = "CRIT"
		message = FormatMonitoringResults(resultMap[NagiosCritical])
	} else {
		rc = NagiosOk
		prefix = "OK"
		message = "All shops have been updated recently"
	}

	// Print output
	fmt.Printf("%s - %s\n", prefix, message)
	return rc
}
