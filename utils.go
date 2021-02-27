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
