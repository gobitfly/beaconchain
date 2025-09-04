package app

import (
	"context"
	"errors"
	"fmt"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (service *ApiService) CreateAPIKey(ctx context.Context, in *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error) {
	user := auth.MustUserFromContext(ctx)

	// limit check
	maxKeys, err := service.limiter.GetMaxAPIKeys(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get max API keys: %v", err)
	}

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API keys: %v", err)
	}
	if len(keys) >= maxKeys {
		return nil, common.NewExternalError(codes.ResourceExhausted, fmt.Sprintf("maximum number of API keys (%d) reached", maxKeys))
	}

	// create API key
	key, rawKey, err := apikey.NewAPIKey(in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create API key: %v", err)
	}

	dbKey, err := service.apiKeyRepository.Create(ctx, user.ID, key)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			return nil, status.Errorf(codes.AlreadyExists, "API key with name '%s' already exists", in.Name)
		}
		return nil, status.Errorf(codes.Internal, "failed to create API key in database: %v", err)
	}

	return &model.CreateAPIKeyResponse{
		RawApiKey: rawKey.ToBase62(),
		ApiKey:    transformAPIKeyToModel(dbKey),
	}, nil
}

func (service *ApiService) DeleteAPIKey(ctx context.Context, in *model.DeleteAPIKeyRequest) (*model.DeleteAPIKeyResponse, error) {
	user := auth.MustUserFromContext(ctx)

	if err := service.apiKeyRepository.Delete(ctx, user.ID, in.Name); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "API key not found: %s", in.Name)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete API key: %v", err)
	}

	return &model.DeleteAPIKeyResponse{}, nil
}

func (service *ApiService) DisableAPIKey(ctx context.Context, in *model.DisableAPIKeyRequest) (*model.DisableAPIKeyResponse, error) {
	user := auth.MustUserFromContext(ctx)

	preconditionKey, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API key: %v", err)
	}

	if len(preconditionKey) == 0 {
		return nil, status.Errorf(codes.NotFound, "API key not found: %s", in.Name)
	}

	if preconditionKey[0].DisabledAt != nil {
		return &model.DisableAPIKeyResponse{
			ApiKey: transformAPIKeyToModel(preconditionKey[0]),
		}, nil
	}

	updatedKey, err := service.apiKeyRepository.Disable(ctx, user.ID, in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to disable API key: %v", err)
	}

	return &model.DisableAPIKeyResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) EnableAPIKey(ctx context.Context, in *model.EnableAPIKeyRequest) (*model.EnableAPIKeyResponse, error) {
	user := auth.MustUserFromContext(ctx)

	preconditionKey, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API key: %v", err)
	}

	if len(preconditionKey) == 0 {
		return nil, status.Errorf(codes.NotFound, "API key not found: %s", in.Name)
	}

	if preconditionKey[0].DisabledAt == nil {
		return &model.EnableAPIKeyResponse{
			ApiKey: transformAPIKeyToModel(preconditionKey[0]),
		}, nil
	}

	updatedKey, err := service.apiKeyRepository.Enable(ctx, user.ID, in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to enable API key: %v", err)
	}

	return &model.EnableAPIKeyResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) GetAPIKeys(ctx context.Context, in *model.GetAPIKeysRequest) (*model.GetAPIKeysResponse, error) {
	user := auth.MustUserFromContext(ctx)

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API keys: %v", err)
	}

	return &model.GetAPIKeysResponse{
		ApiKeys: transformAPIKeysToModel(keys),
	}, nil
}

func (service *ApiService) GetAPIKey(ctx context.Context, in *model.GetAPIKeyRequest) (*model.GetAPIKeyResponse, error) {
	user := auth.MustUserFromContext(ctx)

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API key: %v", err)
	}

	if len(keys) == 0 {
		return nil, status.Errorf(codes.NotFound, "API key not found: %s", in.Name)
	}

	return &model.GetAPIKeyResponse{
		ApiKey: transformAPIKeyToModel(keys[0]), // can only contain one key with that name
	}, nil
}

func transformAPIKeyToModel(key apikey.APIKey) *model.APIKey {
	return &model.APIKey{
		Name:       key.Name,
		ShortKey:   key.ShortKey,
		CreatedAt:  io.TimePtrToProtoTimestamp(key.CreatedAt),
		LastUsedAt: io.TimePtrToProtoTimestamp(key.LastUsedAt),
		DisabledAt: io.TimePtrToProtoTimestamp(key.DisabledAt),
	}
}

func transformAPIKeysToModel(keys []apikey.APIKey) []*model.APIKey {
	modelKeys := make([]*model.APIKey, 0, len(keys))
	for _, k := range keys {
		modelKeys = append(modelKeys, transformAPIKeyToModel(k))
	}
	return modelKeys
}
