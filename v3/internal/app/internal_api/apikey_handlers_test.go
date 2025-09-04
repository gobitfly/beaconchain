package app

import (
	"context"
	"errors"
	"testing"
	"time"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/apikeyrepo"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo/userrepo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestApiService_CreateAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        *model.CreateAPIKeyRequest
		setupMocks   func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectCode   codes.Code
		expectedName string
	}{
		{
			name:  "success",
			input: &model.CreateAPIKeyRequest{Name: "test-key"},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(testKey, nil)
			},
			expectedName: "test-key",
		},
		{
			name:  "duplicate",
			input: &model.CreateAPIKeyRequest{Name: "dupe"},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(apikey.APIKey{}, domain.ErrDuplicate)
			},
			expectErr:  true,
			expectCode: codes.AlreadyExists,
		},
		{
			name:  "max keys reached",
			input: &model.CreateAPIKeyRequest{Name: "max-key"},
			setupMocks: func(apiKeyRepo *apikeyrepo.MockRepository) {
				apiKeyRepo.On("GetAll", mock.Anything, uint64(1337), (*string)(nil)).
					Return([]apikey.APIKey{testKey}, nil)
				apiKeyRepo.On("Create", mock.Anything, uint64(1337), mock.AnythingOfType("apikey.APIKey")).
					Return(apikey.APIKey{}, errors.New("should not be called"))
			},
			expectErr:  true,
			expectCode: codes.ResourceExhausted,
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
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.ApiKey.Name)
			}
		})
	}
}

func TestApiService_DeleteAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		input      *model.DeleteAPIKeyRequest
		setupMock  func(*apikeyrepo.MockRepository)
		expectErr  bool
		expectCode codes.Code
	}{
		{
			name:  "success",
			input: &model.DeleteAPIKeyRequest{Name: "exists"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("Delete", mock.Anything, uint64(1337), "exists").Return(nil)
			},
		},
		{
			name:  "not found",
			input: &model.DeleteAPIKeyRequest{Name: "missing"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("Delete", mock.Anything, uint64(1337), "missing").Return(domain.ErrNotFound)
			},
			expectErr:  true,
			expectCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)

			context := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.DeleteAPIKey(context, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestApiService_DisableAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        *model.DisableAPIKeyRequest
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectCode   codes.Code
		expectedName string
	}{
		{
			name:  "success",
			input: &model.DisableAPIKeyRequest{Name: "ok"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("ok")).Return([]apikey.APIKey{testKey}, nil)
				repo.On("Disable", mock.Anything, uint64(1337), "ok").Return(disabledKey, nil)
			},
			expectedName: "disabled-key",
		},
		{
			name:  "not found",
			input: &model.DisableAPIKeyRequest{Name: "nf"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("nf")).Return([]apikey.APIKey{}, nil)
				repo.On("Disable", mock.Anything, uint64(1337), "nf").Return(apikey.APIKey{}, errors.New("should not be called"))
			},
			expectErr:  true,
			expectCode: codes.NotFound,
		},
		{
			name:  "disable disabled",
			input: &model.DisableAPIKeyRequest{Name: "disabled-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("disabled-key")).Return([]apikey.APIKey{disabledKey}, nil)
				repo.On("Disable", mock.Anything, uint64(1337), "disabled-key").Return([]apikey.APIKey{}, errors.New("should not be called"))
			},
			expectedName: "disabled-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			context := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.DisableAPIKey(context, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.ApiKey.Name)
			}
		})
	}
}

func TestApiService_EnableAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        *model.EnableAPIKeyRequest
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectCode   codes.Code
		expectedName string
	}{
		{
			name:  "success",
			input: &model.EnableAPIKeyRequest{Name: "enable"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("enable")).Return([]apikey.APIKey{disabledKey}, nil)
				repo.On("Enable", mock.Anything, uint64(1337), "enable").Return(testKey, nil)
			},
			expectedName: "test-key",
		},
		{
			name:  "not found",
			input: &model.EnableAPIKeyRequest{Name: "nope"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("nope")).Return([]apikey.APIKey{}, nil)
				repo.On("Enable", mock.Anything, uint64(1337), "nope").Return(apikey.APIKey{}, errors.New("should not be called"))
			},
			expectErr:  true,
			expectCode: codes.NotFound,
		},
		{
			name:  "enable enabled",
			input: &model.EnableAPIKeyRequest{Name: "test-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				repo.On("GetAll", mock.Anything, uint64(1337), getPtr("test-key")).Return([]apikey.APIKey{testKey}, nil)
				repo.On("Enable", mock.Anything, uint64(1337), "test-key").Return([]apikey.APIKey{}, errors.New("should not be called"))
			},
			expectedName: "test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			context := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.EnableAPIKey(context, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.ApiKey.Name)
			}
		})
	}
}

func TestApiService_GetAPIKeys(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectCode   codes.Code
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
			expectErr:  true,
			expectCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)
			context := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.GetAPIKeys(context, &model.GetAPIKeysRequest{})

			if tt.expectErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.ApiKeys[0].Name)
			}
		})
	}
}

func TestApiService_GetAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		input        *model.GetAPIKeyRequest
		setupMock    func(*apikeyrepo.MockRepository)
		expectErr    bool
		expectCode   codes.Code
		expectedName string
	}{
		{
			name:  "success",
			input: &model.GetAPIKeyRequest{Name: "filtered-key"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "filtered-key"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).
					Return([]apikey.APIKey{filteredKey}, nil)
			},
			expectedName: "filtered-key",
		},
		{
			name:  "not found",
			input: &model.GetAPIKeyRequest{Name: "not-found"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "not-found"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).Return([]apikey.APIKey{}, nil)
			},
			expectErr:  true,
			expectCode: codes.NotFound,
		},
		{
			name:  "db error",
			input: &model.GetAPIKeyRequest{Name: "db-error"},
			setupMock: func(repo *apikeyrepo.MockRepository) {
				name := "db-error"
				repo.On("GetAll", mock.Anything, uint64(1337), &name).Return([]apikey.APIKey{}, errors.New("db error"))
			},
			expectErr:  true,
			expectCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(apikeyrepo.MockRepository)
			tt.setupMock(mockRepo)
			svc := newService(mockRepo)

			context := auth.SetUserInContext(context.Background(), domain.User{ID: 1337})
			resp, err := svc.GetAPIKey(context, tt.input)

			if tt.expectErr {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.ApiKey.Name)
			}
		})
	}
}

func getPtr(s string) *string {
	return &s
}
