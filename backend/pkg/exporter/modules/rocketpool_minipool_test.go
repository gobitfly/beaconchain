package modules

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func TestSaveMinipools(t *testing.T) {
	tests := []struct {
		name            string
		minipools       map[string]*RocketpoolMinipool
		mockSaveError   error
		mockUpdateError error
		expectedError   bool
	}{

		{
			name: "single minipool",
			minipools: map[string]*RocketpoolMinipool{
				"0x1": {
					Address:            []byte("0x1"),
					Pubkey:             []byte("0xabc"),
					Status:             "active",
					StatusTime:         time.Now(),
					NodeAddress:        []byte("0x001"),
					NodeFee:            0.1,
					DepositType:        "full",
					PenaltyCount:       0,
					NodeDepositBalance: big.NewInt(100),
					NodeRefundBalance:  big.NewInt(0),
					UserDepositBalance: big.NewInt(100),
					IsVacant:           false,
					Version:            1,
				},
			},
			expectedError: false,
		},
		{
			name: "SaveRocketPoolMiniPools error",
			minipools: map[string]*RocketpoolMinipool{
				"0x1": {
					Address:            []byte("0x1"),
					Pubkey:             []byte("0xabc"),
					Status:             "active",
					StatusTime:         time.Now(),
					NodeAddress:        []byte("0x001"),
					NodeFee:            0.1,
					DepositType:        "full",
					PenaltyCount:       0,
					NodeDepositBalance: big.NewInt(100),
					NodeRefundBalance:  big.NewInt(0),
					UserDepositBalance: big.NewInt(100),
					IsVacant:           false,
					Version:            1,
				},
			},
			mockSaveError: errors.New("error"),
			expectedError: true,
		},
		{
			name: "UpdateRocketPoolMiniPools error",
			minipools: map[string]*RocketpoolMinipool{
				"0x1": {
					Address:            []byte("0x1"),
					Pubkey:             []byte("0xabc"),
					Status:             "active",
					StatusTime:         time.Now(),
					NodeAddress:        []byte("0x001"),
					NodeFee:            0.1,
					DepositType:        "full",
					PenaltyCount:       0,
					NodeDepositBalance: big.NewInt(100),
					NodeRefundBalance:  big.NewInt(0),
					UserDepositBalance: big.NewInt(100),
					IsVacant:           false,
					Version:            1,
				},
			},
			mockUpdateError: errors.New("error"),
			expectedError:   true,
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
				MinipoolsByAddress: tt.minipools,
				API: &rocketpool.RocketPool{
					RocketStorageContract: &rocketpool.Contract{
						Address: func() *common.Address {
							addr := common.HexToAddress("0x1")
							return &addr
						}(),
					},
				},
			}

			nArgs := 14
			valueStringsTpl := createValueStringsTemplate(nArgs)
			minipoolSlice := make([]*RocketpoolMinipool, 0, len(tt.minipools))
			for _, minipool := range tt.minipools {
				minipoolSlice = append(minipoolSlice, minipool)
			}
			valueStrings, valueArgs := rp.prepareMinipoolBatch(minipoolSlice, valueStringsTpl, nArgs)

			// mock SaveRocketPoolMiniPools query
			query := fmt.Sprintf(`
			INSERT INTO rocketpool_minipools (
				rocketpool_storage_address, 
				address, 
				pubkey, 
				status, 
				status_time, 
				node_address, 
				node_fee,
				deposit_type, 
				penalty_count, 
				node_deposit_balance, 
				node_refund_balance,
				user_deposit_balance, 
				is_vacant, 
				version
			) 
			VALUES %s 
			ON CONFLICT (rocketpool_storage_address, address) DO UPDATE SET
				pubkey = excluded.pubkey,
				status = excluded.status,
				status_time = excluded.status_time,
				node_address = excluded.node_address,
				node_fee = excluded.node_fee,
				deposit_type = excluded.deposit_type,
				penalty_count = excluded.penalty_count,
				node_deposit_balance = excluded.node_deposit_balance,
				node_refund_balance = excluded.node_refund_balance,
				user_deposit_balance = excluded.user_deposit_balance,
				is_vacant = excluded.is_vacant,
				version = excluded.version`,
				strings.Join(valueStrings, ","))

			var driverArgs = make([]driver.Value, len(valueArgs))
			for i, v := range valueArgs {
				driverArgs[i] = driver.Value(v)
			}

			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(driverArgs...).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(driverArgs...).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// mock UpdateRocketPoolMiniPools query
			updateQuery := `
		UPDATE rocketpool_minipools
		SET validator_index = validators.validatorindex
		FROM validators
		WHERE rocketpool_minipools.pubkey = validators.pubkey
	`
			if tt.mockUpdateError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WillReturnError(tt.mockUpdateError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.SaveMinipools()

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
