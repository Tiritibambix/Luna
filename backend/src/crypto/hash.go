package crypto

import (
	"crypto/sha256"
)

func GetSha256Hash(data ...[]byte) []byte {
	hash := sha256.New()
	for _, d := range data {
		hash.Write(d)
	}
	digest := hash.Sum(nil)
	return digest
}

var DefaultArgon2Settings = map[string]int{
	"time":    1,
	"memory":  64 * 1024,
	"threads": 4,
	"keylen":  32,
}
