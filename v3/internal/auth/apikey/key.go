package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/deatil/go-encoding/base62"
	"github.com/google/uuid"
)

/* APIKey holds the hashed key and metadata for an API key. */

var ErrMissingCredential = errors.New("api key credential is nil")

func NewAPIKeyFromCredentials(name, shortKey string, value HashedKeyCredential) (APIKey, error) {
	if value.IsEmpty() {
		return APIKey{}, ErrMissingCredential
	}

	return APIKey{
		ID:         nil,
		Value:      value[:],
		Name:       name,
		ShortKey:   shortKey,
		CreatedAt:  nil,
		DisabledAt: nil,
		LastUsedAt: nil,
	}, nil
}

func NewAPIKey(name string) (APIKey, RawKeyCredential, error) {
	rawKey, err := NewRawAPIKey()
	if err != nil {
		return APIKey{}, RawKeyCredential{}, err
	}
	shortKey := rawKey.Obfuscate()
	key, err := NewAPIKeyFromCredentials(name, shortKey, rawKey.GetAPIKeyCredential())
	return key, rawKey, err
}

type APIKey struct {
	ID         *uuid.UUID `db:"api_key_id"`
	Value      []byte     `db:"api_key"`
	Name       string     `db:"name"`
	ShortKey   string     `db:"short_key"`
	CreatedAt  *time.Time `db:"created_at"`
	DisabledAt *time.Time `db:"disabled_at"`
	LastUsedAt *time.Time `db:"last_used_at"`
}

// == Raw Key ==
/*
Raw key is the primitive of an unhashed API key as they passed to the user on creation and
provided on every request. Be mindful of the security implications of dealing with this type.
*/

const RawKeyCredentialLength = 32

type RawKeyCredential [RawKeyCredentialLength]byte

func NewRawAPIKey() (RawKeyCredential, error) {
	b := make([]byte, RawKeyCredentialLength)
	_, err := rand.Read(b)
	if err != nil {
		return RawKeyCredential{}, err
	}
	return RawKeyCredential(b), nil
}

func FromBase62(encoded string) (RawKeyCredential, error) {
	decoded, err := base62.StdEncoding.DecodeString(encoded)
	if err != nil {
		return RawKeyCredential{}, err
	}
	return RawKeyCredential(decoded), nil
}

// ToBase62 returns the raw API key in Base62 encoding.
// Be mindful of the security implications of exposing raw API keys.
func (r RawKeyCredential) ToBase62() string {
	return base62.StdEncoding.EncodeToString(r.Bytes())
}

func (r RawKeyCredential) Obfuscate() string {
	if len(r) < 6 {
		return "[invalid]"
	}
	base62Encoded := r.ToBase62()
	return fmt.Sprintf("%s...%s", base62Encoded[:3], base62Encoded[len(base62Encoded)-3:])
}

func (r RawKeyCredential) Hash() [HashLength]byte {
	return HashSHA256(r.Bytes())
}

func (r RawKeyCredential) GetAPIKeyCredential() HashedKeyCredential {
	return NewHashedKeyCredential(r.Hash())
}

// Bytes returns the raw API key bytes.
// Be mindful of the security implications of exposing raw API keys.
func (r RawKeyCredential) Bytes() []byte {
	return r[:]
}

// == Hashed Key ==
/*
Hashed key is the primitive of a hashed version of the raw API key, used for storage and comparison.
*/

const HashedKeyCredentialLength = HashLength

type HashedKeyCredential [HashedKeyCredentialLength]byte

func NewHashedKeyCredential(raw [HashedKeyCredentialLength]byte) HashedKeyCredential {
	return HashedKeyCredential(raw)
}

func (k HashedKeyCredential) IsEmpty() bool {
	return k.Equal(HashedKeyCredential{})
}

func (k HashedKeyCredential) Equal(other HashedKeyCredential) bool {
	return subtle.ConstantTimeCompare(k.Bytes(), other.Bytes()) == 1
}

func (k HashedKeyCredential) Bytes() []byte {
	return k[:]
}

const (
	HashLength = 32 // SHA-256 produces a 32-byte hash
)

func HashSHA256(data []byte) [32]byte {
	h := sha256.Sum256(data)
	return h
}
