package shortener

import (
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "A"},
		{35, "Z"},
		{36, "a"},
		{61, "z"},
		{62, "10"},
		{125, "21"},
		{3844, "100"},
	}

	for _, tt := range tests {
		got := Encode(tt.input)
		if got != tt.expected {
			t.Errorf("Encode(%d) = %s; want %s", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateRandomCode(t *testing.T) {
	lengths := []int{4, 6, 8, 12}

	for _, l := range lengths {
		code := GenerateRandomCode(l)
		if len(code) != l {
			t.Errorf("GenerateRandomCode(%d) returned string of length %d, want %d", l, len(code), l)
		}

		for _, char := range code {
			if !strings.ContainsRune(base62Chars, char) {
				t.Errorf("GenerateRandomCode(%d) returned invalid character '%c'", l, char)
			}
		}
	}

	// Verify randomness (two sequential codes should not be identical)
	c1 := GenerateRandomCode(8)
	c2 := GenerateRandomCode(8)
	if c1 == c2 {
		t.Errorf("GenerateRandomCode(8) generated duplicate codes consecutively: %s", c1)
	}
}
