package modules

import (
	"database/sql/driver"
	"fmt"
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

func TestSaveDAOProposals(t *testing.T) {
	tests := []struct {
		name          string
		daoTProposal  []*RocketpoolDAOProposal
		mockSaveError error
		expectedError bool
	}{
		{
			name:          "successful save",
			daoTProposal:  daoProposals,
			expectedError: false,
		},
		{
			name:          "SaveRocketPoolDAOProposals error",
			daoTProposal:  daoProposals,
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

			rp := &RocketpoolExporter{
				API: &rocketpool.RocketPool{
					RocketStorageContract: &rocketpool.Contract{
						Address: func() *common.Address {
							addr := common.HexToAddress("0x001")
							return &addr
						}(),
					},
				},
				DAOProposalsByID: map[uint64]*RocketpoolDAOProposal{
					1: &daoProposal,
				},
			}

			nArgs := 18
			valueStringsTpl := createValueStringsTemplate(nArgs)
			valueStrings, valueArgs := rp.prepareDAOProposalBatch(tt.daoTProposal, valueStringsTpl, nArgs)

			saveDAOProposalsQuery := fmt.Sprintf(saveDAOProposalsQ, strings.Join(valueStrings, ","))
			args := parseArgs(valueArgs)

			// mock SaveRocketPoolDAOProposals query
			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOProposalsQuery)).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOProposalsQuery)).WithArgs(args...).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.SaveDAOProposals()

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

func TestSaveDAOProposalsMemberVotes(t *testing.T) {
	tests := []struct {
		name               string
		daoTProposalsVotes []RocketpoolDAOProposalMemberVotes
		mockSaveError      error
		expectedError      bool
	}{
		{
			name:               "successful save",
			daoTProposalsVotes: daoProposal.MemberVotes,
			expectedError:      false,
		},
		{
			name:               "SaveRocketPoolDAOProposalVotes error",
			daoTProposalsVotes: daoProposal.MemberVotes,
			mockSaveError:      errors.New("error"),
			expectedError:      true,
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
				DAOProposalsByID: map[uint64]*RocketpoolDAOProposal{
					1: &daoProposal,
				},
			}

			nArgs := 5
			valueStringsTpl := createValueStringsTemplate(nArgs)
			valueStrings, valueArgs := rp.prepareDAOProposalMemberVotesBatch(tt.daoTProposalsVotes, valueStringsTpl, nArgs)

			saveDAOProposalsVotesQuery := fmt.Sprintf(saveDAOProposalsVotesQ, strings.Join(valueStrings, ","))
			args := parseArgs(valueArgs)

			// mock SaveRocketPoolDAOProposalVotes query
			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOProposalsVotesQuery)).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveDAOProposalsVotesQuery)).WithArgs(args...).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.SaveDAOProposalsMemberVotes()

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

func parseArgs(valueArgs []interface{}) []driver.Value {
	var driverArgs = make([]driver.Value, len(valueArgs))
	for i, v := range valueArgs {
		driverArgs[i] = driver.Value(v)
	}
	return driverArgs
}

var daoProposals = []*RocketpoolDAOProposal{&daoProposal}
var daoProposal = RocketpoolDAOProposal{
	ID:              1,
	DAO:             "dao",
	ProposerAddress: []byte("0x001"),
	Message:         "test-msg",
	CreatedTime:     time.Unix(1, 0),
	StartTime:       time.Unix(2, 0),
	EndTime:         time.Unix(3, 0),
	ExpiryTime:      time.Unix(4, 0),
	VotesRequired:   1,
	VotesFor:        1,
	VotesAgainst:    1,
	MemberVoted:     true,
	MemberSupported: true,
	IsCancelled:     false,
	IsExecuted:      true,
	Payload:         []byte("payload"),
	State:           "state",
	MemberVotes: []RocketpoolDAOProposalMemberVotes{
		{
			Address:    []byte("0x001"),
			Voted:      true,
			ProposalID: 1,
			Supported:  true,
		},
	},
}

var (
	saveDAOProposalsQ = `INSERT INTO rocketpool_dao_proposals (
				rocketpool_storage_address, 
				id, 
				dao, 
				proposer_address,
				message, 
				created_time, 
				start_time, 
				end_time, 
				expiry_time, 
				votes_required, 
				votes_for, 
				votes_against, 
				member_voted, 
				member_supported, 
				is_cancelled, 
				is_executed, 
				payload, 
				state
			) 
			VALUES %s 
			ON CONFLICT (rocketpool_storage_address, id) DO UPDATE SET 
				dao = excluded.dao, 
				proposer_address = excluded.proposer_address, 
				message = excluded.message, 
				created_time = excluded.created_time, 
				start_time = excluded.start_time, 
				end_time = excluded.end_time, 
				expiry_time = excluded.expiry_time, 
				votes_required = excluded.votes_required, 
				votes_for = excluded.votes_for, 
				votes_against = excluded.votes_against, 
				member_voted = excluded.member_voted, 
				member_supported = excluded.member_supported, 
				is_cancelled = excluded.is_cancelled, 
				is_executed = excluded.is_executed, 
				payload = excluded.payload, 
				state = excluded.state
			`
	saveDAOProposalsVotesQ = `
			INSERT INTO rocketpool_dao_proposals_member_votes (
				rocketpool_storage_address, 
				id, 
				member_address, 
				voted, 
				supported
			)
			VALUES %s 
			ON CONFLICT (rocketpool_storage_address, id, member_address) DO UPDATE SET
				voted = excluded.voted,
				supported = excluded.supported
			`
)
