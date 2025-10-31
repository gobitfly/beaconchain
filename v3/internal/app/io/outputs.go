package io

import (
	"encoding/hex"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/oapi-codegen/nullable"
)

const (
	Finalized    model.FinalityParams = "finalized"
	NotFinalized model.FinalityParams = "not_finalized"
)

var (
	OnlyFinalized    model.FinalityParamsOnlyFinalized      = model.FinalityParamsOnlyFinalized(Finalized)
	OnlyNotFinalized model.FinalityParamsOnlyNoFinalization = model.FinalityParamsOnlyNoFinalization(NotFinalized)
)

// FormatFinalized formats the finality status based on the requested and latest finalized epochs.
func FormatFinalized(requestedEpoch, latestFinalizedEpoch int) model.FinalityParams {
	if requestedEpoch < latestFinalizedEpoch {
		return NotFinalized
	}
	return Finalized
}

// EncodeHexString encodes a byte slice to a hex string with "0x" prefix.
func EncodeHexString(data []byte) string {
	return "0x" + hex.EncodeToString(data)
}

// AsNullable converts a pointer to a nullable.Nullable type.
func AsNullable[Pointer *T, T any](input Pointer) nullable.Nullable[T] {
	if input != nil {
		return nullable.NewNullableWithValue(*input)
	}
	return nullable.NewNullNullable[T]()
}

// AsNullableValue converts a value to a nullable.Nullable type.
func AsNullableValue[T any](input T) nullable.Nullable[T] {
	return nullable.NewNullableWithValue(input)
}

// AsModelValidator converts domain.ValidatorIndex and domain.PublicKey to model.Validator.
func AsModelValidator(index domain.ValidatorIndex, publicKey domain.PublicKey) model.Validator {
	return model.Validator{
		Index:     AsNullableValue(index),
		PublicKey: EncodeHexString(publicKey),
	}
}

// AsWithdrawalCredential converts byte slice withdrawal cred to model.WithdrawalCredential.
func AsWithdrawalCredential(cred domain.WithdrawalCredential) (model.WithdrawalCredential, error) {
	if len(cred) != 32 {
		return model.WithdrawalCredential{}, fmt.Errorf("invalid withdrawal credential: expected 32 bytes, got %v", cred)
	}
	prefixStr := model.WithdrawalCredentialPrefix(EncodeHexString(cred[:1]))
	typeStr := model.Bls
	var addr model.ExecutionLayerAddress
	if prefixStr != model.N0x00 {
		typeStr = model.ExecutionAddress
		addr = EncodeHexString(cred[12:])
	}

	return model.WithdrawalCredential{
		Address:    addr,
		Credential: EncodeHexString(cred),
		Prefix:     prefixStr,
		Type:       typeStr,
	}, nil
}
