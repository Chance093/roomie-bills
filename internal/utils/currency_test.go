package utils

import (
	"testing"
)

func TestConvertToCents(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  int
	}{
		{name: "12.34 => 1234", input: 12.34, want: 1234},
		{name: "0.34 => 34", input: 0.34, want: 34},
		{name: "0.00 => 0", input: 0.00, want: 0},
		{name: "100.00 => 10000", input: 100.00, want: 10000},
		{name: "12.345 => 1234", input: 12.345, want: 1234},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := test.input
			want := test.want

			res := ConvertToCents(input)
			if res != want {
				t.Errorf("TestConvertToCents%d failed: want: %d, got: %d", i, want, res)
			}
		})
	}
}

func TestConvertToDollars(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  float64
	}{
		{name: "1234 => 12.34", input: 1234, want: 12.34},
		{name: "34 => 0.34", input: 34, want: 0.34},
		{name: "0 => 0.00", input: 0, want: 0.00},
		{name: "10000 => 100.00", input: 10000, want: 100.00},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := test.input
			want := test.want

			res := ConvertToDollars(input)
			if res != want {
				t.Errorf("TestConvertToDollars%d failed: want: %v, got: %v", i, want, res)
			}
		})
	}
}
