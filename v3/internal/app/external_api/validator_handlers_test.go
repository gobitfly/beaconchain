package app

import (
	"context"
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/common/config"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/ethereumnetworkrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/validatorrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/oapi-codegen/nullable"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initTestDependencies() (*ApiService, *validatorrepo.MockRepository, *ethereumnetworkrepo.MockRepository) {
	valiRepo := &validatorrepo.MockRepository{}
	ethereumNetworkRepo := &ethereumnetworkrepo.MockRepository{}

	chainConfigs := map[domain.Chain]config.ChainConfig{
		domain.ChainMainnet: {
			ID:               1,
			GenesisTimestamp: 0,
			SecondsPerSlot:   12,
			SlotsPerEpoch:    32,
		},
	}

	service, _ := InitDependencies(chainConfigs, nil, ethereumNetworkRepo, valiRepo)
	return service, valiRepo, ethereumNetworkRepo
}

// TODO: use new instead once go 1.26 is used
func ptr[T any](v T) *T {
	return &v
}

func TestValidatorStatus(t *testing.T) {

	// Define all test cases in a slice of structs
	tests := []struct {
		name                       string
		latestEpoch                int
		activationEligibilityEpoch *int
		activationEpoch            *int
		exitEpoch                  *int
		withdrawableEpoch          *int
		slashed                    bool
		balance                    decimal.Decimal
		expectedStatus             model.DataStatus
	}{
		// --- PENDING STATUSES ---
		{
			name:           "Pending Initialized",
			latestEpoch:    100,
			expectedStatus: model.PendingInitialized,
		},
		{
			name:                       "Pending queued, only activation eligibility epoch set",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(120),
			expectedStatus:             model.PendingQueued,
		},
		{
			name:                       "Pending Queued, activation epoch in future",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(120),
			expectedStatus:             model.PendingQueued,
		},
		{
			name:                       "Pending Active, activation epoch is next epoch",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(101),
			expectedStatus:             model.PendingQueued,
		},
		// --- ACTIVE STATUSES ---
		{
			name:                       "Active Ongoing",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			expectedStatus:             model.ActiveOngoing,
		},
		{
			name:                       "Active Ongoing, activation epoch is current epoch",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(100),
			expectedStatus:             model.ActiveOngoing,
		},
		{
			name:                       "Active Exiting",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			expectedStatus:             model.ActiveExiting,
		},
		{
			name:                       "Active Slashed",
			latestEpoch:                100,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(120),
			slashed:                    true,
			expectedStatus:             model.ActiveSlashed,
		},
		// --- EXITED STATUSES ---
		{
			name:                       "Exited Unslashed",
			latestEpoch:                150,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			slashed:                    false,
			expectedStatus:             model.ExitedUnslashed,
		},
		{
			name:                       "Exited Unslashed, exit epoch is current epoch",
			latestEpoch:                110,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			expectedStatus:             model.ExitedUnslashed,
		},
		{
			name:                       "Exited Unslashed, Withdrawal in future",
			latestEpoch:                130,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(150),
			expectedStatus:             model.ExitedUnslashed,
		},
		{
			name:                       "Exited Unslashed, Withdrawal in next epoch",
			latestEpoch:                129,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			expectedStatus:             model.ExitedUnslashed,
		},
		{
			name:                       "Exited Slashed",
			latestEpoch:                150,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			slashed:                    true,
			expectedStatus:             model.ExitedSlashed,
		},
		{
			name:                       "Exited Slashed, exit epoch is current epoch",
			latestEpoch:                110,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			slashed:                    true,
			expectedStatus:             model.ExitedSlashed,
		},
		{
			name:                       "Exited Slashed, Withdrawal in future",
			latestEpoch:                130,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(150),
			slashed:                    true,
			expectedStatus:             model.ExitedSlashed,
		},
		{
			name:                       "Exited Slashed, Withdrawal in next epoch",
			latestEpoch:                129,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			slashed:                    true,
			expectedStatus:             model.ExitedSlashed,
		},
		// --- WITHDRAWABLE STATUSES ---
		{
			name:                       "Withdrawal Possible",
			latestEpoch:                150,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			balance:                    decimal.NewFromInt(1),
			expectedStatus:             model.WithdrawalPossible,
		},
		{
			name:                       "Withdrawal Possible, current epoch is withdrawable epoch",
			latestEpoch:                130,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			balance:                    decimal.NewFromInt(1),
			expectedStatus:             model.WithdrawalPossible,
		},
		{
			name:                       "Withdrawal Done",
			latestEpoch:                150,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			balance:                    decimal.Zero,
			expectedStatus:             model.WithdrawalDone,
		},
		{
			name:                       "Withdrawal Done, current epoch is withdrawable epoch",
			latestEpoch:                130,
			activationEligibilityEpoch: ptr(80),
			activationEpoch:            ptr(90),
			exitEpoch:                  ptr(110),
			withdrawableEpoch:          ptr(130),
			balance:                    decimal.Zero,
			expectedStatus:             model.WithdrawalDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatorStatus(
				tt.latestEpoch,
				tt.activationEligibilityEpoch,
				tt.activationEpoch,
				tt.exitEpoch,
				tt.withdrawableEpoch,
				tt.slashed,
				tt.balance,
			)

			// Use testify/assert to check the result
			assert.Equal(t, tt.expectedStatus, result, "The returned status should match the expected status.")
		})
	}
}

