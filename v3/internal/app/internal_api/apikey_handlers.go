package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	model "github.com/gobitfly/beaconchain-backend/api/inhouse/model"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/common/islices"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/oapi-codegen/nullable"
)

func (service *ApiService) CreateAPIKey(ctx context.Context, in model.CreateAPIKeyRequestObject) (model.CreateAPIKeyResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	// limit check
	maxKeys, err := service.limiter.GetMaxAPIKeys(ctx, user)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get max API key limit"))
	}

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, nil)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get API keys"))
	}
	if len(keys) >= maxKeys {
		return nil, common.NewAPIUserFacingError(http.StatusForbidden, fmt.Sprintf("maximum number of API keys (%d) reached", maxKeys))
	}

	// create API key
	key, rawKey, err := apikey.NewAPIKey(in.Body.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create API key"))
	}

	dbKey, err := service.apiKeyRepository.Create(ctx, user.ID, key)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			return nil, common.NewAPIUserFacingError(http.StatusConflict, fmt.Sprintf("API key with name '%s' already exists", in.Body.Name))
		}
		return nil, errors.Join(err, errors.New("failed to store API key"))
	}

	return model.CreateAPIKey200JSONResponse{
		RawApiKey: rawKey.ToBase62(),
		ApiKey:    transformAPIKeyToModel(dbKey),
	}, nil
}

func (service *ApiService) DeleteAPIKey(ctx context.Context, in model.DeleteAPIKeyRequestObject) (model.DeleteAPIKeyResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	if err := service.apiKeyRepository.Delete(ctx, user.ID, in.Name); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, common.NewAPIUserFacingError(http.StatusNotFound, fmt.Sprintf("API key not found: %s", in.Name))
		}
		return nil, errors.Join(err, errors.New("failed to delete API key"))
	}

	return model.DeleteAPIKey204Response{}, nil
}

func (service *ApiService) DisableAPIKey(ctx context.Context, in model.DisableAPIKeyRequestObject) (model.DisableAPIKeyResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	preconditionKey, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get API key"))
	}

	if len(preconditionKey) == 0 {
		return nil, common.NewAPIUserFacingError(http.StatusNotFound, fmt.Sprintf("API key not found: %s", in.Name))
	}

	if preconditionKey[0].DisabledAt != nil {
		return model.DisableAPIKey200JSONResponse{
			ApiKey: transformAPIKeyToModel(preconditionKey[0]),
		}, nil
	}

	updatedKey, err := service.apiKeyRepository.Disable(ctx, user.ID, in.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to disable API key"))
	}

	return model.DisableAPIKey200JSONResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) EnableAPIKey(ctx context.Context, in model.EnableAPIKeyRequestObject) (model.EnableAPIKeyResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	preconditionKey, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get API key"))
	}

	if len(preconditionKey) == 0 {
		return nil, common.NewAPIUserFacingError(http.StatusNotFound, fmt.Sprintf("API key not found: %s", in.Name))
	}

	if preconditionKey[0].DisabledAt == nil {
		return model.EnableAPIKey200JSONResponse{
			ApiKey: transformAPIKeyToModel(preconditionKey[0]),
		}, nil
	}

	updatedKey, err := service.apiKeyRepository.Enable(ctx, user.ID, in.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to enable API key"))
	}

	return model.EnableAPIKey200JSONResponse{
		ApiKey: transformAPIKeyToModel(updatedKey),
	}, nil
}

func (service *ApiService) GetAPIKeys(ctx context.Context, in model.GetAPIKeysRequestObject) (model.GetAPIKeysResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, nil)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get API keys"))
	}

	return model.GetAPIKeys200JSONResponse{
		ApiKeys: islices.Transform(keys, transformAPIKeyToModel),
	}, nil
}

func (service *ApiService) GetAPIKey(ctx context.Context, in model.GetAPIKeyRequestObject) (model.GetAPIKeyResponseObject, error) {
	user := auth.MustUserFromContext(ctx)

	keys, err := service.apiKeyRepository.GetAll(ctx, user.ID, &in.Name)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to get API key"))
	}

	if len(keys) == 0 {
		return nil, common.NewAPIUserFacingError(http.StatusNotFound, fmt.Sprintf("API key not found: %s", in.Name))
	}

	return model.GetAPIKey200JSONResponse{
		ApiKey: transformAPIKeyToModel(keys[0]), // can only contain one key with that name
	}, nil
}

func transformAPIKeyToModel(key apikey.APIKey) model.APIKey {
	return model.APIKey{
		Name:      key.Name,
		ShortKey:  key.ShortKey,
		CreatedAt: *key.CreatedAt,
		LastUsedAt: func() nullable.Nullable[time.Time] {
			if key.LastUsedAt != nil {
				return nullable.NewNullableWithValue(*key.LastUsedAt)
			}
			return nullable.NewNullNullable[time.Time]()
		}(),
		DisabledAt: func() nullable.Nullable[time.Time] {
			if key.DisabledAt != nil {
				return nullable.NewNullableWithValue(*key.DisabledAt)
			}
			return nullable.NewNullNullable[time.Time]()
		}(),
	}
}
