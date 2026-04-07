package validate

import "regexp"

var agreementPattern = regexp.MustCompile(`^AGR-\d+$`)

func AgreementNumber(s string) bool {
	return agreementPattern.MatchString(s)
}
