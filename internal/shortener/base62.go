package shortener

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Encode takes a positive integer and converts it to a Base62 string.
func Encode(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var builder strings.Builder
	base := uint64(len(base62Chars))

	// Collect remainders
	var remainders []uint64
	for num > 0 {
		remainders = append(remainders, num%base)
		num /= base
	}

	// Reverse and map to characters
	for i := len(remainders) - 1; i >= 0; i-- {
		builder.WriteByte(base62Chars[remainders[i]])
	}

	return builder.String()
}

// GenerateRandomCode generates a random base-62 code of a specific length
// using cryptographically secure pseudo-random number generator (crypto/rand).
func GenerateRandomCode(length int) string {
	var builder strings.Builder
	base := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, base)
		if err != nil {
			builder.WriteByte(base62Chars[0])
			continue
		}
		builder.WriteByte(base62Chars[n.Int64()])
	}

	return builder.String()
}
