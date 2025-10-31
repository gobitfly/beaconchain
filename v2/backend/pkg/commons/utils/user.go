package utils

import (
	securerand "crypto/rand"
	"encoding/base64"
	"math/big"
	"slices"
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

// As a safety precaution we don't want to expose the full email address via the API
// We can rest assured that even if a user session ever leaks, no personal data is provided via api that could link the users addresses or validators to them
func CensorEmail(mail string) string {
	parts := strings.Split(mail, "@")
	if len(parts) != 2 { // invalid mail, should not happen
		return mail
	}
	username := parts[0]
	domain := parts[1]

	if len(username) > 2 {
		username = string(username[0]) + "***" + string(username[len(username)-1])
	}

	// Also censor domain part for not well known domains as they could be used to identify the user if it's a niche domain
	domainParts := strings.Split(domain, ".")
	if len(parts) == 2 {
		// https://email-verify.my-addr.com/list-of-most-popular-email-domains.php
		wellKnownDomains := []string{"gmail", "hotmail", "yahoo", "apple", "aol", "outlook", "gmx", "live", "comcast", "msn"}

		if !slices.Contains(wellKnownDomains, domainParts[0]) && len(domainParts[0]) > 2 {
			domain = string(domainParts[0][0]) + "***" + string(domainParts[0][len(domainParts[0])-1]) + "." + domainParts[1]
		}
	}

	return username + "@" + domain
}
