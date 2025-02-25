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
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func TestSaveDAOMembers(t *testing.T) {
	tests := []struct {
		name            string
		daoTMembers     []*RocketpoolDAOMember
		mockSaveError   error
		mockDeleteError error
		expectedError   bool
	}{
		{
			name:          "successful save and delete",
			daoTMembers:   daoMembers,
			expectedError: false,
		},
		{
			name:          "SaveRocketPoolDAOMembers error",
			daoTMembers:   daoMembers,
			mockSaveError: errors.New("error"),
			expectedError: true,
		},
		{
			name:            "DeleteRocketPoolDAOMembers error",
			daoTMembers:     daoMembers,
			mockDeleteError: errors.New("error"),
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
				API: &rocketpool.RocketPool{
					RocketStorageContract: &rocketpool.Contract{
						Address: func() *common.Address {
							addr := common.HexToAddress("0x001")
							return &addr
						}(),
					},
				},
				DAOMembersByAddress: map[string]*RocketpoolDAOMember{
					"0x001": &daoMember,
				},
			}

			nArgs := 8
			valueStringsTpl := createValueStringsTemplate(nArgs)
			valueStrings, valueArgs, addresses := rp.prepareDAOMemberBatch(tt.daoTMembers, valueStringsTpl, nArgs)

			saveDAOMembersQuery := fmt.Sprintf(saveDAOMembersQ, strings.Join(valueStrings, ","))
			var driverArgs = make([]driver.Value, len(valueArgs))
			for i, v := range valueArgs {
				driverArgs[i] = driver.Value(v)
			}

			// mock SaveRocketPoolDAOMembers query
			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOMembersQuery)).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOMembersQuery)).WithArgs(driverArgs...).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// mock DeleteRocketPoolDAOMembers query
			if tt.mockDeleteError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteDAOMembersQ)).WillReturnError(tt.mockDeleteError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteDAOMembersQ)).WithArgs(pq.Array(addresses)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.SaveDAOMembers()

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

var daoMembers = []*RocketpoolDAOMember{&daoMember}
var daoMember = RocketpoolDAOMember{
	Address:                []byte("0x001"),
	ID:                     "1",
	URL:                    "url",
	JoinedTime:             time.Unix(1, 0),
	LastProposalTime:       time.Unix(2, 0),
	RPLBondAmount:          big.NewInt(1),
	UnbondedValidatorCount: 1,
}

var (
	saveDAOMembersQ = `
			INSERT INTO rocketpool_dao_members (
				rocketpool_storage_address,
				address,
				id,
				url,
				joined_time,
				last_proposal_time,
				rpl_bond_amount,
				unbonded_validator_count
			)
			VALUES %s
			ON CONFLICT (rocketpool_storage_address, address) DO UPDATE SET
				id = excluded.id,
				url = excluded.url,
				joined_time = excluded.joined_time,
				last_proposal_time = excluded.last_proposal_time,
				rpl_bond_amount = excluded.rpl_bond_amount,
				unbonded_validator_count = excluded.unbonded_validator_count`

	deleteDAOMembersQ = `DELETE FROM rocketpool_dao_members WHERE NOT address = ANY($1)`
)
