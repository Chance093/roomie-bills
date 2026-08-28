package utils

import "strings"

func GetHeaderCI(h map[string][]string, name string) string {
	lname := strings.ToLower(name)
	for k, v := range h {
		if strings.ToLower(k) == lname {
			return v[0]
		}
	}
	return ""
}
