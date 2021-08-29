package util

import (
	"encoding/hex"
	"math/rand"
)

// RandomHexString creates a base16 random text with given length
func RandomHexString(length int) string {
	b := make([]byte, length/2+length%2)
	rand.Read(b)

	return hex.EncodeToString(b)[:length]
}
