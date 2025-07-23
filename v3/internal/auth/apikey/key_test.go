package apikey

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIKeyFromCredentials_MissingCredential(t *testing.T) {
	name := "test-key"
	shortKey := "short"
	cred := HashedKeyCredential{} // Empty credential

	_, err := NewAPIKeyFromCredentials(name, shortKey, cred)
	if !errors.Is(err, ErrMissingCredential) {
		t.Errorf("expected ErrMissingCredential, got %v", err)
	}
}

// == Raw Key

func TestNewRawAPIKey_UniqueAndLength(t *testing.T) {
	key1, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	key2, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	if key1 == key2 {
		t.Error("NewRawAPIKey() should generate unique keys")
	}
	if len(key1.Bytes()) != RawKeyCredentialLength {
		t.Errorf("Expected key length %d, got %d", RawKeyCredentialLength, len(key1.Bytes()))
	}
}

func TestRawKeyCredential_ToBase62_And_FromBase62(t *testing.T) {
	key, _ := NewRawAPIKey()
	encoded := key.ToBase62()
	if encoded == "" {
		t.Error("ToBase62() returned empty string")
	}
	decoded, err := FromBase62(encoded)
	if err != nil {
		t.Fatalf("FromBase62() error: %v", err)
	}
	if decoded != key {
		t.Error("Decoded key does not match original")
	}
}

func TestFromBase62_InvalidInput(t *testing.T) {
	_, err := FromBase62("!!!invalidbase62!!!")
	if err == nil {
		t.Error("Expected error for invalid base62 input")
	}
}

func TestRawKeyCredential_Obfuscate(t *testing.T) {
	key, _ := NewRawAPIKey()
	obfuscated := key.Obfuscate()
	if !strings.Contains(obfuscated, "...") {
		t.Errorf("Obfuscate() output missing '...': %s", obfuscated)
	}
	if len(obfuscated) < 9 {
		t.Errorf("Obfuscate() output too short: %s", obfuscated)
	}
	if obfuscated[:3] != key.ToBase62()[:3] {
		t.Errorf("Obfuscate() first three chars do not match: got %s, want %s", obfuscated[:3], key.ToBase62()[:3])
	}
	if obfuscated[len(obfuscated)-3:] != key.ToBase62()[len(key.ToBase62())-3:] {
		t.Errorf("Obfuscate() last three chars do not match: got %s, want %s", obfuscated[len(obfuscated)-3:], key.ToBase62()[len(key.ToBase62())-3:])
	}
}

func TestRawKeyCredential_Bytes(t *testing.T) {
	key, _ := NewRawAPIKey()
	b := key.Bytes()
	if len(b) != RawKeyCredentialLength {
		t.Errorf("Bytes() length mismatch: got %d, want %d", len(b), RawKeyCredentialLength)
	}
}

func TestRawKeyCredential_Hash_And_GetAPIKeyCredential(t *testing.T) {
	key, _ := NewRawAPIKey()
	if len(key.Hash()) != HashLength {
		t.Errorf("Hash() length mismatch: got %d, want %d", len(key.Hash()), HashLength)
	}
	cred := key.GetAPIKeyCredential()
	_ = cred // Just ensure it doesn't panic or error
}

func TestNewRawAPIKey_NotEmpty(t *testing.T) {
	key, err := NewRawAPIKey()
	if err != nil {
		t.Fatalf("NewRawAPIKey() error: %v", err)
	}
	zeroKey := make([]byte, RawKeyCredentialLength)
	if string(key.Bytes()) == string(zeroKey) {
		t.Error("NewRawAPIKey() generated an all-zero key")
	}
}

// == Hashed Key

func TestNewHashedKeyCredential(t *testing.T) {
	raw := [HashedKeyCredentialLength]byte{}
	for i := range raw {
		raw[i] = byte(i)
	}
	cred := NewHashedKeyCredential(raw)
	assert.Equal(t, raw[:], cred.Bytes())
}

func TestHashedKeyCredential_IsEmpty(t *testing.T) {
	var empty HashedKeyCredential
	assert.True(t, empty.IsEmpty())

	raw := [HashedKeyCredentialLength]byte{}
	raw[0] = 1
	cred := NewHashedKeyCredential(raw)
	assert.False(t, cred.IsEmpty())
}

func TestHashedKeyCredential_Equal(t *testing.T) {
	raw := [HashedKeyCredentialLength]byte{}
	for i := range raw {
		raw[i] = byte(i)
	}
	cred1 := NewHashedKeyCredential(raw)
	cred2 := NewHashedKeyCredential(raw)
	assert.True(t, cred1.Equal(cred2))

	raw2 := raw
	raw2[0]++
	cred3 := NewHashedKeyCredential(raw2)
	assert.False(t, cred1.Equal(cred3))
}
