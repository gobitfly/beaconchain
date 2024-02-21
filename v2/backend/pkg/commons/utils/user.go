package utils

import (
	securerand "crypto/rand"
	"encoding/base64"
	"math/big"
)

// GenerateAPIKey generates an API key for a user
func GenerateRandomAPIKey() (string, error) {
	const apiLength = 28
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	max := big.NewInt(int64(len(letters)))
	key := make([]byte, apiLength)
	for i := 0; i < apiLength; i++ {
		num, err := securerand.Int(securerand.Reader, max)
		if err != nil {
			return "", err
		}
		key[i] = letters[num.Int64()]
	}

	apiKeyBase64 := base64.RawURLEncoding.EncodeToString(key)
	return apiKeyBase64, nil
}
