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

// compileRegex transforms a regex from string to regexp instance
func compileRegex(pattern string) (regex *regexp.Regexp, err error) {
	if pattern != "" {
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}
	return regex, nil
}
