package hash

import (
	"crypto/md5"
	"encoding/hex"
)

func HashString(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}
