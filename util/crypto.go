package util

import (
	"encoding/base64"
	"math/rand"
)

// GenKey returns a base64 encoded 256 bit key
func GenKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// GenURLRandStr returns a base64 encoded string
func GenURLRandStr(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(key), nil
}
