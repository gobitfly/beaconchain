package app

import (
	"context"
	"errors"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/app/io"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (service *ApiService) CreateAPIKey(ctx context.Context, in *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error) {
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	key, rawKey, err := apikey.NewAPIKey(in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create API key: %v", err)
	}

	dbKey, err := service.authRepository.CreateAPIKey(ctx, dummyUserId, key)
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
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	if err := service.authRepository.DeleteAPIKey(ctx, dummyUserId, in.Name); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "API key not found: %s", in.Name)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete API key: %v", err)
	}

	return &model.DeleteAPIKeyResponse{}, nil
}

func (service *ApiService) DisableAPIKey(ctx context.Context, in *model.DisableAPIKeyRequest) (*model.DisableAPIKeyResponse, error) {
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	preconditionKey, err := service.authRepository.GetAPIKeys(ctx, dummyUserId, &in.Name)
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

	updatedKey, err := service.authRepository.DisableAPIKey(ctx, dummyUserId, in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to disable API key: %v", err)
	}

	return &model.DisableAPIKeyResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) EnableAPIKey(ctx context.Context, in *model.EnableAPIKeyRequest) (*model.EnableAPIKeyResponse, error) {
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	preconditionKey, err := service.authRepository.GetAPIKeys(ctx, dummyUserId, &in.Name)
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

	updatedKey, err := service.authRepository.EnableAPIKey(ctx, dummyUserId, in.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to enable API key: %v", err)
	}

	return &model.EnableAPIKeyResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) GetAPIKeys(ctx context.Context, in *model.GetAPIKeysRequest) (*model.GetAPIKeysResponse, error) {
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	keys, err := service.authRepository.GetAPIKeys(ctx, dummyUserId, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get API keys: %v", err)
	}

	return &model.GetAPIKeysResponse{
		ApiKeys: transformAPIKeysToModel(keys),
	}, nil
}

func (service *ApiService) GetAPIKey(ctx context.Context, in *model.GetAPIKeyRequest) (*model.GetAPIKeyResponse, error) {
	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	keys, err := service.authRepository.GetAPIKeys(ctx, dummyUserId, &in.Name)
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
