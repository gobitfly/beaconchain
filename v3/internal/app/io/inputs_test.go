package io

import (
	"context"
	"encoding/hex"
	"errors"
	"slices"
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mock repo
type mockLatestStateRepo struct {
	state domain.LatestState
	err   error
}

func (m *mockLatestStateRepo) GetLatestState(ctx context.Context, chain domain.Chain, view domain.ConsensusView) (domain.LatestState, error) {
	if m.err != nil {
		return domain.LatestState{}, m.err
	}
	return m.state, nil
}

func TestResolveSlot(t *testing.T) {
	var latestParam, finalizedParam, numberParam, badParam model.SlotParam
	err := latestParam.FromChainView("latest")
	assert.NoError(t, err)
	err = finalizedParam.FromChainView("finalized")
	assert.NoError(t, err)
	err = numberParam.FromSlot(42)
	assert.NoError(t, err)
	err = badParam.FromChainView("bad")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		input   model.SlotParam
		repo    ethereumnetworkrepo.LatestStateRepository
		want    int
		wantErr bool
	}{
		{
			name:  "head view resolves from repo",
			input: latestParam,
			repo:  &mockLatestStateRepo{state: domain.LatestState{Epoch: 10}},
			want:  10,
		},
		{
			name:  "finalized view resolves from repo",
			input: finalizedParam,
			repo:  &mockLatestStateRepo{state: domain.LatestState{Epoch: 20}},
			want:  20,
		},
		{
			name:  "numeric input parses directly",
			input: numberParam,
			repo:  &mockLatestStateRepo{},
			want:  42,
		},
		{
			name:    "invalid numeric input returns error",
			input:   badParam,
			repo:    &mockLatestStateRepo{},
			wantErr: true,
		},
		{
			name:    "repo error returned",
			input:   finalizedParam,
			repo:    &mockLatestStateRepo{err: errors.New("repo fail")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveSlot(context.Background(), tt.repo, domain.ChainMainnet, tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestResolveEpoch(t *testing.T) {
	var latestParam, finalizedParam, numberParam, badParam model.EpochParam
	err := latestParam.FromChainView("latest")
	assert.NoError(t, err)
	err = finalizedParam.FromChainView("finalized")
	assert.NoError(t, err)
	err = numberParam.FromEpoch(42)
	assert.NoError(t, err)
	err = badParam.FromChainView("bad")
	assert.NoError(t, err)
	tests := []struct {
		name    string
		input   model.EpochParam
		repo    ethereumnetworkrepo.LatestStateRepository
		want    int
		wantErr bool
	}{
		{
			name:  "head view resolves from repo",
			input: latestParam,
			repo:  &mockLatestStateRepo{state: domain.LatestState{Epoch: 10}},
			want:  10,
		},
		{
			name:  "finalized view resolves from repo",
			input: finalizedParam,
			repo:  &mockLatestStateRepo{state: domain.LatestState{Epoch: 20}},
			want:  20,
		},
		{
			name:  "numeric input parses directly",
			input: numberParam,
			repo:  &mockLatestStateRepo{},
			want:  42,
		},
		{
			name:    "invalid numeric input returns error",
			input:   badParam,
			repo:    &mockLatestStateRepo{},
			wantErr: true,
		},
		{
			name:    "repo error returned",
			input:   finalizedParam,
			repo:    &mockLatestStateRepo{err: errors.New("repo fail")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveEpoch(context.Background(), tt.repo, domain.ChainSepolia, tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAsChain(t *testing.T) {
	tests := []struct {
		name    string
		input   model.Chain
		want    domain.Chain
		wantErr bool
	}{
		{"mainnet string resolves", model.Mainnet, domain.ChainMainnet, false},
		{"empty string defaults to mainnet", model.Chain(""), domain.ChainMainnet, false},
		{"hoodi resolves", model.Hoodi, domain.ChainHoodi, false},
		{"unsupported chain errors", model.Chain("randomchain"), domain.ChainUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsChain(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

const testAddressStr = "61d1B77B20b7CAF25545117336bB172f6950DEeC"

func TestParseEthereumAddress(t *testing.T) {
	// A standard, valid 20-byte Ethereum address (40 hex characters)
	validHex := testAddressStr
	addrBytes, _ := hex.DecodeString(validHex)
	ethereumAddress := domain.EthereumAddress(addrBytes)

	tests := []struct {
		name    string
		input   string
		want    domain.EthereumAddress // Assuming domain.EthereumAddress is compatible with []byte for DeepEqual
		wantErr bool
	}{
		// --- Success Cases ---
		{
			name:    "Valid Address with 0x Prefix",
			input:   "0x" + validHex,
			want:    ethereumAddress,
			wantErr: false,
		},
		{
			name:    "Valid Address without 0x Prefix",
			input:   validHex,
			want:    ethereumAddress,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: We rely on the actual ParseEthereumAddress function being available here.
			got, err := ParseEthereumAddress(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("Invalid Decoding Characters", func(t *testing.T) {
		_, err := ParseEthereumAddress("0xGHIJKL1234567890MNOPQR1234567890STUVWXYZ")
		assert.Error(t, err)
	})
}

// setupSelectorInputs prepares various selector inputs for testing.
func setupSelectorInputs() (
	//nolint:unparam
	emptyInput model.ValidatorsSelector,
	depositAddressInput model.ValidatorsSelector,
	withdrawalAddressInput model.ValidatorsSelector,
	withdrawalCredentialInput model.ValidatorsSelector,
	identifiersInput model.ValidatorsSelector) {
	// setup deposit address input
	_ = depositAddressInput.FromValidatorsByDeposit(model.ValidatorsByDeposit{
		DepositAddress: "0x61d1B77B20b7CAF25545117336bB172f6950DEeC",
	})

	// setup withdrawal address and credential inputs
	_ = withdrawalAddressInput.FromValidatorsByWithdrawal(model.ValidatorsByWithdrawal{
		Withdrawal: testAddressStr,
	})
	_ = withdrawalCredentialInput.FromValidatorsByWithdrawal(model.ValidatorsByWithdrawal{
		Withdrawal: "010000000000000000000000" + testAddressStr,
	})

	// setup identifiers input
	var index, publicKey model.ValidatorIndexPublicKey
	_ = index.FromValidatorIndex(42)
	_ = publicKey.FromValidatorPublicKey(testAddressStr)

	_ = identifiersInput.FromValidatorsByIdentifiers(model.ValidatorsByIdentifiers{
		ValidatorIdentifiers: []model.ValidatorIndexPublicKey{index, publicKey},
	})

	return emptyInput, depositAddressInput, withdrawalAddressInput, withdrawalCredentialInput, identifiersInput
}

func TestAsDepositAddress(t *testing.T) {
	var badInput model.ValidatorsSelector
	_ = badInput.FromValidatorsByDeposit(model.ValidatorsByDeposit{
		DepositAddress: "0xInvalidHexAddress",
	})
	emptyInput, depositAddressInput, withdrawalAddressInput, withdrawalCredentialInput, identifiersInput := setupSelectorInputs()

	// success case
	t.Run("deposit address parses", func(t *testing.T) {
		selector, err := asDepositAddress(depositAddressInput)
		require.NoError(t, err)

		ethereumAddress, _ := ParseEthereumAddress(testAddressStr)
		want := domain.ValidatorsSelector{
			DepositAddress: &ethereumAddress,
		}
		assert.Equal(t, want, selector)
	})

	// error cases
	tests := []struct {
		name  string
		input model.ValidatorsSelector
	}{
		{"empty input returns error", emptyInput},
		{"withdrawal address input returns error", withdrawalAddressInput},
		{"withdrawal credential input returns error", withdrawalCredentialInput},
		{"identifiers input returns error", identifiersInput},
		{"bad deposit address returns error", badInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := asDepositAddress(tt.input)
			assert.Error(t, err)
		})
	}

}

func TestAsWithdrawalInput(t *testing.T) {
	var invalidInput, invalidLengthInput model.ValidatorsSelector
	_ = invalidInput.FromValidatorsByWithdrawal(model.ValidatorsByWithdrawal{
		Withdrawal: "invalid",
	})
	_ = invalidLengthInput.FromValidatorsByWithdrawal(model.ValidatorsByWithdrawal{
		Withdrawal: "01000" + testAddressStr,
	})

	emptyInput, depositAddressInput, withdrawalAddressInput, withdrawalCredentialInput, identifiersInput := setupSelectorInputs()

	ethereumAddress, _ := ParseEthereumAddress(testAddressStr)
	credential := domain.WithdrawalCredential(slices.Concat([]byte{0x01}, slices.Repeat([]byte{0x00}, 11), []byte(ethereumAddress)))
	// success cases
	test := []struct {
		name  string
		input model.ValidatorsSelector
		want  domain.ValidatorsSelector
	}{
		{
			name:  "withdrawal address parses",
			input: withdrawalAddressInput,
			want: domain.ValidatorsSelector{
				WithdrawalAddress: &ethereumAddress,
			},
		},
		{
			name:  "withdrawal credential parses",
			input: withdrawalCredentialInput,
			want: domain.ValidatorsSelector{
				WithdrawalCredential: &credential,
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := asWithdrawalInput(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, selector)
		})
	}

	// error cases
	tests := []struct {
		name  string
		input model.ValidatorsSelector
	}{
		{"empty input returns error", emptyInput},
		{"deposit address input returns error", depositAddressInput},
		{"identifiers input returns error", identifiersInput},
		{"invalid withdrawal input returns error", invalidInput},
		{"invalid length withdrawal input returns error", invalidLengthInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := asWithdrawalInput(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestAsIdentifiers(t *testing.T) {
	var badPublicKey, emptyPublicKey model.ValidatorIndexPublicKey
	_ = badPublicKey.FromValidatorPublicKey("0xInvalidHexPublicKey")
	_ = emptyPublicKey.FromValidatorPublicKey("")

	var badInputPublicKey, emptyPublicKeyInput model.ValidatorsSelector
	_ = badInputPublicKey.FromValidatorsByIdentifiers(model.ValidatorsByIdentifiers{
		ValidatorIdentifiers: []model.ValidatorIndexPublicKey{badPublicKey},
	})
	_ = emptyPublicKeyInput.FromValidatorsByIdentifiers(model.ValidatorsByIdentifiers{
		ValidatorIdentifiers: []model.ValidatorIndexPublicKey{emptyPublicKey},
	})
	emptyInput, depositAddressInput, withdrawalAddressInput, withdrawalCredentialInput, identifiersInput := setupSelectorInputs()

	t.Run("identifiers input parses", func(t *testing.T) {
		selector, err := asIdentifiers(identifiersInput)
		assert.NoError(t, err)
		key, err := decodeHexString(testAddressStr)
		require.NoError(t, err)
		want := domain.ValidatorsSelector{
			Identifiers: &domain.ValidatorsByIdentifier{
				Indices:    []int{42},
				PublicKeys: [][]byte{key},
			},
		}
		assert.Equal(t, want, selector)
	})

	// error cases
	tests := []struct {
		name  string
		input model.ValidatorsSelector
	}{
		{"withdrawal address input parses", withdrawalAddressInput},
		{"withdrawal credential input parses", withdrawalCredentialInput},
		{"empty input returns error", emptyInput},
		{"deposit address input returns error", depositAddressInput},
		{"bad public key returns error", badInputPublicKey},
		{"empty public key returns error", emptyPublicKeyInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := asIdentifiers(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestAsValidatorsSelector(t *testing.T) {
	emptyInput, depositAddressInput, withdrawalAddressInput, withdrawalCredentialInput, identifiersInput := setupSelectorInputs()

	ethereumAddress, _ := ParseEthereumAddress(testAddressStr)
	credential := domain.WithdrawalCredential(slices.Concat([]byte{0x01}, slices.Repeat([]byte{0x00}, 11), []byte(ethereumAddress)))

	tests := []struct {
		name    string
		input   model.ValidatorsSelector
		want    domain.ValidatorsSelector
		wantErr bool
	}{
		{
			name:  "deposit address input parses",
			input: depositAddressInput,
			want: domain.ValidatorsSelector{
				DepositAddress: &ethereumAddress,
			},
			wantErr: false,
		},
		{
			name:  "withdrawal address input parses",
			input: withdrawalAddressInput,
			want: domain.ValidatorsSelector{
				WithdrawalAddress: &ethereumAddress,
			},
			wantErr: false,
		},
		{
			name:  "withdrawal credential input parses",
			input: withdrawalCredentialInput,
			want: domain.ValidatorsSelector{
				WithdrawalCredential: &credential,
			},
			wantErr: false,
		},
		{
			name:  "identifiers input parses",
			input: identifiersInput,
			want: domain.ValidatorsSelector{
				Identifiers: &domain.ValidatorsByIdentifier{
					Indices:    []int{42},
					PublicKeys: [][]byte{ethereumAddress},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty input returns error",
			input:   emptyInput,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := AsValidatorsSelector(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, selector)
			}
		})
	}
}
