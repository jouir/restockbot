package main

import (
	"net/url"
	"regexp"
	"strings"
)

// ExtractShopName parses a link to extract the hostname, then remove leading www, to build the Shop name
// "https://www.ldlc.com/informatique/pieces-informatique/[...]" -> "ldlc.com"
func ExtractShopName(link string) (name string, err error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`^www\.`)
	return strings.ToLower(re.ReplaceAllString(u.Hostname(), "")), nil
}

// ContainsString returns true when string is found in the array of strings
func ContainsString(arr []string, str string) bool {
	for _, elem := range arr {
		if elem == str {
			return true
		}
	}
	return false
}

// CoalesceInt64 returns the first non zero value from variadic int64 arguments
func CoalesceInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
