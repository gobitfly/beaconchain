package modules

import (
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestSaveNetworkStats(t *testing.T) {
	tests := []struct {
		name          string
		rpExporter    *RocketpoolExporter
		mockSaveError error
		expectedError bool
	}{
		{
			name: "valid network stats save",
			rpExporter: &RocketpoolExporter{
				NetworkStats: RocketpoolNetworkStats{
					RPLPrice:               big.NewInt(100),
					ClaimIntervalTime:      time.Duration(10),
					ClaimIntervalTimeStart: time.Now(),
					CurrentNodeFee:         0.05,
					CurrentNodeDemand:      big.NewInt(100),
					RETHSupply:             big.NewInt(100),
					NodeOperatorRewards:    big.NewInt(100),
					RETHPrice:              1.5,
					TotalEthStaking:        big.NewInt(100),
					TotalEthBalance:        big.NewInt(100),
				},
				NodesByAddress: map[string]*RocketpoolNode{
					"0x001": &RocketpoolNode{
						Address:           []byte("0x001"),
						EffectiveRPLStake: big.NewInt(100),
					},
				},
				MinipoolsByAddress: map[string]*RocketpoolMinipool{
					"0x001": {
						Address: []byte("0x1"),
					},
				},
				DAOMembersByAddress: map[string]*RocketpoolDAOMember{
					"0x001": &RocketpoolDAOMember{
						Address: []byte("0x001"),
					},
				},
			},
			expectedError: false,
		},
		{
			name:          "SaveRocketPoolNetworkStats error",
			rpExporter:    &RocketpoolExporter{},
			mockSaveError: errors.New("error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer dbMock.Close()

			sqlxDB := sqlx.NewDb(dbMock, "sqlmock")
			db.WriterDb = sqlxDB

			// mock SaveRocketPoolNetworkStats query
			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveStatsQ)).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveStatsQ)).WithArgs(
					tt.rpExporter.NetworkStats.RPLPrice.String(),
					tt.rpExporter.NetworkStats.ClaimIntervalTime.String(),
					tt.rpExporter.NetworkStats.ClaimIntervalTimeStart,
					tt.rpExporter.NetworkStats.CurrentNodeFee,
					tt.rpExporter.NetworkStats.CurrentNodeDemand.String(),
					tt.rpExporter.NetworkStats.RETHSupply.String(),
					tt.rpExporter.NetworkStats.NodeOperatorRewards.String(),
					tt.rpExporter.NetworkStats.RETHPrice,
					len(tt.rpExporter.NodesByAddress),
					len(tt.rpExporter.MinipoolsByAddress),
					len(tt.rpExporter.DAOMembersByAddress),
					tt.rpExporter.NetworkStats.TotalEthStaking.String(),
					tt.rpExporter.NetworkStats.TotalEthBalance.String(),
				).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = tt.rpExporter.SaveNetworkStats()

			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error got nil")
				}
			}

		})
	}
}

var (
	saveStatsQ = `
		INSERT INTO rocketpool_network_stats
		(
			ts, 
			rpl_price, 
			claim_interval_time, 
			claim_interval_time_start, 
			current_node_fee, 
			current_node_demand,
			reth_supply, 
			node_operator_rewards, 
			reth_exchange_rate, 
			node_count, 
			minipool_count, 
			odao_member_count,
			total_eth_staking, 
			total_eth_balance, 
			effective_rpl_staked
		)
		VALUES(
			now(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			(SELECT sum(effective_rpl_stake) FROM rocketpool_nodes)
		)`
)
