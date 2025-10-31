package userrepo

import (
	"context"

	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DBRepository struct {
	roConnectionAdminDb *sqlx.DB
	rwConnectionAdminDb *sqlx.DB
}

type dbUser struct {
	ID       uint64 `db:"user_id"`
	TierName string `db:"tier_name"`
}

func (r *DBRepository) Initialize(roConnectionAdminDb data_sources.AdminRoConnection, rwConnectionAdminDb data_sources.AdminRwConnection) {
	r.roConnectionAdminDb = roConnectionAdminDb
	r.rwConnectionAdminDb = rwConnectionAdminDb
}

func (r *DBRepository) Ping() error {
	if r.roConnectionAdminDb == nil {
		return fmt.Errorf("read connection not initialized")
	}
	if err := r.roConnectionAdminDb.Ping(); err != nil {
		return fmt.Errorf("read connection ping failed: %w", err)
	}

	if r.rwConnectionAdminDb == nil {
		return fmt.Errorf("write connection not initialized")
	}
	if err := r.rwConnectionAdminDb.Ping(); err != nil {
		return fmt.Errorf("write connection ping failed: %w", err)
	}
	return nil
}

func (r *DBRepository) Get(ctx context.Context, id uint64) (domain.User, error) {
	return queryUser(ctx, r.roConnectionAdminDb, id)
}

func queryUser[T uint64 | *goqu.SelectDataset](ctx context.Context, db *sqlx.DB, id T) (domain.User, error) {
	dialect := goqu.Dialect("postgres")

	ds := dialect.From(goqu.T("users").As("u")).
		Select(
			goqu.I("u.id").As("user_id"),
			goqu.COALESCE(goqu.I("pt.tier_name"), goqu.L("'FREE'")).As("tier_name"),
		).
		LeftJoin(
			goqu.T("users_stripe_subscriptions").As("uss"),
			goqu.On(goqu.I("u.stripe_customer_id").Eq(goqu.I("uss.customer_id"))),
		).
		LeftJoin(
			goqu.T("subscription_tiers").As("pt"),
			goqu.On(goqu.I("uss.price_id").Eq(goqu.I("pt.price_id"))),
		).
		Where(
			goqu.I("u.id").Eq(id),
			goqu.Or(
				goqu.I("uss.purchase_group").Eq("api"), // filter out mobile subscriptions
				goqu.I("uss.purchase_group").IsNull(),
			),
			goqu.Or(
				goqu.I("uss.active").IsTrue(),
				goqu.I("uss.active").IsNull(),
			),
		).
		Limit(1)

	dbUser, err := repo.RunQuery[dbUser](ctx, db, ds)
	if err != nil {
		return domain.User{}, errors.Wrap(err, "failed to query user")
	}

	user := domain.User{
		ID:               dbUser.ID,
		SubscriptionTier: domain.Tier(dbUser.TierName),
	}

	return user, nil
}

func (r *DBRepository) Create(ctx context.Context, email string, initialApiKey string, hashedPassword string) (domain.User, error) {
	user := domain.User{}

	err := r.rwConnectionAdminDb.GetContext(ctx, &user, `
	    	INSERT INTO users (password, email, register_ts, api_key)
	      		VALUES ($1, $2, NOW(), $3)
			RETURNING *`,
		hashedPassword, email, initialApiKey)

	return user, err
}

func (r *DBRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.rwConnectionAdminDb.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
