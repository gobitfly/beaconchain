package modules

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

func TestDownloadRewardsFile(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		interval       uint64
		cid            string
		isDaemon       bool
		mockResponses  []*http.Response
		mockErrors     []error
		expectedError  bool
		expectedOutput []byte
	}{
		{
			name:     "successful download from ipfs.dweb URL",
			fileName: "rewards-file",
			interval: 12345,
			cid:      "test",
			isDaemon: false,
			mockResponses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("{index:1}")))),
				},
			},
			mockErrors:     []error{nil},
			expectedError:  false,
			expectedOutput: []byte("{index:1}"),
		},
		{
			name:     "successful download from ipfs.io URL",
			fileName: "rewards-file",
			interval: 12345,
			cid:      "test",
			isDaemon: false,
			mockResponses: []*http.Response{
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("not found")))),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("{index:1}")))),
				},
			},
			mockErrors:     []error{nil, nil},
			expectedError:  false,
			expectedOutput: []byte("{index:1}"),
		},
		{
			name:     "successful download from github URL",
			fileName: "rewards-file",
			interval: 12345,
			cid:      "test",
			isDaemon: false,
			mockResponses: []*http.Response{
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("not found")))),
				},
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("not found")))),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("{index:1}"))),
				},
			},
			mockErrors:     []error{nil, nil, nil},
			expectedError:  false,
			expectedOutput: []byte("{index:1}"),
		},
		{
			name:     "all URLs fail",
			fileName: "rewards-file",
			interval: 12345,
			cid:      "test",
			isDaemon: false,
			mockResponses: []*http.Response{
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("not found")))),
				},
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(createZSTData(t, []byte("not found")))),
				},
				{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader([]byte("not found"))),
				},
			},
			mockErrors:     []error{nil, nil, nil},
			expectedError:  true,
			expectedOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{
				Transport: &mockTransport{
					responses: tt.mockResponses,
					errors:    tt.mockErrors,
				},
			}
			output, err := DownloadRewardsFile(tt.fileName, tt.interval, tt.cid, tt.isDaemon, client)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !bytes.Equal(output, tt.expectedOutput) {
					t.Errorf("expected output %v, got %v", tt.expectedOutput, output)
				}
			}
		})
	}
}

func createZSTData(t *testing.T, input []byte) []byte {
	var buf bytes.Buffer

	compressor, err := zstd.NewWriter(&buf)
	if err != nil {
		t.Fatalf("failed to create zstd compressor: %v", err)
	}

	_, err = compressor.Write(input)
	if err != nil {
		t.Fatalf("failed to write compressed data: %v", err)
	}

	err = compressor.Close()
	if err != nil {
		t.Fatalf("failed to close zstd compressor: %v", err)
	}

	return buf.Bytes()
}

type mockTransport struct {
	responses []*http.Response
	errors    []error
	index     int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.index >= len(m.responses) {
		return nil, errors.New("no mock responses")
	}
	resp := m.responses[m.index]
	err := m.errors[m.index]
	m.index++
	return resp, err
}

func TestSaveRewardTrees(t *testing.T) {
	tests := []struct {
		name              string
		downloadQueue     []RocketpoolRewardTreeDownloadable
		mockSaveError     error
		mockRefreshError  error
		mockGetTreesError error
		expectedError     bool
	}{
		{
			name: "successful save and refresh",
			downloadQueue: []RocketpoolRewardTreeDownloadable{
				{ID: 1, Data: []byte("data1")},
				{ID: 2, Data: []byte("data2")},
			},
			expectedError: false,
		},
		{
			name:          "empty download queue",
			downloadQueue: []RocketpoolRewardTreeDownloadable{},
			expectedError: false,
		},
		{
			name: "SaveRocketPoolRewardTree error",
			downloadQueue: []RocketpoolRewardTreeDownloadable{
				{ID: 1, Data: []byte("data1")},
			},
			mockSaveError: errors.New("error"),
			expectedError: true,
		},
		{
			name: "RefreshRocketPoolMV error",
			downloadQueue: []RocketpoolRewardTreeDownloadable{
				{ID: 1, Data: []byte("data1")},
			},
			mockRefreshError: errors.New("error"),
			expectedError:    true,
		},
		{
			name: "GetRocketPoolRewardTrees error",
			downloadQueue: []RocketpoolRewardTreeDownloadable{
				{ID: 1, Data: []byte("data1")},
			},
			mockGetTreesError: errors.New("error"),
			expectedError:     true,
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
			db.ReaderDb = sqlxDB

			rp := &RocketpoolExporter{
				RocketpoolRewardTreesDownloadQueue: tt.downloadQueue,
			}

			// mock SaveRocketPoolRewardTree query
			for _, rewardTree := range tt.downloadQueue {
				if tt.mockSaveError != nil {
					mock.ExpectBegin()
					mock.ExpectExec(regexp.QuoteMeta(saveRewardTreeQ)).
						WithArgs(rewardTree.ID, rewardTree.Data).
						WillReturnError(tt.mockSaveError)
					mock.ExpectRollback()
				} else {
					mock.ExpectBegin()
					mock.ExpectExec(regexp.QuoteMeta(saveRewardTreeQ)).
						WithArgs(rewardTree.ID, rewardTree.Data).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			}

			// mock CheckRocketPoolMVExists & RefreshRocketPoolMV query
			if tt.mockRefreshError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(checkIfExistsQ)).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(regexp.QuoteMeta(refreshQ)).
					WillReturnError(tt.mockRefreshError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(checkIfExistsQ)).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectExec(regexp.QuoteMeta(refreshQ)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			// mock GetRocketPoolRewardTrees query
			if tt.mockGetTreesError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(getTreesQ)).
					WillReturnError(tt.mockGetTreesError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(getTreesQ)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "data"}))
			}

			err = rp.SaveRewardTrees()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}

var (
	saveRewardTreeQ = `INSERT INTO rocketpool_reward_tree (id, data) VALUES($1, $2) 
	ON CONFLICT DO NOTHING`

	checkIfExistsQ = `SELECT EXISTS ( SELECT 1 FROM pg_catalog.pg_matviews 
	WHERE matviewname = 'rocketpool_rewards_summary')`

	refreshQ  = `REFRESH MATERIALIZED VIEW CONCURRENTLY rocketpool_rewards_summary`
	getTreesQ = `SELECT id, data FROM rocketpool_reward_tree`
)
