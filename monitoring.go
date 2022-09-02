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
	Message    string
	ReturnCode int
}

// String returns a string to print a MonitoringResult nicely
func (m MonitoringResult) String() string {
	return fmt.Sprintf("%s %s (rc = %d)", m.ShopName, m.Message, m.ReturnCode)
}

// ReturnCodeString returns a string to print a ReturnCode nicely
func ReturnCodeString(rc int) string {
	switch rc {
	case NagiosOk:
		return "OK"
	case NagiosWarning:
		return "WARN"
	case NagiosCritical:
		return "CRIT"
	default:
		return "UNK"
	}
}

// FormatMonitoringResults to print a list of MonitoringResult nicely
func FormatMonitoringResults(results []MonitoringResult) string {
	var s []string
	for _, result := range results {
		s = append(s, fmt.Sprintf("%s %s", result.ShopName, result.Message))
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

		result := MonitoringResult{
			ShopName:   shop.Name,
			ReturnCode: NagiosOk,
		}

		// Fetch last execution time
		var product Product
		trx := db.Where(Product{ShopID: shop.ID}).Order("updated_at desc").First(&product)
		if trx.Error == gorm.ErrRecordNotFound {
			result.Message = "has not been updated"
			result.ReturnCode = NagiosCritical
			resultMap[NagiosCritical] = append(resultMap[result.ReturnCode], result)
			continue
		}
		if trx.Error != nil {
			fmt.Printf("%s\n", trx.Error)
			return NagiosUnknown
		}

		// Compare to thresholds and add to result map
		diff := int(time.Now().Sub(product.UpdatedAt.Local()).Seconds())
		result.Message = fmt.Sprintf("updated %d seconds ago", diff)

		if product.UpdatedAt.Before(criticalTime) {
			result.ReturnCode = NagiosCritical
		} else if product.UpdatedAt.Before(warningTime) {
			result.ReturnCode = NagiosWarning
		} else {
		}
		log.Info(result)
		resultMap[result.ReturnCode] = append(resultMap[result.ReturnCode], result)
	}

	var message string

	if len(resultMap[NagiosWarning]) > 0 {
		rc = NagiosWarning
		message = FormatMonitoringResults(resultMap[NagiosWarning])
	} else if len(resultMap[NagiosCritical]) > 0 {
		rc = NagiosCritical
		message = FormatMonitoringResults(resultMap[NagiosCritical])
	} else {
		rc = NagiosOk
		message = "All shops have been updated recently"
	}

	// Print output
	fmt.Printf("%s - %s\n", ReturnCodeString(rc), message)
	return
}
