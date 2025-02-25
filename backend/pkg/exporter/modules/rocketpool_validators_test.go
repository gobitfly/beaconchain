package modules

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestPrepareValidatorTagData(t *testing.T) {
	tests := []struct {
		name      string
		minipools map[string]*RocketpoolMinipool
		expected  []*RocketpoolMinipool
	}{
		{
			name:      "no minipools",
			minipools: map[string]*RocketpoolMinipool{},
			expected:  []*RocketpoolMinipool{},
		},
		{
			name: "single minipool",
			minipools: map[string]*RocketpoolMinipool{
				"0x1": {Pubkey: []byte("0xabc")},
			},
			expected: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
			},
		},
		{
			name: "multiple minipools",
			minipools: map[string]*RocketpoolMinipool{
				"0x1": {Pubkey: []byte("0xabc")},
				"0x2": {Pubkey: []byte("0xdef")},
			},
			expected: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
				{Pubkey: []byte("0xdef")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &RocketpoolExporter{
				MinipoolsByAddress: tt.minipools,
			}

			result := rp.prepareValidatorTagData()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d minipools, got %d", len(tt.expected), len(result))
			}

			for i, mp := range result {
				if string(mp.Pubkey) != string(tt.expected[i].Pubkey) {
					t.Errorf("Expected pubkey %s, got %s", tt.expected[i].Pubkey, mp.Pubkey)
				}
			}
		})
	}
}

func TestSaveValidatorTags(t *testing.T) {
	tests := []struct {
		name             string
		data             []*RocketpoolMinipool
		mockValTagsError error
		mockValPoolError error
		expectedError    bool
	}{
		{
			name: "single minipool",
			data: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
			},
			expectedError: false,
		},
		{
			name: "SaveValidatorTags error",
			data: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
			},
			mockValTagsError: errors.New("error"),
			expectedError:    true,
		},
		{
			name: "SaveValidatorPool error",
			data: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
			},
			mockValPoolError: errors.New("error"),
			expectedError:    true,
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
			rp := &RocketpoolExporter{}

			valueStrings, valueArgs := rp.prepareValidatorTagBatch(tt.data)

			// mock SaveValidatorTags query
			query := fmt.Sprintf(saveValidatorTagsQ,
				strings.Join(valueStrings, ","))
			if tt.mockValTagsError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(valueArgs[0]).WillReturnError(tt.mockValTagsError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(valueArgs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// mock SaveValidatorPool query
			valPoolQuery := fmt.Sprintf(saveValidatorPoolQ,
				strings.Join(valueStrings, ","))
			if tt.mockValPoolError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(valPoolQuery)).WithArgs(valueArgs[0]).WillReturnError(tt.mockValPoolError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(valPoolQuery)).WithArgs(valueArgs[0]).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.saveValidatorTags(tt.data)

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

func TestPrepareValidatorTagBatch(t *testing.T) {
	tests := []struct {
		name                 string
		data                 []*RocketpoolMinipool
		expectedValueStrings []string
		expectedValueArgs    []interface{}
	}{
		{
			name: "single minipool",
			data: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
			},
			expectedValueStrings: []string{"($1, 'rocketpool')"},
			expectedValueArgs:    []interface{}{[]byte("0xabc")},
		},
		{
			name: "multiple minipools",
			data: []*RocketpoolMinipool{
				{Pubkey: []byte("0xabc")},
				{Pubkey: []byte("0xdef")},
			},
			expectedValueStrings: []string{"($1, 'rocketpool')", "($2, 'rocketpool')"},
			expectedValueArgs:    []interface{}{[]byte("0xabc"), []byte("0xdef")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &RocketpoolExporter{}

			valueStrings, valueArgs := rp.prepareValidatorTagBatch(tt.data)

			if len(valueStrings) != len(tt.expectedValueStrings) {
				t.Errorf("Expected %d value strings, got %d", len(tt.expectedValueStrings), len(valueStrings))
			}

			if len(valueArgs) != len(tt.expectedValueArgs) {
				t.Errorf("Expected %d value args, got %d", len(tt.expectedValueArgs), len(valueArgs))
			}

			for i, vs := range valueStrings {
				if vs != tt.expectedValueStrings[i] {
					t.Errorf("Expected value string '%s', got '%s'", tt.expectedValueStrings[i], vs)
				}
			}

			for i, va := range valueArgs {
				if string(va.([]byte)) != string(tt.expectedValueArgs[i].([]byte)) {
					t.Errorf("Expected value args '%s', got '%s'", tt.expectedValueArgs[i], va)
				}
			}
		})
	}
}

var (
	saveValidatorTagsQ = `INSERT INTO validator_tags (publickey, tag) VALUES %s
				ON CONFLICT (publickey, tag) DO NOTHING`
	saveValidatorPoolQ = `INSERT INTO validator_pool (publickey, pool) VALUES %s
				ON CONFLICT (publickey) DO NOTHING`
)
