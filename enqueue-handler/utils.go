package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func hash(string1, string2 string) string {
	hasher := sha256.New()
	hasher.Write([]byte(string1 + string2))
	hashedString := hex.EncodeToString(hasher.Sum(nil))
	return hashedString
}
