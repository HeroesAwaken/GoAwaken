package GameSpy

import (
	"crypto/md5"
	"encoding/hex"
)

// ShortHash returns a MD5 hash of "str" reduced to 12 chars.
func ShortHash(str string) string {

	// Generate md5 hash
	hash := md5.New()
	sum := hash.Sum([]byte(str))

	// Convert to Hex and save first 12 characters
	hexSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(hexSum, sum)
	shortendHexSum := string(hexSum[0:12])

	return shortendHexSum
}
