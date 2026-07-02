package utils

import (
	"testing"
)

func TestParseRoomieName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "MADISON=>Madison", input: "MADISON", want: "Madison"},
		{name: "madison=>Madison", input: "madison", want: "Madison"},
		{name: "MaDiSoN=>Madison", input: "MaDiSoN", want: "Madison"},
		{name: "CHANCE=>Chance", input: "CHANCE", want: "Chance"},
		{name: "chance=>Chance", input: "chance", want: "Chance"},
		{name: "ChAnCe=>Chance", input: "ChAnCe", want: "Chance"},
		{name: "KANE=>Kane", input: "KANE", want: "Kane"},
		{name: "kane=>Kane", input: "kane", want: "Kane"},
		{name: "KaNe=>Kane", input: "KaNe", want: "Kane"},
		{name: "ALEX=>Alex", input: "ALEX", want: "Alex"},
		{name: "alex=>Alex", input: "alex", want: "Alex"},
		{name: "AlEx=>Alex", input: "AlEx", want: "Alex"},
		{name: "maddison fails", input: "maddison", want: ""},
		{name: "chancee fails", input: "chancee", want: ""},
		{name: "kaane fails", input: "kaane", want: ""},
		{name: "alexx fails", input: "alexx", want: ""},
		{name: "empty string fails", input: "", want: ""},
	}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := test.input
			want := test.want

			res, _ := ParseRoomieName(input)
			if res != want {
				t.Errorf("TestParseRoomieName%d failed: want: %s, got: %s", i, want, res)
			}
		})
	}
}
