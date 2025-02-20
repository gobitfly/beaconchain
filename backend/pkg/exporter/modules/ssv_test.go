package modules

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestInsertSSVTags(t *testing.T) {
	tests := []struct {
		name          string
		response      *types.SSVExporterResponse
		mockError     error
		expectedError bool
	}{
		{
			name: "successful insert",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "database error",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
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

			valueStrings, valueArgs := prepareBatchInsert(tt.response.Data)

			query := fmt.Sprintf(`
				INSERT INTO validator_tags (publickey, tag)
				VALUES %s
				ON CONFLICT (publickey, tag) DO NOTHING`,
				strings.Join(valueStrings, ","))

			if tt.mockError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(valueArgs[0]).WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(valueArgs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = insertSSVTags(tt.response)

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
func TestPrepareBatchInsert(t *testing.T) {
	tests := []struct {
		name          string
		data          []types.SSVExporterData
		expectedVals  []string
		expectedArgs  []interface{}
		expectedError bool
	}{
		{
			name: "successful batch insert prep",
			data: []types.SSVExporterData{
				{Publickey: "0x123456"},
				{Publickey: "0xabcdef"},
			},
			expectedVals: []string{
				"($1, 'ssv')",
				"($2, 'ssv')",
			},
			expectedArgs: []interface{}{
				[]byte{0x12, 0x34, 0x56},
				[]byte{0xab, 0xcd, 0xef},
			},
			expectedError: false,
		},
		{
			name: "error decoding public key",
			data: []types.SSVExporterData{
				{Publickey: "0x123456"},
				{Publickey: "invalid-key"},
			},
			expectedVals: []string{
				"($1, 'ssv')",
			},
			expectedArgs: []interface{}{
				[]byte{0x12, 0x34, 0x56},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, args := prepareBatchInsert(tt.data)

			if len(vals) != len(tt.expectedVals) {
				t.Errorf("expected %d value, got %d", len(tt.expectedVals), len(vals))
				t.Errorf("got values: %s ", vals)
			}

			for i, v := range vals {
				if v != tt.expectedVals[i] {
					t.Errorf("expected value %s, got %s", tt.expectedVals[i], v)
				}
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d value args, got %d", len(tt.expectedArgs), len(args))
			}

			for i, a := range args {
				if !equal(a.([]byte), tt.expectedArgs[i].([]byte)) {
					t.Errorf("expected value arg %v, got %v", tt.expectedArgs[i], a)
				}
			}
		})
	}
}

func TestSaveSSV(t *testing.T) {
	tests := []struct {
		name                 string
		response             *types.SSVExporterResponse
		mockDeleteErr        error
		mockInsertErr        error
		mockDeleteValTagsErr error
		expectedError        bool
	}{
		{
			name: "successful save",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
			},
			mockDeleteErr:        nil,
			mockInsertErr:        nil,
			mockDeleteValTagsErr: nil,
			expectedError:        false,
		},
		{
			name: "DeleteInvalidTags error",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
			},
			mockDeleteErr: errors.New("error"),
			mockInsertErr: nil,
			expectedError: true,
		},
		{
			name: "insert SSV tags error",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
			},
			mockDeleteErr: nil,
			mockInsertErr: errors.New("insert SSV tags error"),
			expectedError: true,
		},
		{
			name: "DeleteValidatorTags error",
			response: &types.SSVExporterResponse{
				Data: []types.SSVExporterData{
					{Publickey: "0x123456"},
				},
			},
			mockDeleteErr:        nil,
			mockInsertErr:        nil,
			mockDeleteValTagsErr: errors.New("error"),
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

			// mock DeleteInvalidTags query
			deleteInvalidTagsQuery := `DELETE FROM validator_tags WHERE publickey IN (SELECT publickey FROM validator_tags WHERE tag = 'ssv' LIMIT 1000)`
			if tt.mockDeleteErr != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteInvalidTagsQuery)).WillReturnError(tt.mockDeleteErr)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteInvalidTagsQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			}

			// mock SaveValidatorTags query
			valueStrings, valueArgs := prepareBatchInsert(tt.response.Data)
			insertValTagsQuery := fmt.Sprintf(`
				INSERT INTO validator_tags (publickey, tag)
				VALUES %s
				ON CONFLICT (publickey, tag) DO NOTHING`,
				strings.Join(valueStrings, ","))

			if tt.mockInsertErr != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(insertValTagsQuery)).WithArgs(valueArgs[0]).WillReturnError(tt.mockInsertErr)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(insertValTagsQuery)).WithArgs(valueArgs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// mock DeleteValidatorTags query
			deleteValTagsQuery := `DELETE FROM validator_tags WHERE publickey IN (SELECT publickey FROM validator_tags WHERE publickey NOT IN (SELECT pubkey FROM validators) LIMIT 1000)`
			if tt.mockDeleteValTagsErr != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteValTagsQuery)).WillReturnError(tt.mockDeleteValTagsErr)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(deleteValTagsQuery)).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			}

			err = saveSSV(tt.response)

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

func equal(a, b []byte) bool {
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
