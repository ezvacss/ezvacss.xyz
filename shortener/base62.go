package shortener

import (
	"math/rand"
	"strings"
	"time"
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
// as an alternative to encoding an auto-incrementing ID.
func GenerateRandomCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder
	
	for i := 0; i < length; i++ {
		idx := r.Intn(len(base62Chars))
		builder.WriteByte(base62Chars[idx])
	}
	
	return builder.String()
}
