package invite

import (
	"crypto/rand"
	"math/big"
)

// Characters excluding ambiguous ones (0/O, 1/I/L)
const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func GenerateCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}
