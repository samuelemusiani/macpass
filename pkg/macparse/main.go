package macparse

import (
	"errors"
	"unicode"
)

func ParseMac(mac string, format string) (string, error) {
	if format == "linux" {
		var tmp []rune

		for _, value := range mac {
			if unicode.IsLetter(value) && unicode.IsUpper(value) {
				tmp = append(tmp, unicode.ToLower(value))
			} else if value == rune('-') {
				tmp = append(tmp, rune(':'))
			} else {
				tmp = append(tmp, value)
			}
		}

		return string(tmp), nil
	} else {
		return "", errors.New("Unsupported or unrecognized format")
	}
}
