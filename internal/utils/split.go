package utils

import "math"

func SplitFourWay(amount float64) float64 {
	return math.Round(amount/4*100) / 100
}
