package utils

import (
	securerand "crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"
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

// RandomString returns a random hex-string
func RandomString(length int) string {
	b, _ := GenerateRandomBytesSecure(length)
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func GenerateRandomBytesSecure(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := securerand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func CensorEmail(mail string) string {
	parts := strings.Split(mail, "@")
	username := parts[0]
	domain := parts[1]

	if len(username) > 2 {
		username = string(username[0]) + "***" + string(username[len(username)-1])
	}

	return username + "@" + domain
}
