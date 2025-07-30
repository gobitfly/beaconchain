package dataaccess

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/pkg/errors"
)

type DBAuthRepository struct {
	roConnectionAdminDb data_sources.AdminRoConnection
	rwConnectionAdminDb data_sources.AdminRwConnection
}

func (r *DBAuthRepository) Initialize(roConnectionAdminDb data_sources.AdminRoConnection, rwConnectionAdminDb data_sources.AdminRwConnection) {
	r.roConnectionAdminDb = roConnectionAdminDb
	r.rwConnectionAdminDb = rwConnectionAdminDb
}

func (r *DBAuthRepository) CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error) {
	ds := goqu.Dialect("postgres").Insert("api_keys_v2").
		Cols("user_id", "api_key", "short_key", "name").
		Vals(goqu.Vals{userID, key.Value, key.ShortKey, key.Name}).
		Returning("api_key_id", "api_key", "short_key", "name", "created_at")

	createdKey, err := runQuery[apikey.APIKey](ctx, r.rwConnectionAdminDb, ds)
	if err != nil {
		return apikey.APIKey{}, errors.Wrap(err, "failed to create user API key")
	}
	return createdKey, nil
}

func (r *DBAuthRepository) DeleteAPIKey(ctx context.Context, userID uint64, name string) error {
	ds := goqu.Dialect("postgres").Update("api_keys_v2").
		Set(goqu.Record{"deleted_at": goqu.L("NOW()")}).
		Where(
			goqu.Ex{
				"user_id":    userID,
				"name":       name,
				"deleted_at": nil,
			},
		)
	if err := execAndCheckRows(ctx, r.rwConnectionAdminDb, ds); err != nil {
		return errors.Wrap(err, "failed to delete user API key")
	}
	return nil
}

func (r *DBAuthRepository) DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	ds := goqu.Dialect("postgres").Update("api_keys_v2").
		Set(goqu.Record{
			"disabled_at": goqu.L("NOW()"),
		}).
		Where(
			goqu.And(
				goqu.C("user_id").Eq(userID),
				goqu.C("name").Eq(name),
				goqu.C("deleted_at").IsNull(),
				goqu.C("disabled_at").IsNull(),
			),
		).
		Returning(
			"api_key_id", "api_key", "short_key", "name",
			"created_at", "last_used_at", "disabled_at",
		)

	disabledKey, err := runQuery[apikey.APIKey](ctx, r.rwConnectionAdminDb, ds)
	if err != nil {
		return apikey.APIKey{}, errors.Wrap(err, "failed to disable user API key")
	}
	return disabledKey, nil
}

func (r *DBAuthRepository) EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	ds := goqu.Dialect("postgres").Update("api_keys_v2").
		Prepared(true).
		Set(goqu.Record{"disabled_at": nil}).
		Where(
			goqu.And(
				goqu.C("user_id").Eq(userID),
				goqu.C("name").Eq(name),
				goqu.C("deleted_at").IsNull(),
				goqu.C("disabled_at").IsNotNull(),
			),
		).
		Returning("api_key_id", "api_key", "short_key", "name", "created_at", "last_used_at", "disabled_at")

	enabledKey, err := runQuery[apikey.APIKey](ctx, r.rwConnectionAdminDb, ds)
	if err != nil {
		return apikey.APIKey{}, errors.Wrap(err, "failed to enable user API key")
	}
	return enabledKey, nil
}

func (r *DBAuthRepository) GetAPIKeys(ctx context.Context, userID uint64, keyName *string) ([]apikey.APIKey, error) {
	ds := goqu.Dialect("postgres").From("api_keys_v2").
		Select("api_key_id", "api_key", "short_key", "name", "created_at", "last_used_at", "disabled_at").
		Where(goqu.Ex{"user_id": userID, "deleted_at": nil})

	if keyName != nil {
		ds = ds.Where(goqu.Ex{"name": *keyName})
	}
	keys, err := runQueryRows[[]apikey.APIKey](ctx, r.roConnectionAdminDb, ds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user API keys")
	}
	return keys, nil
}
