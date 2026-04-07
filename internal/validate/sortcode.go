package validate

import "regexp"

var sortCodePattern = regexp.MustCompile(`^\d{2}-\d{2}-\d{2}$`)

func SortCode(s string) bool {
	return sortCodePattern.MatchString(s)
}
