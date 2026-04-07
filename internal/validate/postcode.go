package validate

import (
	"regexp"
	"strings"
)

var postcodePattern = regexp.MustCompile(`(?i)^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`)

func UKPostcode(s string) bool {
	return postcodePattern.MatchString(strings.TrimSpace(s))
}

func NormalizePostcode(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	// Remove all spaces, then insert single space before last 3 chars
	s = strings.ReplaceAll(s, " ", "")
	if len(s) > 3 {
		return s[:len(s)-3] + " " + s[len(s)-3:]
	}
	return s
}

func NormalizePostcodeForComparison(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, " ", ""))
}
