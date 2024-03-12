package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
)

type ValidatorStatus string

func (s ValidatorStatus) IsActive() bool {
	return s == ActiveOngoing || s == ActiveExiting || s == ActiveSlashed || s == Active
}

type bytesHexStr []byte

func (s *bytesHexStr) UnmarshalText(b []byte) error {
	if s == nil {
		return fmt.Errorf("cannot unmarshal bytes into nil")
	}
	if len(b) >= 2 && b[0] == '0' && b[1] == 'x' {
		b = b[2:]
	}
	out := make([]byte, len(b)/2)
	_, err := hex.Decode(out, b)
	if err != nil {
		return fmt.Errorf("error unmarshalling text: %w", err)
	}
	*s = out
	return nil
}

func (s *bytesHexStr) String() string {
	return fmt.Sprintf("0x%x", *s)
}

type Uint64Str uint64

func (s *Uint64Str) UnmarshalJSON(b []byte) error {
	return Uint64Unmarshal((*uint64)(s), b)
}

// Parse a uint64, with or without quotes, in any base, with common prefixes accepted to change base.
func Uint64Unmarshal(v *uint64, b []byte) error {
	if v == nil {
		return errors.New("nil dest in uint64 decoding")
	}
	if len(b) == 0 {
		return errors.New("empty uint64 input")
	}
	if b[0] == '"' || b[0] == '\'' {
		if len(b) == 1 || b[len(b)-1] != b[0] {
			return errors.New("uneven/missing quotes")
		}
		b = b[1 : len(b)-1]
	}
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = n
	return nil
}
