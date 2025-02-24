package blobindexer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/blobstore"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/consapi/mocks"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	lru "github.com/hashicorp/golang-lru/v2"

	"go.uber.org/atomic"
)

var (
	bucket      = "bucket"
	contentType = "application/json"
	networkID   = "test"
)

// TestNewBlobIndexer tests the NewBlobIndexer function
// by creating a new BlobIndexer instance with different S3 configurations
// and checking for expected errors and non-nil BlobIndexer instances
func TestNewBlobIndexer(t *testing.T) {
	tests := []struct {
		name          string
		S3Config      *types.S3Config
		expectedError bool
	}{
		{
			name: "valid config",
			S3Config: &types.S3Config{
				AccessKeyId:     "test-access-key",
				AccessKeySecret: "test-key-secret",
				Endpoint:        "http://test-endpoint",
			},
			expectedError: false,
		},
		{
			name:          "empty S3 config",
			S3Config:      &types.S3Config{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Indexer: types.IndexerConfig{
					Node: types.NodeConfig{
						Host: "localhost",
						Port: "5052",
					},
				},
				BlobIndexer: types.BlobIndexerConfig{
					DisableStatusReports: true,
					S3:                   *tt.S3Config,
				},
			}
			blobIndexer, err := NewBlobIndexer()
			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
			if blobIndexer == nil {
				t.Fatalf("expected non-nil blobIndexer")
			}
		})
	}
}

// TestGetIndexerStatusFromS3 tests the getIndexerStatusFromS3 function
// by mocking S3 client responses for fetching the indexer status
// and checking the returned status and expected errors
func TestGetIndexerStatusFromS3(t *testing.T) {
	tests := []struct {
		name          string
		stubGetFunc   func(ctx context.Context, bucket string, key string) (blobstore.Blob, error)
		expectedSlot  uint64
		expectedError bool
	}{
		{
			name: "valid status",
			stubGetFunc: func(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
				return blobstore.Blob{
					Body: io.NopCloser(strings.NewReader(`{"last_indexed_finalized_slot":1}`)),
				}, nil
			},
			expectedSlot:  1,
			expectedError: false,
		},
		{
			name: "fetching error",
			stubGetFunc: func(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
				return blobstore.Blob{}, errors.New("error")
			},
			expectedSlot:  0,
			expectedError: true,
		},
		{
			name: "invalid JSON response",
			stubGetFunc: func(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
				return blobstore.Blob{
					Body: io.NopCloser(strings.NewReader(`invalid`)),
				}, nil
			},
			expectedSlot:  0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubBlobStore := &StubBlobStore{
				GetFunc: tt.stubGetFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore: stubBlobStore,
				networkID: networkID,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			status, err := blobIndexer.getIndexerStatusFromS3()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if status == nil {
					t.Error("expected non-nil status, got nil")
				}
				if status != nil {
					if status.LastIndexedFinalizedSlot != tt.expectedSlot {
						t.Errorf("expected LastIndexedFinalizedSlot to be %d, got %d", tt.expectedSlot, status.LastIndexedFinalizedSlot)
					}
				}
			}
		})
	}
}

// TestStoreIndexerStatusInS3 tests the storeIndexerStatusInS3 function
// by mocking S3 client responses for storing the indexer status
// and checking for expected errors
func TestStoreIndexerStatusInS3(t *testing.T) {
	tests := []struct {
		name          string
		status        BlobIndexerStatus
		stubPutFunc   func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error
		expectedError bool
	}{
		{
			name: "valid status upload",
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     100,
				LastIndexedFinalizedBlobSlot: 90,
				CurrentBlobIndexerId:         "1",
				LastUpdate:                   time.Now(),
				BlobIndexerVersion:           version.Version,
			},
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return nil
			},
			expectedError: false,
		},
		{
			name:   "empty blob status",
			status: BlobIndexerStatus{},
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return nil
			},
			expectedError: false,
		},
		{
			name:   "fetching error",
			status: BlobIndexerStatus{},
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return errors.New("error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubBlobStore := &StubBlobStore{
				PutFunc: tt.stubPutFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore: stubBlobStore,
				networkID: networkID,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			err := blobIndexer.storeIndexerStatusInS3(tt.status)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestGetBlobSidecarsAtSlot tests the getBlobSidecarsAtSlot function
// by mocking client responses for fetching blob sidecars
// and checking for expected errors
func TestGetBlobSidecarsAtSlot(t *testing.T) {
	tests := []struct {
		name             string
		slot             uint64
		status           BlobIndexerStatus
		mockResponse     *constypes.StandardBlobSidecarsResponse
		mockError        error
		expectedResponse *constypes.StandardBlobSidecarsResponse
		expectedError    bool
	}{
		{
			name: "valid blob side cars response",
			slot: 1,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 1,
			},
			mockResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 1,
					},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "error fetching blob side cars",
			slot: 1,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 1,
			},
			mockResponse:  nil,
			mockError:     errors.New("error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.Client)
			if tt.status.LastIndexedFinalizedBlobSlot > 0 {
				mockClient.On("GetBlobSidecars", tt.status.LastIndexedFinalizedBlobSlot).Return(tt.mockResponse, tt.mockError)
			}
			blobIndexer := &BlobIndexer{
				cl: mockClient,
			}

			response, err := blobIndexer.getBlobSidecarsAtSlot(tt.slot)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tt.mockError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if response != tt.mockResponse {
					t.Fatalf("expected response, got nil")
				}

				if response != tt.mockResponse {
					t.Errorf("expected response: %v, got: %v", tt.expectedResponse, response)
				}
			}
		})
	}
}

