package model

import "fmt"

// FormatPence formats pence as £X,XXX.XX with comma grouping.
func FormatPence(pence int64) string {
	negative := pence < 0
	if negative {
		pence = -pence
	}
	whole := pence / 100
	frac := pence % 100

	s := fmt.Sprintf("%d", whole)
	n := len(s)
	if n > 3 {
		var result []byte
		for i, c := range s {
			if i > 0 && (n-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, byte(c))
		}
		s = string(result)
	}

	prefix := "£"
	if negative {
		prefix = "-£"
	}
	return fmt.Sprintf("%s%s.%02d", prefix, s, frac)
}

// DayWithSuffix returns day with ordinal suffix (1st, 2nd, 3rd, 4th, etc.)
func DayWithSuffix(day int) string {
	return fmt.Sprintf("%d%s", day, DaySuffix(day))
}

func DaySuffix(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
