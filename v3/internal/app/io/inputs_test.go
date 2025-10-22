package io

import (
	"context"
	"errors"
	"fmt"
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
				assert.Contains(t, err.Error(), fmt.Sprintf("unsupported chain: %s", tt.input))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