// TestPutObjectInS3 tests the putObjectInS3 function
// by mocking client responses for putting objects in S3
// and checking for expected errors
func TestPutObjectInS3(t *testing.T) {
	tests := []struct {
		name          string
		contentType   *string
		data          []byte
		metadata      map[string]string
		stubPutFunc   func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error
		expectedError bool
	}{
		{
			name:        "valid put object",
			contentType: &contentType,
			data:        []byte(`{"key": "value"}`),
			metadata:    map[string]string{"metadata": "data"},
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return nil
			},
			expectedError: false,
		},
		{
			name:        "error putting object",
			contentType: &contentType,
			data:        []byte(`{"key": "value"}`),
			metadata:    map[string]string{"metadata": "data"},
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return errors.New("error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubBlobStore := &StubBlobStore{
				PutFunc: tt.stubPutFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore: stubBlobStore,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			err := blobIndexer.putObjectInS3("key", "log-label", tt.contentType, tt.data, tt.metadata)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestGetObjectFromS3 tests the getObjectFromS3 function
// by mocking client responses for getting objects from S3
// and checking for expected errors
func TestGetObjectFromS3(t *testing.T) {
	tests := []struct {
		name          string
		stubGetFunc   func(ctx context.Context, bucket string, key string) (blobstore.Blob, error)
		expectedError bool
	}{
		{
			name: "successful get object",
			stubGetFunc: func(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
				return blobstore.Blob{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"key": "value"}`))),
				}, nil
			},
			expectedError: false,
		},
		{
			name: "error getting object",
			stubGetFunc: func(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
				return blobstore.Blob{}, errors.New("error")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubBlobStore := &StubBlobStore{
				GetFunc: tt.stubGetFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore: stubBlobStore,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			_, err := blobIndexer.getObjectFromS3("key", "log-label")

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestGetObjectMetadataFromS3 tests the getObjectMetadataFromS3 function
// by mocking client responses for getting object metadata from S3
// and checking for expected errors
func TestGetObjectMetadataFromS3(t *testing.T) {
	tests := []struct {
		name          string
		stubExistFunc func(ctx context.Context, bucket string, key string) (bool, error)
		expectedError bool
	}{
		{
			name: "successful get object metadata",
			stubExistFunc: func(ctx context.Context, bucket string, key string) (bool, error) {
				return true, nil
			},
			expectedError: false,
		},
		{
			name: "error getting object metadata",
			stubExistFunc: func(ctx context.Context, bucket string, key string) (bool, error) {
				return false, errors.New("error")
			},
			expectedError: true,
		},
		{
			name: "object does not exist",
			stubExistFunc: func(ctx context.Context, bucket string, key string) (bool, error) {
				return false, nil
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubBlobStore := &StubBlobStore{
				ExistFunc: tt.stubExistFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore: stubBlobStore,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			err := blobIndexer.checkIfObjectExistsInS3("key", "log-label")

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestFetchNodeData tests the fetchNodeData function
// by mocking client responses for fetching node data
// and checking for expected errors
func TestFetchNodeData(t *testing.T) {
	tests := []struct {
		name              string
		mockSpecResp      *constypes.StandardSpecResponse
		mockHeadResp      *constypes.StandardBeaconHeaderResponse
		mockFinalizedResp *constypes.StandardBeaconHeaderResponse
		mockError         error
		expectedError     bool
	}{
		{
			name:              "successful fetch",
			mockSpecResp:      &constypes.StandardSpecResponse{},
			mockHeadResp:      &constypes.StandardBeaconHeaderResponse{},
			mockFinalizedResp: &constypes.StandardBeaconHeaderResponse{},
			expectedError:     false,
		},
		{
			name:              "spec fetch error",
			mockSpecResp:      nil,
			mockHeadResp:      &constypes.StandardBeaconHeaderResponse{},
			mockFinalizedResp: &constypes.StandardBeaconHeaderResponse{},
			mockError:         errors.New("spec error"),
			expectedError:     true,
		},
		{
			name:              "head fetch error",
			mockSpecResp:      &constypes.StandardSpecResponse{},
			mockHeadResp:      nil,
			mockFinalizedResp: &constypes.StandardBeaconHeaderResponse{},
			mockError:         errors.New("head error"),
			expectedError:     true,
		},
		{
			name:              "finalized fetch error",
			mockSpecResp:      &constypes.StandardSpecResponse{},
			mockHeadResp:      &constypes.StandardBeaconHeaderResponse{},
			mockFinalizedResp: nil,
			mockError:         errors.New("finalized error"),
			expectedError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.Client)
			mockClient.On("GetSpec").Return(tt.mockSpecResp, tt.mockError)
			mockClient.On("GetBlockHeader", "head").Return(tt.mockHeadResp, tt.mockError)
			mockClient.On("GetBlockHeader", "finalized").Return(tt.mockFinalizedResp, tt.mockError)

			blobIndexer := &BlobIndexer{
				cl: mockClient,
			}

			headHeader, finalizedHeader, spec, err := blobIndexer.fetchNodeData(context.Background())

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.mockError.Error()) {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if headHeader == nil {
					t.Fatalf("expected non-nil head header")
				}
				if finalizedHeader == nil {
					t.Fatalf("expected non-nil finalized header")
				}
				if spec == nil {
					t.Fatalf("expected non-nil spec")
				}
			}
		})
	}
}

// TestFetchSpec tests the fetchSpec function
// by mocking client responses for fetching the spec
// and checking for expected errors
func TestFetchSpec(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *constypes.StandardSpecResponse
		mockError     error
		expectedError bool
		expectedSpec  *constypes.StandardSpecResponse
	}{
		{
			name: "valid fetch spec",
			mockResponse: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{},
			},
			mockError:     nil,
			expectedError: false,
			expectedSpec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{},
			},
		},
		{
			name:          "error fetching spec",
			mockResponse:  nil,
			mockError:     errors.New("spec not available"),
			expectedError: true,
			expectedSpec:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &constypes.StandardSpecResponse{}
			mockClient := new(mocks.Client)
			mockClient.On("GetSpec").Return(tt.mockResponse, tt.mockError)
			blobIndexer := &BlobIndexer{
				cl: mockClient,
			}

			err := blobIndexer.fetchSpec(spec)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.mockError.Error()) {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if tt.expectedSpec != nil {
					if *spec != *tt.expectedSpec {
						t.Errorf("expected spec %v, got %v", tt.expectedSpec, spec)
					}
				}
			}
		})
	}
}

// TestFetchBlockHeader tests the fetchBlockHeader function
// by mocking client responses for fetching block headers
// and checking for expected errors
func TestFetchBlockHeader(t *testing.T) {
	tests := []struct {
		name          string
		blockID       string
		mockResponse  *constypes.StandardBeaconHeaderResponse
		mockError     error
		expectedError bool
		expectedSlot  uint64
	}{
		{
			name:    "valid fetch for head block",
			blockID: "head",
			mockResponse: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 1,
						},
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedSlot:  1,
		},
		{
			name:    "valid fetch for finalized block",
			blockID: "finalized",
			mockResponse: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 2,
						},
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedSlot:  2,
		},
		{
			name:          "error fetching block header",
			blockID:       "missing",
			mockResponse:  nil,
			mockError:     errors.New("block not found"),
			expectedError: true,
			expectedSlot:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &constypes.StandardBeaconHeaderResponse{}
			mockClient := new(mocks.Client)
			mockClient.On("GetBlockHeader", tt.blockID).Return(tt.mockResponse, tt.mockError)
			blobIndexer := &BlobIndexer{
				cl: mockClient,
			}

			err := blobIndexer.fetchBlockHeader(tt.blockID, header)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.mockError.Error()) {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if tt.mockResponse != nil {
					if header.Data.Header.Message.Slot != tt.expectedSlot {
						t.Errorf("expected slot %d, got %d", tt.expectedSlot, header.Data.Header.Message.Slot)
					}
				}
			}
		})
	}
}

// TestValidateSpec tests the validateSpec function
// by validating the spec with different configurations
// and checking for expected errors
func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            *constypes.StandardSpecResponse
		configDepositID uint64
		mockError       error
		expectedError   bool
	}{
		{
			name: "valid spec",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch:                   uint64Ptr(4),
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					DepositNetworkID:                 1,
				},
			},
			configDepositID: 1,
			mockError:       nil,
			expectedError:   false,
		},
		{
			name: "missing DenebForkEpoch value",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch:                   nil,
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					DepositNetworkID:                 1,
				},
			},
			configDepositID: 1,
			mockError:       errors.New("DENEB_FORK_EPOCH not set in spec"),
			expectedError:   true,
		},
		{
			name: "missing MinEpochsForBlobSidecarsRequests value",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch:                   uint64Ptr(4),
					MinEpochsForBlobSidecarsRequests: nil,
					DepositNetworkID:                 1,
				},
			},
			configDepositID: 1,
			mockError:       errors.New("MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS not set in spec"),
			expectedError:   true,
		},
		{
			name: "DepositNetworkID mismatch",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch:                   uint64Ptr(4),
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					DepositNetworkID:                 1,
				},
			},
			configDepositID: 2,
			mockError:       errors.New("ClConfig.DepositNetworkID mismatch: 2 != 1"),
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						DepositNetworkID: tt.configDepositID,
					},
				},
			}

			err := validateSpec(tt.spec)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tt.mockError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestShouldSkipBlobIndexing tests the shouldSkipBlobIndexing function
// by checking if blob indexing should be skipped
// based on the current status and ID
func TestShouldSkipBlobIndexing(t *testing.T) {
	tests := []struct {
		name           string
		status         *BlobIndexerStatus
		id             string
		expectedResult bool
	}{
		{
			name: "CurrentBlobIndexerId and ID are the same",
			status: &BlobIndexerStatus{
				CurrentBlobIndexerId: "1",
				LastUpdate:           time.Now(),
			},
			id:             "1",
			expectedResult: false,
		},
		{
			name: "different CurrentBlobIndexerId and ID",
			status: &BlobIndexerStatus{
				CurrentBlobIndexerId: "2",
				LastUpdate:           time.Now(),
			},
			id:             "1",
			expectedResult: true,
		},
		{
			name: "valid LastUpdate time",
			status: &BlobIndexerStatus{
				CurrentBlobIndexerId: "1",
				LastUpdate:           time.Now().Add(-2 * waitForOtherBlobIndexerDuration),
			},
			id:             "1",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blobIndexer := &BlobIndexer{
				id: tt.id,
			}
			result := blobIndexer.shouldSkipBlobIndexing(tt.status)
			if result != tt.expectedResult {
				t.Errorf("got %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestCheckNodePruning tests the checkNodePruning function
// by checking node pruning based on the minimum blob slot
// and status and checking for expected errors
func TestCheckNodePruning(t *testing.T) {
	tests := []struct {
		name          string
		minBlobSlot   uint64
		status        BlobIndexerStatus
		mockResponse  *constypes.StandardBlobSidecarsResponse
		mockError     error
		expectedError bool
	}{
		{
			name:        "valid last finalized slot value",
			minBlobSlot: 10,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 0,
			},
			mockResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 1,
					},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:        "last finalized slot is smaller than minBlobSlot",
			minBlobSlot: 10,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     9,
				LastIndexedFinalizedBlobSlot: 0,
			},
			mockResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 1,
					},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:        "last finalized slot and blob slot is 0",
			minBlobSlot: 10,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     0,
				LastIndexedFinalizedBlobSlot: 0,
			},
			mockResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 1,
					},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:        "no blobs found at lastIndexedFinalizedBlobSlot",
			minBlobSlot: 10,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 2,
			},
			mockResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{},
			},
			mockError:     errors.New("no blobs found at lastIndexedFinalizedBlobSlot: 2, node has pruned too far?"),
			expectedError: true,
		},
		{
			name:        "error fetching blob sidecars",
			minBlobSlot: 10,
			status: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 2,
			},
			mockError:     errors.New("fetch error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.Client)
			if tt.status.LastIndexedFinalizedBlobSlot > 0 {
				mockClient.On("GetBlobSidecars", tt.status.LastIndexedFinalizedBlobSlot).Return(tt.mockResponse, tt.mockError)
			}
			blobIndexer := &BlobIndexer{
				cl: mockClient,
			}

			err := blobIndexer.checkNodePruning(tt.minBlobSlot, &tt.status)

			if tt.expectedError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tt.mockError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.mockError, err)
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestIndexBlobsInBatches tests the indexBlobsInBatches function
// by indexing blobs in batches and checking for expected errors
func TestIndexBlobsInBatches(t *testing.T) {
	tests := []struct {
		name                     string
		status                   *BlobIndexerStatus
		headHeader               *constypes.StandardBeaconHeaderResponse
		finalizedHeader          *constypes.StandardBeaconHeaderResponse
		slot                     uint64
		startSlot                uint64
		stubPutFunc              func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error
		mockBlobSidecarsResponse *constypes.StandardBlobSidecarsResponse
		expectedError            bool
	}{
		{
			name: "valid indexing",
			status: &BlobIndexerStatus{
				LastIndexedFinalizedSlot:     2,
				LastIndexedFinalizedBlobSlot: 2,
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 4,
						},
					},
				},
			},
			finalizedHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 4,
						},
					},
				},
			},
			slot:      4,
			startSlot: 4,
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return nil
			},
			mockBlobSidecarsResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 0,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "error updating indexer status",
			status: &BlobIndexerStatus{
				LastIndexedFinalizedSlot:     1,
				LastIndexedFinalizedBlobSlot: 1,
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 4,
						},
					},
				},
			},
			finalizedHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 4,
						},
					},
				},
			},
			slot:      4,
			startSlot: 4,
			stubPutFunc: func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
				return errors.New("error")
			},
			mockBlobSidecarsResponse: &constypes.StandardBlobSidecarsResponse{
				Data: []constypes.BlobSidecarsData{
					{
						Index: 0,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.Client)
			mockClient.On("GetBlobSidecars", tt.slot).Return(tt.mockBlobSidecarsResponse, nil)

			writtenBlobsCache, err := lru.New[string, bool](1000)
			if err != nil {
				t.Fatalf("error creating lru cache: %v", err)
			}

			stubBlobStore := &StubBlobStore{
				PutFunc: tt.stubPutFunc,
			}

			blobIndexer := &BlobIndexer{
				BlobStore:         stubBlobStore,
				cl:                mockClient,
				networkID:         networkID,
				writtenBlobsCache: writtenBlobsCache,
			}

			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					S3: types.S3Config{
						Bucket: bucket,
					},
				},
			}

			err = blobIndexer.indexBlobsInBatches(tt.status, tt.headHeader, tt.finalizedHeader, tt.startSlot)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
			if !tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestUpdateIndexerStatus tests the updateIndexerStatus function
// by updating the indexer status with new finalized slots
// and checking the updated status
func TestUpdateIndexerStatus(t *testing.T) {
	tests := []struct {
		name                     string
		batchEnd                 uint64
		finalizedHeader          *constypes.StandardBeaconHeaderResponse
		lastIndexedFinalizedSlot uint64
		status                   *BlobIndexerStatus
		expectedStatus           BlobIndexerStatus
	}{
		{
			name:     "update status with new finalized slot",
			batchEnd: 100,
			finalizedHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 100,
						},
					},
				},
			},
			lastIndexedFinalizedSlot: 99,
			status: &BlobIndexerStatus{
				LastIndexedFinalizedBlobSlot: 99,
			},
			expectedStatus: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     100,
				LastIndexedFinalizedBlobSlot: 99,
				CurrentBlobIndexerId:         "1",
				BlobIndexerVersion:           version.Version,
			},
		},
		{
			name:     "update status with higher LastIndexedFinalizedBlobSlot in status",
			batchEnd: 100,
			finalizedHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 100,
						},
					},
				},
			},
			lastIndexedFinalizedSlot: 98,
			status: &BlobIndexerStatus{
				LastIndexedFinalizedBlobSlot: 99,
			},
			expectedStatus: BlobIndexerStatus{
				LastIndexedFinalizedSlot:     100,
				LastIndexedFinalizedBlobSlot: 99,
				CurrentBlobIndexerId:         "1",
				BlobIndexerVersion:           version.Version,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blobIndexer := &BlobIndexer{
				id: "1",
			}

			lastIndexedFinalizedBlobSlot := atomic.NewUint64(tt.lastIndexedFinalizedSlot)
			updatedStatus := updateIndexerStatus(tt.batchEnd, tt.finalizedHeader, lastIndexedFinalizedBlobSlot, tt.status, blobIndexer)

			if updatedStatus.LastIndexedFinalizedSlot != tt.expectedStatus.LastIndexedFinalizedSlot {
				t.Errorf("expected LastIndexedFinalizedSlot %v, got %v", tt.expectedStatus.LastIndexedFinalizedSlot, updatedStatus.LastIndexedFinalizedSlot)
			}
			if updatedStatus.LastIndexedFinalizedBlobSlot != tt.expectedStatus.LastIndexedFinalizedBlobSlot {
				t.Errorf("expected LastIndexedFinalizedBlobSlot %v, got %v", tt.expectedStatus.LastIndexedFinalizedBlobSlot, updatedStatus.LastIndexedFinalizedBlobSlot)
			}
			if updatedStatus.BlobIndexerVersion != tt.expectedStatus.BlobIndexerVersion {
				t.Errorf("expected BlobIndexerVersion %v, got %v", tt.expectedStatus.BlobIndexerVersion, updatedStatus.BlobIndexerVersion)
			}
		})
	}
}

// TestCalculateBatchEnd tests the calculateBatchEnd function
// by calculating the batch end based on the batch start,
// batch size and head slot
func TestCalculateBatchEnd(t *testing.T) {
	tests := []struct {
		name           string
		batchStart     uint64
		batchSize      uint64
		headSlot       uint64
		expectedResult uint64
	}{
		{
			name:           "batch end within head slot",
			batchStart:     100,
			batchSize:      50,
			headSlot:       200,
			expectedResult: 150,
		},
		{
			name:           "batch end exceeds head slot",
			batchStart:     100,
			batchSize:      150,
			headSlot:       200,
			expectedResult: 200,
		},
		{
			name:           "batch start equals head slot",
			batchStart:     200,
			batchSize:      50,
			headSlot:       200,
			expectedResult: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBatchEnd(tt.batchStart, tt.batchSize, tt.headSlot)
			if result != tt.expectedResult {
				t.Errorf("expected batch end %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// TestCalculateMinBlobSlot tests the calculateMinBlobSlot function
// by calculating the minimum blob slot based on the spec,
// head header and prune margin epochs
func TestCalculateMinBlobSlot(t *testing.T) {
	tests := []struct {
		name                string
		spec                *constypes.StandardSpecResponse
		headHeader          *constypes.StandardBeaconHeaderResponse
		pruneMarginEpochs   uint64
		expectedMinBlobSlot uint64
	}{
		{
			name: "normal case: head header slot greater than min blob slot range",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					SlotsPerEpoch:                    32,
				},
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          100,
							ProposerIndex: 0,
							ParentRoot:    hexutil.Bytes{},
							StateRoot:     hexutil.Bytes{},
							BodyRoot:      hexutil.Bytes{},
						},
						Signature: hexutil.Bytes{},
					},
				},
			},
			pruneMarginEpochs:   1,
			expectedMinBlobSlot: 4,
		},
		{
			name: "head header slot equal to min blob slot range",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					SlotsPerEpoch:                    32,
				},
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          64,
							ProposerIndex: 0,
							ParentRoot:    hexutil.Bytes{},
							StateRoot:     hexutil.Bytes{},
							BodyRoot:      hexutil.Bytes{},
						},
						Signature: hexutil.Bytes{},
					},
				},
			},
			pruneMarginEpochs:   1,
			expectedMinBlobSlot: 0,
		},
		{
			name: "head header slot less than min blob slot range",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					SlotsPerEpoch:                    32,
				},
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          50,
							ProposerIndex: 0,
							ParentRoot:    hexutil.Bytes{},
							StateRoot:     hexutil.Bytes{},
							BodyRoot:      hexutil.Bytes{},
						},
						Signature: hexutil.Bytes{},
					},
				},
			},
			pruneMarginEpochs:   1,
			expectedMinBlobSlot: 0,
		},
		{
			name: "prune margin slot range greater than calculated min blob slot",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					SlotsPerEpoch:                    32,
				},
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          100,
							ProposerIndex: 0,
							ParentRoot:    hexutil.Bytes{},
							StateRoot:     hexutil.Bytes{},
							BodyRoot:      hexutil.Bytes{},
						},
						Signature: hexutil.Bytes{},
					},
				},
			},
			pruneMarginEpochs:   4,
			expectedMinBlobSlot: 36,
		},
		{
			name: "prune margin slot range equal to calculated min blob slot",
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					MinEpochsForBlobSidecarsRequests: uint64Ptr(2),
					SlotsPerEpoch:                    32,
				},
			},
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          128,
							ProposerIndex: 0,
							ParentRoot:    hexutil.Bytes{},
							StateRoot:     hexutil.Bytes{},
							BodyRoot:      hexutil.Bytes{},
						},
						Signature: hexutil.Bytes{},
					},
				},
			},
			pruneMarginEpochs:   2,
			expectedMinBlobSlot: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				BlobIndexer: types.BlobIndexerConfig{
					PruneMarginEpochs: tt.pruneMarginEpochs,
				},
			}

			minBlobSlot := calculateMinBlobSlot(tt.spec, tt.headHeader)
			if minBlobSlot != tt.expectedMinBlobSlot {
				t.Errorf("got minBlobSlot %v, want %v", minBlobSlot, tt.expectedMinBlobSlot)
			}
		})
	}
}

// TestCalculateStartSlot tests the calculateStartSlot function
// by calculating the start slot based on the status and spec
func TestCalculateStartSlot(t *testing.T) {
	tests := []struct {
		name              string
		status            *BlobIndexerStatus
		spec              *constypes.StandardSpecResponse
		expectedStartSlot uint64
		expectedDenebSlot uint64
	}{
		{
			name: "last indexed slot before deneb fork",
			status: &BlobIndexerStatus{
				LastIndexedFinalizedSlot: 100,
			},
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch: uint64Ptr(4),
					SlotsPerEpoch:  32,
				},
			},
			expectedStartSlot: 128,
			expectedDenebSlot: 128,
		},
		{
			name: "last indexed slot after deneb fork",
			status: &BlobIndexerStatus{
				LastIndexedFinalizedSlot: 150,
			},
			spec: &constypes.StandardSpecResponse{
				Data: constypes.StandardSpec{
					DenebForkEpoch: uint64Ptr(4),
					SlotsPerEpoch:  32,
				},
			},
			expectedStartSlot: 151,
			expectedDenebSlot: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startSlot, denebSlot := calculateStartSlot(tt.status, tt.spec)
			if startSlot != tt.expectedStartSlot {
				t.Errorf("got startSlot %v, want %v", startSlot, tt.expectedStartSlot)
			}
			if denebSlot != tt.expectedDenebSlot {
				t.Errorf("got denebSlot %v, want %v", denebSlot, tt.expectedDenebSlot)
			}
		})
	}
}

// TestCalculateEstimatedTimeToHeadDuration tests the
// calculateEstimatedTimeToHeadDuration function
// by calculating the estimated time to head duration
// based on the estimated time to head
func TestCalculateEstimatedTimeToHeadDuration(t *testing.T) {
	tests := []struct {
		name                string
		estimatedTimeToHead float64
		expectedDuration    time.Duration
	}{
		{
			name:                "120 seconds",
			estimatedTimeToHead: 120.0,
			expectedDuration:    120 * time.Second,
		},
		{
			name:                "0 seconds",
			estimatedTimeToHead: 0.0,
			expectedDuration:    0 * time.Second,
		},
		{
			name:                "invalid input",
			estimatedTimeToHead: -50.0,
			expectedDuration:    -50 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := calculateEstimatedTimeToHeadDuration(tt.estimatedTimeToHead)
			if duration != tt.expectedDuration {
				t.Errorf("got duration %v, want %v", duration, tt.expectedDuration)
			}
		})
	}
}

// TestCalculateEstimatedTimeToHead tests the calculateEstimatedTimeToHead function
// by calculating the estimated time to head based on the head header,
// batch start and slots per second
func TestCalculateEstimatedTimeToHead(t *testing.T) {
	tests := []struct {
		name               string
		headHeader         *constypes.StandardBeaconHeaderResponse
		batchStart         uint64
		slotsPerSecond     float64
		expectedTimeToHead float64
	}{
		{
			name: "head slot is grater than batch start",
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          200,
							ProposerIndex: 0,
						},
					},
				},
			},
			batchStart:         100,
			slotsPerSecond:     2.0,
			expectedTimeToHead: 50.0,
		},
		{
			name: "head slot equals batch start",
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          100,
							ProposerIndex: 0,
						},
					},
				},
			},
			batchStart:         100,
			slotsPerSecond:     2.0,
			expectedTimeToHead: 0.0,
		},
		{
			name: "batch start is grater head slot",
			headHeader: &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Root: hexutil.Bytes{},
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot:          50,
							ProposerIndex: 0,
						},
					},
				},
			},
			batchStart:         100,
			slotsPerSecond:     2.0,
			expectedTimeToHead: -25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimatedTimeToHead := calculateEstimatedTimeToHead(tt.headHeader, tt.batchStart, tt.slotsPerSecond)
			if estimatedTimeToHead != tt.expectedTimeToHead {
				t.Errorf("got estimatedTimeToHead %v, want %v", estimatedTimeToHead, tt.expectedTimeToHead)
			}
		})
	}
}

// TestCalculateBlobsPerSecond tests the calculateBlobsPerSecond function
// by calculating the blobs per second  based on the batch blobs indexed
// and batch start timestamp
func TestCalculateBlobsPerSecond(t *testing.T) {
	tests := []struct {
		name                string
		batchBlobsIndexed   int64
		batchStartTs        time.Time
		expectedBlobsPerSec float64
	}{
		{
			name:                "100 blobs indexed in 10 seconds",
			batchBlobsIndexed:   100,
			batchStartTs:        time.Now().Add(-10 * time.Second),
			expectedBlobsPerSec: 10.0,
		},
		{
			name:                "0 blobs indexed",
			batchBlobsIndexed:   0,
			batchStartTs:        time.Now().Add(-10 * time.Second),
			expectedBlobsPerSec: 0.0,
		},
		{
			name:                "invalid input",
			batchBlobsIndexed:   100,
			batchStartTs:        time.Now().Add(10 * time.Second),
			expectedBlobsPerSec: -10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchBlobsIndexed := atomic.Int64{}
			batchBlobsIndexed.Store(tt.batchBlobsIndexed)

			blobsPerSecond := calculateBlobsPerSecond(&batchBlobsIndexed, tt.batchStartTs)
			if math.Round(blobsPerSecond) != tt.expectedBlobsPerSec {
				t.Errorf("got blobsPerSecond %v, want %v", blobsPerSecond, tt.expectedBlobsPerSec)
			}
		})
	}
}

// TestCalculateSlotsPerSecond tests the calculateSlotsPerSecond function
// by calculating the slots per second based on the batch end,
// batch start and batch start timestamp
func TestCalculateSlotsPerSecond(t *testing.T) {
	tests := []struct {
		name                string
		batchEnd            uint64
		batchStart          uint64
		batchStartTs        time.Time
		expectedSlotsPerSec float64
	}{
		{
			name:                "100 slots in 10 seconds",
			batchEnd:            200,
			batchStart:          100,
			batchStartTs:        time.Now().Add(-10 * time.Second),
			expectedSlotsPerSec: 10.0,
		},
		{
			name:                "0 slots processed",
			batchEnd:            100,
			batchStart:          100,
			batchStartTs:        time.Now().Add(-10 * time.Second),
			expectedSlotsPerSec: 0.0,
		},
		{
			name:                "invalid input",
			batchEnd:            200,
			batchStart:          100,
			batchStartTs:        time.Now().Add(10 * time.Second),
			expectedSlotsPerSec: -10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slotsPerSecond := calculateSlotsPerSecond(tt.batchEnd, tt.batchStart, tt.batchStartTs)
			if math.Round(slotsPerSecond) != tt.expectedSlotsPerSec {
				t.Errorf("got slotsPerSecond %v, want %v", slotsPerSecond, tt.expectedSlotsPerSec)
			}
		})
	}
}

// TestGetLastIndexedFinalizedSlot tests the getLastIndexedFinalizedSlot function
// by getting the last indexed finalized slot based on the batch end
// and finalized header
func TestGetLastIndexedFinalizedSlot(t *testing.T) {
	tests := []struct {
		name                    string
		batchEnd                uint64
		expectedLastIndexedSlot uint64
	}{

		{
			name:                    "batchEnd before finalized slot",
			batchEnd:                100,
			expectedLastIndexedSlot: 100,
		},
		{
			name:                    "batchEnd equal to finalized slot",
			batchEnd:                150,
			expectedLastIndexedSlot: 150,
		},
		{
			name:                    "batchEnd after finalized slot",
			batchEnd:                200,
			expectedLastIndexedSlot: 150,
		},
		{
			name:                    "batchEnd is 0",
			batchEnd:                0,
			expectedLastIndexedSlot: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalizedHeader := &constypes.StandardBeaconHeaderResponse{
				Data: constypes.BeaconHeaderData{
					Header: constypes.BeaconHeaderMessage{
						Message: constypes.BeaconHeaderMessageData{
							Slot: 150,
						},
					},
				},
			}

			lastIndexedSlot := getLastIndexedFinalizedSlot(tt.batchEnd, finalizedHeader)
			if lastIndexedSlot != tt.expectedLastIndexedSlot {
				t.Errorf("got lastIndexedSlot %v, want %v", lastIndexedSlot, tt.expectedLastIndexedSlot)
			}
		})
	}
}

// unit64Ptr returns a pointer to a uint64
func uint64Ptr(i uint64) *uint64 {
	return &i
}

type StubBlobStore struct {
	PutFunc   func(ctx context.Context, bucket string, key string, blob blobstore.Blob) error
	ExistFunc func(ctx context.Context, bucket string, key string) (bool, error)
	GetFunc   func(ctx context.Context, bucket string, key string) (blobstore.Blob, error)
}

func (s *StubBlobStore) Put(ctx context.Context, bucket string, key string, blob blobstore.Blob) error {
	return s.PutFunc(ctx, bucket, key, blob)
}

func (s *StubBlobStore) Exist(ctx context.Context, bucket string, key string) (bool, error) {
	return s.ExistFunc(ctx, bucket, key)
}

func (s *StubBlobStore) Get(ctx context.Context, bucket string, key string) (blobstore.Blob, error) {
	return s.GetFunc(ctx, bucket, key)
}
