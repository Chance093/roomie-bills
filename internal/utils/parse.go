package utils

import (
	"errors"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ParseRoomieName(roomie string) (string, error) {
	lower := strings.ToLower(roomie)

	if lower != "chance" && lower != "kane" && lower != "alex" && lower != "madison" {
		return "", errors.New("Roomie does not exist. Please check spelling.")
	}

	return cases.Title(language.English).String(lower), nil
}
