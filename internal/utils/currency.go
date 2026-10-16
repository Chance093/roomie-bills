package utils

import (
	"math"
)

func SplitFourWay(amount float64) float64 {
	return math.Round(amount/4*100) / 100
}

// Assumes amount will never have more than 2 decimal places
func ConvertToCents(amount float64) int {
	// 12.34 => 1234
	res := int(amount * 100)
	return res
}

func ConvertToDollars(amount int) float64 {
	// 1234 => 12.34
	res := float64(amount) / 100
	return res
}
