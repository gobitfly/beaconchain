package app

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	model "github.com/gobitfly/beaconchain-backend/api/inhouse/model"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	now      = time.Now()
	testUUID = uuid.New()
	testKey  = apikey.APIKey{
		ID:        &testUUID,
		Name:      "test-key",
		ShortKey:  "abc...xyz",
		CreatedAt: &now,
	}
	filteredKey = apikey.APIKey{
		Name:      "filtered-key",
		ShortKey:  "filter...xyz",
		CreatedAt: &now,
	}
	disabledKey = apikey.APIKey{
		Name:       "disabled-key",
		ShortKey:   "disabled...xyz",
		CreatedAt:  &now,
		DisabledAt: &now,
	}
)

func newService(apiKeyRepo *apikeyrepo.MockRepository) *ApiService {
	return &ApiService{
		userRepository:   &userrepo.MockRepository{},
		apiKeyRepository: apiKeyRepo,
		limiter:          limits.NewLimiter(),
	}
}

func getPtr(s string) *string {
	return &s
}

func checkHTTPError(t *testing.T, err error, expectedStatus int) {
	require.Error(t, err)
	code := common.Code(err)
	assert.Equal(t, expectedStatus, code)
}

// --- TestApiService_CreateAPIKey ---
func TestApiService_CreateAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        model.CreateAPIKeyRequestObject
		setupMocks   func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
		expectedName string
	}{
		{
			name: "success",
			input: model.CreateAPIKeyRequestObject{
				Body: &model.CreateAPIKeyJSONRequestBody{Name: "test-key"},
			},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(testKey, nil)
			},
			expectedName: "test-key",
		},
		{
			name: "duplicate",
			input: model.CreateAPIKeyRequestObject{
				Body: &model.CreateAPIKeyJSONRequestBody{Name: "dupe"},
			},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(apikey.APIKey{}, domain.ErrDuplicate)
			},
			expectErr:    true,
			expectedCode: http.StatusConflict,
		},
		{
			name: "max keys reached",
			input: model.CreateAPIKeyRequestObject{
				Body: &model.CreateAPIKeyJSONRequestBody{Name: "max-key"},
			},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{testKey}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(apikey.APIKey{}, errors.New("should not be called"))
			},
			expectErr:    true,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKeyRepo := new(apikeyrepo.MockRepository)

			tt.setupMocks(apiKeyRepo)
			svc := newService(apiKeyRepo)

			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337, SubscriptionTier: domain.TierFree})
			resp, err := svc.CreateAPIKey(ctx, tt.input)

			if tt.expectErr {
				checkHTTPError(t, err, tt.expectedCode)
			} else {
				require.NoError(t, err)
				concreteResp, ok := resp.(model.CreateAPIKey200JSONResponse)
				require.True(t, ok, "Response should be CreateAPIKey200JSONResponse")
				assert.Equal(t, tt.expectedName, concreteResp.ApiKey.Name)
			}
		})
	}
}

// --- TestApiService_DeleteAPIKey ---
func TestApiService_DeleteAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        model.DeleteAPIKeyRequestObject
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
	}{
		{
			name:  "success",
			input: model.DeleteAPIKeyRequestObject{Name: "exists"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("Delete", mock.Anything, uint64(1337), "exists").Return(nil)
			},
			expectedCode: http.StatusNoContent, // 204 response has no body
		},
		{
			name:  "not found",
			input: model.DeleteAPIKeyRequestObject{Name: "missing"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("Delete", mock.Anything, uint64(1337), "missing").Return(domain.ErrNotFound)
			},
			expectErr:    true,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)

			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.DeleteAPIKey(ctx, tt.input)

			if tt.expectErr {
				checkHTTPError(t, err, tt.expectedCode)
			} else {
				require.NoError(t, err)
				_, ok := resp.(model.DeleteAPIKey204Response)
				assert.True(t, ok, "Response should be DeleteAPIKey204Response")
			}
		})
	}
}

