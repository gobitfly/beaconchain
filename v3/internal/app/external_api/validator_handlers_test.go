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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetValidatorBalances(t *testing.T) {

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

	epoch := 100
	ethereumNetworkRepo.On("GetLatestState", mock.Anything, domain.ChainMainnet, domain.ConsensusViewFinalized).Return(domain.LatestState{
		Epoch: epoch,
	}, nil)
	ethereumNetworkRepo.On("GetLatestState", mock.Anything, domain.ChainMainnet, domain.ConsensusViewFinalized).Return(domain.LatestState{
		Epoch: epoch,
	}, nil)

	balanceStr := "32000000000"
	d, _ := decimal.NewFromString(balanceStr)
	valiRepo.On("GetBalances", mock.Anything, domain.ChainMainnet, io.EpochToStartTimestamp(epoch, chainConfigs[domain.ChainMainnet]), mock.Anything, (*domain.ValidatorIndexCursor)(nil), 2).Return([]domain.ValidatorBalance{
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

	service, _ := InitDependencies(chainConfigs, nil, ethereumNetworkRepo, valiRepo)
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
	data := *resp.Data
	assert.Len(t, data, 1)
	assert.Equal(t, 1, data[0].Validator.Index)
	assert.Equal(t, "0x01", data[0].Validator.PublicKey)
	assert.Equal(t, balanceStr, data[0].Balance.Current)
	assert.Equal(t, balanceStr, data[0].Balance.Effective)

	require.NotNil(t, resp.Range)
	assert.Equal(t, resp.Range.Epoch.Start, 100)

	require.NotNil(t, resp.Paging)
	assert.NotEmpty(t, resp.Paging.NextCursor)
}
