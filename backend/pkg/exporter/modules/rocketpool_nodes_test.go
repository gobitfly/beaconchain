package modules

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func TestPrepareNodeData(t *testing.T) {
	tests := []struct {
		name     string
		nodes    map[string]*RocketpoolNode
		expected []*RocketpoolNode
	}{
		{
			name:     "no nodes",
			nodes:    map[string]*RocketpoolNode{},
			expected: []*RocketpoolNode{},
		},
		{
			name: "single node",
			nodes: map[string]*RocketpoolNode{
				"0x1": {Address: []byte("0x001")},
			},
			expected: []*RocketpoolNode{
				{Address: []byte("0x001")},
			},
		},
		{
			name: "multiple nodes",
			nodes: map[string]*RocketpoolNode{
				"0x1": {Address: []byte("0x001")},
				"0x2": {Address: []byte("0x002")},
			},
			expected: []*RocketpoolNode{
				{Address: []byte("0x001")},
				{Address: []byte("0x002")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &RocketpoolExporter{
				NodesByAddress: tt.nodes,
			}

			result := rp.prepareNodeData()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d nodes, got %d", len(tt.expected), len(result))
			}

			for i, node := range result {
				if string(node.Address) != string(tt.expected[i].Address) {
					t.Errorf("Expected address %s, got %s", tt.expected[i].Address, node.Address)
				}
			}
		})
	}
}

func TestSaveNodeData(t *testing.T) {
	tests := []struct {
		name          string
		data          []*RocketpoolNode
		mockError     error
		expectedError bool
	}{
		{
			name: "single node",
			data: []*RocketpoolNode{
				{
					Address:                []byte("0x001"),
					TimezoneLocation:       "UTC",
					RPLStake:               big.NewInt(100),
					MinRPLStake:            big.NewInt(10),
					MaxRPLStake:            big.NewInt(1000),
					RPLCumulativeRewards:   big.NewInt(1000),
					SmoothingPoolOptedIn:   true,
					ClaimedSmoothingPool:   big.NewInt(100),
					UnclaimedSmoothingPool: big.NewInt(1000),
					UnclaimedRPLRewards:    big.NewInt(1000),
					EffectiveRPLStake:      big.NewInt(1000),
					DepositCredit:          big.NewInt(1000),
				},
			},
			expectedError: false,
		},
		{
			name: "SaveRocketPoolNodes error",
			data: []*RocketpoolNode{
				{Address: []byte("0x001")},
			},
			mockError:     errors.New("error"),
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
			rp := &RocketpoolExporter{
				API: &rocketpool.RocketPool{
					RocketStorageContract: &rocketpool.Contract{
						Address: func() *common.Address {
							addr := common.HexToAddress("0x001")
							return &addr
						}(),
					},
				},
			}

			nArgs := 13
			valueStringsTpl := createValueStringsTemplate(nArgs)
			valueStrings, valueArgs := rp.prepareNodeBatch(tt.data, valueStringsTpl, nArgs)

			// mock SaveRocketPoolNodes query
			query := fmt.Sprintf(`
			INSERT INTO rocketpool_nodes (
				rocketpool_storage_address,
				address,
				timezone_location,
				rpl_stake,
				min_rpl_stake,
				max_rpl_stake,
				rpl_cumulative_rewards,
				smoothing_pool_opted_in,
				claimed_smoothing_pool,
				unclaimed_smoothing_pool,
				unclaimed_rpl_rewards,
				effective_rpl_stake,
				deposit_credit
			)
			VALUES %s
			ON CONFLICT (rocketpool_storage_address, address) DO UPDATE SET
				rpl_stake = excluded.rpl_stake,
				min_rpl_stake = excluded.min_rpl_stake,
				max_rpl_stake = excluded.max_rpl_stake,
				rpl_cumulative_rewards = excluded.rpl_cumulative_rewards,
				smoothing_pool_opted_in = excluded.smoothing_pool_opted_in,
				claimed_smoothing_pool = excluded.claimed_smoothing_pool,
				unclaimed_smoothing_pool = excluded.unclaimed_smoothing_pool,
				unclaimed_rpl_rewards = excluded.unclaimed_rpl_rewards,
				effective_rpl_stake = excluded.effective_rpl_stake,
				timezone_location = excluded.timezone_location,
				deposit_credit = excluded.deposit_credit
		`, strings.Join(valueStrings, ","))

			var driverArgs = make([]driver.Value, len(valueArgs))
			for i, v := range valueArgs {
				driverArgs[i] = driver.Value(v)
			}

			if tt.mockError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(driverArgs...).WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(driverArgs...).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.saveNodeData(tt.data)

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