// --- TestApiService_DisableAPIKey ---
func TestApiService_DisableAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        model.DisableAPIKeyRequestObject
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
		expectedName string
	}{
		{
			name:  "success",
			input: model.DisableAPIKeyRequestObject{Name: "ok"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("ok")).Return([]apikey.APIKey{testKey}, nil)
				repo.On("Disable", mock.Anything, uint64(1337), "ok").Return(disabledKey, nil)
			},
			expectedName: "disabled-key",
		},
		{
			name:  "not found",
			input: model.DisableAPIKeyRequestObject{Name: "nf"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("nf")).Return([]apikey.APIKey{}, nil)
				repo.On("Disable", mock.Anything, uint64(1337), "nf").Return(apikey.APIKey{}, errors.New("should not be called"))
			},
			expectErr:    true,
			expectedCode: http.StatusNotFound,
		},
		{
			name:  "disable disabled (idempotent)",
			input: model.DisableAPIKeyRequestObject{Name: "disabled-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("disabled-key")).Return([]apikey.APIKey{disabledKey}, nil)
				// Disable should not be called
			},
			expectedName: "disabled-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.DisableAPIKey(ctx, tt.input)

			if tt.expectErr {
				checkHTTPError(t, err, tt.expectedCode)
			} else {
				require.NoError(t, err)
				concreteResp, ok := resp.(model.DisableAPIKey200JSONResponse)
				require.True(t, ok, "Response should be DisableAPIKey200JSONResponse")
				assert.Equal(t, tt.expectedName, concreteResp.ApiKey.Name)
			}
		})
	}
}

// --- TestApiService_EnableAPIKey ---
func TestApiService_EnableAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        model.EnableAPIKeyRequestObject // Changed type
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
		expectedName string
	}{
		{
			name:  "success",
			input: model.EnableAPIKeyRequestObject{Name: "enable"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("enable")).Return([]apikey.APIKey{disabledKey}, nil)
				repo.On("Enable", mock.Anything, uint64(1337), "enable").Return(testKey, nil)
			},
			expectedName: "test-key",
		},
		{
			name:  "not found",
			input: model.EnableAPIKeyRequestObject{Name: "nope"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("nope")).Return([]apikey.APIKey{}, nil)
				// Enable should not be called
			},
			expectErr:    true,
			expectedCode: http.StatusNotFound,
		},
		{
			name:  "enable enabled (idempotent)",
			input: model.EnableAPIKeyRequestObject{Name: "test-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("test-key")).Return([]apikey.APIKey{testKey}, nil)
				// Enable should not be called
			},
			expectedName: "test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.EnableAPIKey(ctx, tt.input)

			if tt.expectErr {
				checkHTTPError(t, err, tt.expectedCode)
			} else {
				require.NoError(t, err)
				concreteResp, ok := resp.(model.EnableAPIKey200JSONResponse)
				require.True(t, ok, "Response should be EnableAPIKey200JSONResponse")
				assert.Equal(t, tt.expectedName, concreteResp.ApiKey.Name)
			}
		})
	}
}

// --- TestApiService_GetAPIKeys ---
func TestApiService_GetAPIKeys(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
		expectedName string
	}{
		{
			name: "success",
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).Return([]apikey.APIKey{testKey}, nil)
			},
			expectedName: "test-key",
		},
		{
			name: "db error",
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).Return([]apikey.APIKey{}, errors.New("fail"))
			},
			expectErr:    true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.GetAPIKeys(ctx, model.GetAPIKeysRequestObject{})

			if tt.expectErr {
				require.Error(t, err)
				checkHTTPError(t, err, tt.expectedCode)
			} else {
				require.NoError(t, err)
				concreteResp, ok := resp.(model.GetAPIKeys200JSONResponse)
				require.True(t, ok, "Response should be GetAPIKeys200JSONResponse")
				assert.Equal(t, tt.expectedName, concreteResp.ApiKeys[0].Name)
			}
		})
	}
}

// --- TestApiService_GetAPIKey ---
func TestApiService_GetAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        model.GetAPIKeyRequestObject
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectedCode int
		expectedName string
	}{
		{
			name:  "success",
			input: model.GetAPIKeyRequestObject{Name: "filtered-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "filtered-key"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).
					Return([]apikey.APIKey{filteredKey}, nil)
			},
			expectedName: "filtered-key",
		},
		{
			name:  "not found",
			input: model.GetAPIKeyRequestObject{Name: "not-found"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "not-found"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).Return([]apikey.APIKey{}, nil)
			},
			expectErr:    true,
			expectedCode: http.StatusNotFound,
		},
		{
			name:  "db error",
			input: model.GetAPIKeyRequestObject{Name: "db-error"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "db-error"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).Return([]apikey.APIKey{}, errors.New("db error"))
			},
			expectErr:    true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)

			ctx := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.GetAPIKey(ctx, tt.input)

			if tt.expectErr {
				if tt.expectedCode == http.StatusInternalServerError {
					require.Error(t, err)
					checkHTTPError(t, err, tt.expectedCode)
				} else {
					checkHTTPError(t, err, tt.expectedCode)
				}
			} else {
				require.NoError(t, err)
				concreteResp, ok := resp.(model.GetAPIKey200JSONResponse)
				require.True(t, ok, "Response should be GetAPIKey200JSONResponse")
				assert.Equal(t, tt.expectedName, concreteResp.ApiKey.Name)
			}
		})
	}
}
