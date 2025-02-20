package modules

import (
	"testing"

	"github.com/coocood/freecache"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

func TestCalculateStateIDAndEpoch(t *testing.T) {
	tests := []struct {
		name          string
		period        uint64
		config        *types.Config
		expectedID    uint64
		expectedEpoch uint64
	}{
		{
			name:   "period > 0",
			period: 1,
			config: &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              4,
						EpochsPerSyncCommitteePeriod: 2,
						SlotsPerEpoch:                10,
					},
				},
			},
			expectedID:    40,
			expectedEpoch: 4,
		},
		{
			name:   "period == 0",
			period: 0,
			config: &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              1,
						EpochsPerSyncCommitteePeriod: 1,
						SlotsPerEpoch:                10,
					},
				},
			},
			expectedID:    10,
			expectedEpoch: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = tt.config
			id, epoch := calculateStateIDAndEpoch(tt.period)
			if id != tt.expectedID {
				t.Errorf("expected ID: %v, got: %v", tt.expectedID, id)
			}
			if epoch != tt.expectedEpoch {
				t.Errorf("expected epoch: %v, got: %v", tt.expectedEpoch, epoch)
			}
		})
	}
}

func TestCalculateStateID(t *testing.T) {
	tests := []struct {
		name       string
		period     uint64
		expectedID uint64
	}{
		{
			name:       "period > 0",
			period:     1,
			expectedID: 384,
		},
		{
			name:       "period == 0",
			period:     0,
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              12,
						EpochsPerSyncCommitteePeriod: 12,
						SlotsPerEpoch:                32,
					},
				},
			}

			stateID := calculateStateID(tt.period)
			if stateID != tt.expectedID {
				t.Errorf("expected ID: %v, got: %v", tt.expectedID, stateID)
			}
		})
	}
}

func TestCalculateSyncPeriodRange(t *testing.T) {
	tests := []struct {
		name                string
		latestEpoch         uint64
		expectedFirstPeriod uint64
		expectedLastPeriod  uint64
	}{
		{
			name:                "latest epoch > 0",
			latestEpoch:         10,
			expectedFirstPeriod: 0,
			expectedLastPeriod:  2,
		},
		{
			name:                "latest epoch = 0",
			latestEpoch:         0,
			expectedFirstPeriod: 0,
			expectedLastPeriod:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              8,
						EpochsPerSyncCommitteePeriod: 1,
						SlotsPerEpoch:                32,
						DepositChainID:               1,
					},
				},
			}

			cache.TieredCache = &cache.TieredCacheBase{
				LocalGoCache: freecache.NewCache(100 * 1024 * 1024), // 100 MB
				RemoteCache:  &MockRemoteCache{},
			}

			err := cache.LatestFinalizedEpoch.Set(tt.latestEpoch)
			if err != nil {
				t.Errorf("unexpected latest finalized epoch error: %v", err)
			}

			firstPeriod, lastPeriod := calculateSyncPeriodRange()
			if firstPeriod != tt.expectedFirstPeriod {
				t.Errorf("expected first period: %v, got: %v", tt.expectedFirstPeriod, firstPeriod)
			}
			if lastPeriod != tt.expectedLastPeriod {
				t.Errorf("expected last period: %v, got: %v", tt.expectedLastPeriod, lastPeriod)
			}
		})
	}
}

func TestParseSyncCommitteeResult(t *testing.T) {
	tests := []struct {
		name           string
		input          *constypes.StandardSyncCommittee
		period         uint64
		expectedResult []SyncCommittee
	}{
		{
			name:   "valid input",
			input:  &constypes.StandardSyncCommittee{Validators: []constypes.Uint64Str{0, 1}},
			period: 1,
			expectedResult: []SyncCommittee{
				{
					Period:         1,
					ValidatorIndex: 0,
					CommitteeIndex: 0,
				},
				{
					Period:         1,
					ValidatorIndex: 1,
					CommitteeIndex: 1,
				},
			},
		},
		{
			name:           "empty input",
			input:          &constypes.StandardSyncCommittee{Validators: []constypes.Uint64Str{}},
			period:         1,
			expectedResult: []SyncCommittee{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncCommittee := parseSyncCommitteeResult(tt.input, tt.period)
			if !compareSyncCommittees(syncCommittee, tt.expectedResult) {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, syncCommittee)
			}
		})
	}
}

func TestParseSyncArgsAndIDs(t *testing.T) {
	tests := []struct {
		name         string
		data         []SyncCommittee
		expectedArgs []interface{}
		expectedIDs  []string
	}{
		{
			name: "single sync committee entry",
			data: []SyncCommittee{
				{
					Period:         1,
					ValidatorIndex: 2,
					CommitteeIndex: 3,
				},
			},
			expectedArgs: []interface{}{
				uint64(1),
				uint64(2),
				uint64(3),
			},
			expectedIDs: []string{"($1,$2,$3)"},
		},
		{
			name: "multiple sync committee entries",
			data: []SyncCommittee{
				{
					Period:         1,
					ValidatorIndex: 2,
					CommitteeIndex: 3,
				},
				{
					Period:         4,
					ValidatorIndex: 5,
					CommitteeIndex: 6,
				},
			},
			expectedArgs: []interface{}{
				uint64(1),
				uint64(2),
				uint64(3),
				uint64(4),
				uint64(5),
				uint64(6),
			},
			expectedIDs: []string{"($1,$2,$3)", "($4,$5,$6)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, ids := parseSyncArgsAndIDs(tt.data)
			if !compareInterfaces(args, tt.expectedArgs) {
				t.Errorf("expected args: %v, got: %v", tt.expectedArgs, args)
			}
			if !compareStrings(ids, tt.expectedIDs) {
				t.Errorf("expected IDs: %v, got: %v", tt.expectedIDs, ids)
			}
		})
	}
}

func compareInterfaces(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareSyncCommittees(a, b []SyncCommittee) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
