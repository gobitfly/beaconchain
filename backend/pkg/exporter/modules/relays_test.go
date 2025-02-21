package modules

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	exportertypes "github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestHandleRelayExportSuccess(t *testing.T) {
	tests := []struct {
		name          string
		relay         types.Relay
		mockError     error
		expectedError bool
	}{
		{
			name: "successful update",
			relay: types.Relay{
				ID:       "testRelayID",
				Endpoint: "http://localhost",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "update error",
			relay: types.Relay{
				ID:       "testRelayID",
				Endpoint: "http://localhost",
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

			// mock UpdateRelays query
			updateQuery := `
		UPDATE relays SET
			export_failure_count = 0,
			last_export_try_ts = NOW() AT TIME ZONE 'utc',
			last_export_success_ts = NOW() AT TIME ZONE 'utc'
		WHERE tag_id = $1 AND endpoint = $2`
			if tt.mockError != nil {
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs(tt.relay.ID, tt.relay.Endpoint).WillReturnError(tt.mockError)
			} else {
				mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs(tt.relay.ID, tt.relay.Endpoint).WillReturnResult(sqlmock.NewResult(1, 1))
			}

			var mux sync.Mutex
			err = handleRelayExportSuccess(tt.relay, &mux)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUpdateRelayExportFailure(t *testing.T) {
	tests := []struct {
		name                 string
		relay                types.Relay
		isMaxWaitTime        bool
		mockUpdateLastError  error
		mockUpdateCountError error
		expectedError        bool
	}{
		{
			name: "max wait time reached",
			relay: types.Relay{
				ID:                 "testRelayID",
				Endpoint:           "http://localhost",
				ExportFailureCount: 50,
			},
			isMaxWaitTime:       true,
			mockUpdateLastError: nil,
			expectedError:       false,
		},
		{
			name: "max wait time not reached",
			relay: types.Relay{
				ID:                 "testRelayID",
				Endpoint:           "http://localhost",
				ExportFailureCount: 5,
			},
			isMaxWaitTime:        false,
			mockUpdateCountError: nil,
			expectedError:        false,
		},
		{
			name:                "UpdateRelayLastExportTry error",
			relay:               types.Relay{},
			isMaxWaitTime:       true,
			mockUpdateLastError: errors.New("error"),
			expectedError:       true,
		},
		{
			name:                 "UpdateRelayExportFailureCount error",
			relay:                types.Relay{},
			isMaxWaitTime:        false,
			mockUpdateCountError: errors.New("error"),
			expectedError:        true,
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

			if tt.isMaxWaitTime == true {
				// mock UpdateRelayLastExportTry query
				lastExportTryQuery := `
					UPDATE relays SET
						last_export_try_ts = (NOW() AT TIME ZONE 'utc')
					WHERE tag_id = $1 AND endpoint = $2`
				if tt.mockUpdateLastError != nil {
					mock.ExpectExec(regexp.QuoteMeta(lastExportTryQuery)).WithArgs(tt.relay.ID, tt.relay.Endpoint).WillReturnError(tt.mockUpdateLastError)
				} else {
					mock.ExpectExec(regexp.QuoteMeta(lastExportTryQuery)).WithArgs(tt.relay.ID, tt.relay.Endpoint).WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			if tt.isMaxWaitTime == false {
				// mock UpdateRelayExportFailureCount query
				exportFailureCountQuery := `
	UPDATE relays SET
		export_failure_count = $1,
		last_export_try_ts = (NOW() AT TIME ZONE 'utc')
	WHERE tag_id = $2 AND endpoint = $3`
				if tt.mockUpdateCountError != nil {
					mock.ExpectExec(regexp.QuoteMeta(exportFailureCountQuery)).WithArgs(tt.relay.ExportFailureCount+1, tt.relay.ID, tt.relay.Endpoint).WillReturnError(tt.mockUpdateCountError)
				} else {
					mock.ExpectExec(regexp.QuoteMeta(exportFailureCountQuery)).WithArgs(tt.relay.ExportFailureCount+1, tt.relay.ID, tt.relay.Endpoint).WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			var mux sync.Mutex
			err = updateRelayExportFailure(tt.relay, &mux)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCalculateMinSlot(t *testing.T) {
	tests := []struct {
		name     string
		lowBound uint64
		expected uint64
	}{
		{
			name:     "lowBound greater than 10",
			lowBound: 15,
			expected: 5,
		},
		{
			name:     "lowBound equal to 10",
			lowBound: 10,
			expected: 0,
		},
		{
			name:     "lowBound less than 10",
			lowBound: 5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMinSlot(tt.lowBound)
			if result != tt.expected {
				t.Errorf("got results: %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestInsertPayloadIntoDB(t *testing.T) {
	tests := []struct {
		name                 string
		payload              exportertypes.BidTrace
		mockBlockTagsError   error
		mockBlockRelaysError error
		expectedError        bool
	}{
		{
			name: "successful insert",
			payload: exportertypes.BidTrace{
				Slot:                 12345,
				ParentHash:           "0xaaabbb",
				BlockHash:            "0xabcdef",
				BuilderPubkey:        "0x123456",
				ProposerPubkey:       "0x234567",
				ProposerFeeRecipient: "0x345678",
				GasLimit:             1000,
				GasUsed:              100,
				Value:                toWeiString("1000"),
			},
			expectedError: false,
		},
		{
			name: "SaveBlocksTags error",
			payload: exportertypes.BidTrace{
				Slot:                 12345,
				ParentHash:           "0xaaabbb",
				BlockHash:            "0xabcdef",
				BuilderPubkey:        "0x123456",
				ProposerPubkey:       "0x234567",
				ProposerFeeRecipient: "0x345678",
				GasLimit:             1000,
				GasUsed:              100,
				Value:                toWeiString("1000"),
			},
			mockBlockTagsError: errors.New("error"),
			expectedError:      true,
		},
		{
			name: "SaveBlocksRelays error",
			payload: exportertypes.BidTrace{
				Slot:                 12345,
				ParentHash:           "0xaaabbb",
				BlockHash:            "0xabcdef",
				BuilderPubkey:        "0x123456",
				ProposerPubkey:       "0x234567",
				ProposerFeeRecipient: "0x345678",
				GasLimit:             1000,
				GasUsed:              100,
				Value:                toWeiString("1000"),
			},
			mockBlockRelaysError: errors.New("error"),
			expectedError:        true,
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

			tagID := "testTagID"
			blockHash := utils.MustParseHex(tt.payload.BlockHash)
			builderPubkey := utils.MustParseHex(tt.payload.BuilderPubkey)
			proposerPubkey := utils.MustParseHex(tt.payload.ProposerPubkey)
			proposerFeeRecipient := utils.MustParseHex(tt.payload.ProposerFeeRecipient)

			// mock SaveBlocksTags query
			blockTagsQuery := `
	INSERT INTO blocks_tags
	SELECT blocks.slot, blocks.blockroot, $1
	FROM blocks
	WHERE blocks.slot = $2 AND blocks.exec_block_hash = $3
	ON CONFLICT DO NOTHING`

			if tt.mockBlockTagsError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(blockTagsQuery)).WillReturnError(tt.mockBlockTagsError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(blockTagsQuery)).WithArgs(tagID, tt.payload.Slot, blockHash).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			blockRelaysQuery := `
			INSERT INTO relays_blocks
			(
				tag_id,
				block_slot,
				block_root,
				exec_block_hash,
				value,
				builder_pubkey,
				proposer_pubkey,
				proposer_fee_recipient
			)
			SELECT
				$1,	blocks.slot, blocks.blockroot, blocks.exec_block_hash, $4, $5, $6, $7
			FROM blocks
			WHERE
				blocks.slot = $2 and
				blocks.exec_block_hash = $3
			ON CONFLICT (block_slot, block_root, tag_id) DO NOTHING`

			if tt.mockBlockRelaysError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(blockRelaysQuery)).WillReturnError(tt.mockBlockRelaysError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(blockRelaysQuery)).WithArgs(tagID, tt.payload.Slot, blockHash, tt.payload.Value, builderPubkey, proposerPubkey, proposerFeeRecipient).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = insertPayloadIntoDB(tagID, tt.payload)
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
		})
	}
}

func TestDecodePayloads(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedError bool
		expectedLen   int
	}{
		{
			name: "successful decode",
			responseBody: `[
				{
					"slot": "12345",
					"parent_hash": "0xaaabbb",
					"block_hash": "0xabcdef",
					"builder_pubkey": "0x123456",
					"proposer_pubkey": "0x234567",
					"proposer_fee_recipient": "0x345678",
					"gas_limit": "1000",
					"gas_used": "100",
					"value": "1000"
				}
			]`,
			expectedError: false,
			expectedLen:   1,
		},
		{
			name:          "empty response body",
			responseBody:  `[]`,
			expectedError: false,
			expectedLen:   0,
		},
		{
			name:          "invalid JSON",
			responseBody:  `{invalid json}`,
			expectedError: true,
			expectedLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.responseBody)),
			}

			payloads, err := decodePayloads(resp, "testTagID")
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(payloads) != tt.expectedLen {
					t.Errorf("expected length: %d, got: %d", tt.expectedLen, len(payloads))
				}
			}
		})
	}
}

func TestShouldTryToExportRelay(t *testing.T) {
	tests := []struct {
		name           string
		relay          types.Relay
		expectedResult bool
	}{
		{
			name: "no export failures",
			relay: types.Relay{
				ExportFailureCount: 0,
			},
			expectedResult: true,
		},
		{
			name: "export failures but enough time has passed",
			relay: types.Relay{
				ExportFailureCount: 3,
				LastExportTryTs:    time.Now().Add(-time.Hour),
			},
			expectedResult: true,
		},
		{
			name: "export failures and not enough time has passed",
			relay: types.Relay{
				ExportFailureCount: 3,
				LastExportTryTs:    time.Now(),
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldTryToExportRelay(tt.relay)
			if result != tt.expectedResult {
				t.Errorf("got result: %v, expected %v", result, tt.expectedResult)
			}
		})
	}
}

func TestWaitTimeToExportRelay(t *testing.T) {
	tests := []struct {
		name             string
		relay            types.Relay
		expectedWaitTime time.Duration
		expectedMaxWait  bool
	}{
		{
			name: "no export failures",
			relay: types.Relay{
				ExportFailureCount: 0,
			},
			expectedWaitTime: time.Minute,
			expectedMaxWait:  false,
		},
		{
			name: "few export failures",
			relay: types.Relay{
				ExportFailureCount: 3,
			},
			expectedWaitTime: 8 * time.Minute,
			expectedMaxWait:  false,
		},
		{
			name: "many export failures",
			relay: types.Relay{
				ExportFailureCount: 20,
			},
			expectedWaitTime: utils.Day,
			expectedMaxWait:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitTime, isMaxWaitTime := waitTimeToExportRelay(tt.relay)
			if waitTime != tt.expectedWaitTime {
				t.Errorf("got waitTime: %v, expected %v", waitTime, tt.expectedWaitTime)
			}
			if isMaxWaitTime != tt.expectedMaxWait {
				t.Errorf("got isMaxWaitTime: %v, expected %v", isMaxWaitTime, tt.expectedMaxWait)
			}
		})
	}
}

func TestShouldLogExportAsError(t *testing.T) {
	tests := []struct {
		name          string
		relay         types.Relay
		expectedError bool
	}{
		{
			name: "recent successful export",
			relay: types.Relay{
				LastExportSuccessTs: time.Now().Add(-time.Hour),
			},
			expectedError: false,
		},
		{
			name: "long time since last successful export",
			relay: types.Relay{
				LastExportSuccessTs: time.Now().Add(-2 * utils.Month),
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldLogExportAsError(tt.relay)
			if result != tt.expectedError {
				t.Errorf("got result: %v, expected %v", result, tt.expectedError)
			}
		})
	}
}

func toWeiString(value string) types.WeiString {
	weiValue := &types.WeiString{}
	err := weiValue.Set(value)
	if err != nil {
		panic(err)
	}
	return *weiValue
}