func TestGetValidatorOverview(t *testing.T) {

	service, valiRepo, ethereumNetworkRepo := initTestDependencies()

	pageSize := 1
	network := domain.ChainMainnet
	//nolint:gosec
	withdrawalCredStr := "0x00d8cfc9e2ad8166ce4b610373d3ccba57ceaddea3498eb8ec3ddb90202d87b5"
	withdrawalCred, err := io.DecodeHexString(withdrawalCredStr)
	require.NoError(t, err)
	valiRepo.On("GetOverview", mock.Anything, network, mock.Anything, mock.Anything, pageSize+1).Return([]domain.ValidatorOverview{
		{
			ValidatorIndex:             1,
			ValidatorPublicKey:         []byte{0x01},
			Slashed:                    false,
			Online:                     ptr(true),
			WithdrawalCredential:       withdrawalCred,
			ActivationEligibilityEpoch: ptr(50),
			ActivationEpoch:            ptr(60),
		},
		{
			// entry will be trimmed due to pagination
			ValidatorIndex: 2,
		},
	}, nil)

	epoch := 100
	ethereumNetworkRepo.On("GetLatestState", mock.Anything, domain.ChainMainnet, domain.ConsensusViewHead).Return(domain.LatestState{
		Epoch: epoch,
	}, nil)
	balanceStr := "32000000000"
	d, _ := decimal.NewFromString(balanceStr)
	valiRepo.On("GetHeadBalances", mock.Anything, network, io.EpochToStartTime(epoch, service.chainConfigs[network]), []domain.ValidatorIndex{1}).Return([]domain.ValidatorBalance{
		{
			ValidatorIndex:     1,
			ValidatorPublicKey: []byte{0x01},
			CurrentBalance:     d,
			EffectiveBalance:   d,
		},
	}, nil)

	var validatorSelector model.ValidatorsSelector
	_ = validatorSelector.FromValidatorsByDeposit(model.ValidatorsByDeposit{DepositAddress: "0x01"})
	request := model.GetValidatorOverviewRequestObject{
		Body: &model.GetValidatorOverviewJSONRequestBody{
			Chain:     model.Chain(network),
			PageSize:  pageSize,
			Validator: validatorSelector,
		},
	}

	rawResp, err := service.GetValidatorOverview(context.Background(), request)
	require.NoError(t, err)
	resp, ok := rawResp.(model.GetValidatorOverview200JSONResponse)
	require.True(t, ok)
	require.NotNil(t, resp.Data)
	data := resp.Data
	require.Len(t, data, 1)

	entry := data[0]
	assert.Equal(t, nullable.NewNullableWithValue(1), entry.Validator.Index)
	assert.Equal(t, "0x01", entry.Validator.PublicKey)
	assert.Equal(t, balanceStr, entry.Balances.Current)
	assert.Equal(t, balanceStr, entry.Balances.Effective)
	assert.Equal(t, entry.Status, model.ActiveOngoing)

	assert.Equal(t, entry.WithdrawalCredentials.Credential, withdrawalCredStr)

	require.NotNil(t, resp.Range)
	assert.Equal(t, resp.Range.Epoch.Start, 100)

	require.NotNil(t, resp.Paging)
	assert.NotEmpty(t, resp.Paging.NextCursor)
}

func TestGetValidatorBalances(t *testing.T) {

	service, valiRepo, ethereumNetworkRepo := initTestDependencies()

	epoch := 100
	ethereumNetworkRepo.On("GetLatestState", mock.Anything, domain.ChainMainnet, domain.ConsensusViewFinalized).Return(domain.LatestState{
		Epoch: epoch,
	}, nil)
	ethereumNetworkRepo.On("GetLatestState", mock.Anything, domain.ChainMainnet, domain.ConsensusViewFinalized).Return(domain.LatestState{
		Epoch: epoch,
	}, nil)

	balanceStr := "32000000000"
	d, _ := decimal.NewFromString(balanceStr)
	valiRepo.On("GetBalances", mock.Anything, domain.ChainMainnet, io.EpochToStartTime(epoch, service.chainConfigs[domain.ChainMainnet]), mock.Anything, (*domain.ValidatorIndexCursor)(nil), 2).Return([]domain.ValidatorBalance{
		{
			ValidatorIndex:     1,
			ValidatorPublicKey: []byte{0x01},
			CurrentBalance:     d,
			EffectiveBalance:   d,
		},
		{
			ValidatorIndex:     2,
			ValidatorPublicKey: []byte{0x01},
			CurrentBalance:     d,
			EffectiveBalance:   d,
		},
	}, nil)
	var epochParam model.EpochParam
	_ = epochParam.FromChainView("finalized")
	var validatorSelector model.ValidatorsSelector
	_ = validatorSelector.FromValidatorsByDeposit(model.ValidatorsByDeposit{DepositAddress: "0x01"})
	request := model.GetValidatorBalancesRequestObject{
		Body: &model.GetValidatorBalancesJSONRequestBody{
			Chain:     model.Mainnet,
			Epoch:     epochParam,
			PageSize:  1,
			Validator: validatorSelector,
		},
	}

	rawResp, err := service.GetValidatorBalances(context.Background(), request)
	require.NoError(t, err)
	resp, ok := rawResp.(model.GetValidatorBalances200JSONResponse)
	require.True(t, ok)
	require.NotNil(t, resp.Data)
	data := resp.Data
	assert.Len(t, data, 1)
	assert.Equal(t, nullable.NewNullableWithValue(1), data[0].Validator.Index)
	assert.Equal(t, "0x01", data[0].Validator.PublicKey)
	assert.Equal(t, balanceStr, data[0].Balance.Current)
	assert.Equal(t, balanceStr, data[0].Balance.Effective)

	require.NotNil(t, resp.Range)
	assert.Equal(t, resp.Range.Epoch.Start, 100)

	require.NotNil(t, resp.Paging)
	assert.NotEmpty(t, resp.Paging.NextCursor)
}
