package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (uint64, error)
	CreateUser(ctx context.Context, email, password string) (uint64, error)
	RemoveUser(ctx context.Context, userId uint64) error
	UpdateUserEmail(ctx context.Context, userId uint64) error
	UpdateUserPassword(ctx context.Context, userId uint64, password string) error
	GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error)
	GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error)
	IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error)
	UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error
	UpdatePasswordResetTime(ctx context.Context, userId uint64) error
	UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error
	UpdatePasswordResetHash(ctx context.Context, userId uint64, passwordHash string) error
	GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error)
	GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error)
	GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error)
	GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error)
	GetUserInfo(ctx context.Context, id uint64) (*t.UserInfo, error)
	GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error)
	GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error)
}

func (d *DataAccessService) GetUserByEmail(ctx context.Context, email string) (uint64, error) {
	result := uint64(0)
	err := d.userReader.GetContext(ctx, &result, `SELECT id FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user not found", ErrNotFound)
	}
	return result, err
}

func (d *DataAccessService) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	// (password is already hashed)
	var result uint64

	apiKey, err := utils.GenerateRandomAPIKey()
	if err != nil {
		return 0, err
	}

	err = d.userWriter.GetContext(ctx, &result, `
    	INSERT INTO users (password, email, register_ts, api_key)
      		VALUES ($1, $2, NOW(), $3)
		RETURNING id`,
		password, email, apiKey,
	)

	return result, err
}

func (d *DataAccessService) RemoveUser(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userId)
	return err
}

func (d *DataAccessService) UpdateUserEmail(ctx context.Context, userId uint64) error {
	// Called after user clicked link for email confirmations + changes, so:
	// set email_confirmed true, set email (from email_change_to_value), update stripe email
	// unset email_confirmation_hash

	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email = email_change_to_value,
			email_change_to_value = NULL,
			email_confirmed = true,
			email_confirmation_hash = NULL,
			stripe_email_pending = true
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) UpdateUserPassword(ctx context.Context, userId uint64, password string) error {
	// (password is already hashed)

	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password = $1,
			password_reset_hash = NULL
		WHERE id = $2
	`, password, userId)

	return err
}

func (d *DataAccessService) GetEmailConfirmationTime(ctx context.Context, userId uint64) (time.Time, error) {
	result := time.Time{}

	var queryResult sql.NullTime
	err := d.userReader.GetContext(ctx, &queryResult, `
    	SELECT
			email_confirmation_ts
		FROM users
		WHERE id = $1`, userId)

	if queryResult.Valid {
		result = queryResult.Time
	}

	return result, err
}

func (d *DataAccessService) GetPasswordResetTime(ctx context.Context, userId uint64) (time.Time, error) {
	result := time.Time{}

	var queryResult sql.NullTime
	err := d.userReader.GetContext(ctx, &queryResult, `
    	SELECT
			password_reset_ts
		FROM users
		WHERE id = $1`, userId)

	if queryResult.Valid {
		result = queryResult.Time
	}

	return result, err
}

func (d *DataAccessService) UpdateEmailConfirmationTime(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email_confirmation_ts = NOW()
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) IsPasswordResetAllowed(ctx context.Context, userId uint64) (bool, error) {
	var result bool

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			password_reset_not_allowed
		FROM users
		WHERE id = $1`, userId)

	return !result, err
}

func (d *DataAccessService) UpdatePasswordResetTime(ctx context.Context, userId uint64) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password_reset_ts = NOW()
		WHERE id = $1
	`, userId)

	return err
}

func (d *DataAccessService) UpdateEmailConfirmationHash(ctx context.Context, userId uint64, email, confirmationHash string) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			email_confirmation_hash = $1,
			email_change_to_value = $2
		WHERE id = $3
	`, confirmationHash, email, userId)

	return err
}

func (d *DataAccessService) UpdatePasswordResetHash(ctx context.Context, userId uint64, confirmationHash string) error {
	_, err := d.userWriter.ExecContext(ctx, `
		UPDATE users 
		SET 
			password_reset_hash = $1
		WHERE id = $2
	`, confirmationHash, userId)

	return err
}

func (d *DataAccessService) GetUserCredentialInfo(ctx context.Context, userId uint64) (*t.UserCredentialInfo, error) {
	// TODO @patrick post-beta improve product-mgmt
	// TODO @DATA-ACCESS i quickly hacked this together, maybe improve
	result := &t.UserCredentialInfo{}
	err := d.userReader.GetContext(ctx, result, `
		WITH
			latest_and_greatest_sub AS (
				SELECT user_id, product_id FROM users_app_subscriptions
				LEFT JOIN users ON users.id = user_id AND product_id IN ('orca.yearly', 'orca', 'dolphin.yearly', 'dolphin', 'guppy.yearly', 'guppy', 'whale', 'goldfish', 'plankton')
				WHERE users.id = $1 AND active = true
				ORDER BY CASE product_id
					WHEN 'orca.yearly'    THEN  1
					WHEN 'orca'           THEN  2
					WHEN 'dolphin.yearly' THEN  3
					WHEN 'dolphin'        THEN  4
					WHEN 'guppy.yearly'   THEN  5
					WHEN 'guppy'          THEN  6
					WHEN 'whale'          THEN  7
					WHEN 'goldfish'       THEN  8
					WHEN 'plankton'       THEN  9
					ELSE                       10  -- For any other product_id values
				END, users_app_subscriptions.created_at DESC LIMIT 1
			)
		SELECT users.id AS id, users.email, users.email_confirmed, password, COALESCE(product_id, '') AS product_id, COALESCE(user_group, '') AS user_group
		FROM users
		LEFT JOIN latest_and_greatest_sub ON latest_and_greatest_sub.user_id = users.id
		WHERE users.id = $1`, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: user not found", ErrNotFound)
	}
	return result, err
}

func (d *DataAccessService) GetUserIdByApiKey(ctx context.Context, apiKey string) (uint64, error) {
	var userId uint64
	err := d.userReader.GetContext(ctx, &userId, `SELECT user_id FROM api_keys WHERE api_key = $1 LIMIT 1`, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user for api_key not found", ErrNotFound)
	}
	return userId, err
}

func (d *DataAccessService) GetUserIdByConfirmationHash(ctx context.Context, hash string) (uint64, error) {
	var result uint64

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			id
		FROM users
		WHERE email_confirmation_hash = $1`, hash)

	return result, err
}

func (d *DataAccessService) GetUserIdByResetHash(ctx context.Context, hash string) (uint64, error) {
	var result uint64

	err := d.userReader.GetContext(ctx, &result, `
    	SELECT
			id
		FROM users
		WHERE password_reset_hash = $1`, hash)

	return result, err
}

func (d *DataAccessService) GetUserInfo(ctx context.Context, userId uint64) (*t.UserInfo, error) {
	return db.GetUserInfo(ctx, userId, d.userReader)
}

func (d *DataAccessService) GetProductSummary(ctx context.Context) (*t.ProductSummary, error) {
	return db.GetProductSummary(ctx)
}

func (d *DataAccessService) GetFreeTierPerks(ctx context.Context) (*t.PremiumPerks, error) {
	return db.GetFreeTierPerks(ctx)
}

func (d *DataAccessService) GetUserDashboards(ctx context.Context, userId uint64) (*t.UserDashboardsData, error) {
	result := &t.UserDashboardsData{}

	wg := errgroup.Group{}

	validatorDashboardMap := make(map[uint64]*t.ValidatorDashboard, 0)
	wg.Go(func() error {
		dbReturn := []struct {
			Id           uint64         `db:"id"`
			Name         string         `db:"name"`
			Network      uint64         `db:"network"`
			IsArchived   sql.NullString `db:"is_archived"`
			PublicId     sql.NullString `db:"public_id"`
			PublicName   sql.NullString `db:"public_name"`
			SharedGroups sql.NullBool   `db:"shared_groups"`
		}{}

		err := d.alloyReader.SelectContext(ctx, &dbReturn, `
		SELECT
			uvd.id,
			uvd.name,
			uvd.network,
			uvd.is_archived,
			uvds.public_id,
			uvds.name AS public_name,
			uvds.shared_groups
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_sharing uvds ON uvd.id = uvds.dashboard_id
		WHERE uvd.user_id = $1
	`, userId)
		if err != nil {
			return err
		}

		for _, row := range dbReturn {
			if _, ok := validatorDashboardMap[row.Id]; !ok {
				validatorDashboardMap[row.Id] = &t.ValidatorDashboard{
					Id:             row.Id,
					Name:           row.Name,
					Network:        row.Network,
					PublicIds:      []t.VDBPublicId{},
					IsArchived:     row.IsArchived.Valid,
					ArchivedReason: row.IsArchived.String,
				}
			}
			if row.PublicId.Valid {
				publicId := t.VDBPublicId{}
				publicId.PublicId = row.PublicId.String
				publicId.Name = row.PublicName.String
				publicId.ShareSettings.ShareGroups = row.SharedGroups.Bool

				validatorDashboardMap[row.Id].PublicIds = append(validatorDashboardMap[row.Id].PublicIds, publicId)
			}
		}

		return nil
	})

	type DashboardCount struct {
		Id             uint64 `db:"id"`
		GroupCount     uint64 `db:"group_count"`
		ValidatorCount uint64 `db:"validator_count"`
	}

	validatorDashboardCountMap := make(map[uint64]DashboardCount, 0)
	wg.Go(func() error {
		dbReturn := []DashboardCount{}

		err := d.alloyReader.SelectContext(ctx, &dbReturn, `
		SELECT
			uvd.id,
			COUNT(DISTINCT(uvdg.id)) AS group_count,
			COUNT(DISTINCT(uvdv.validator_index)) AS validator_count
		FROM users_val_dashboards uvd
		LEFT JOIN users_val_dashboards_groups uvdg ON uvd.id = uvdg.dashboard_id
		LEFT JOIN users_val_dashboards_validators uvdv ON uvd.id = uvdv.dashboard_id
		WHERE uvd.user_id = $1
		GROUP BY uvd.id
	`, userId)
		if err != nil {
			return err
		}

		for _, row := range dbReturn {
			validatorDashboardCountMap[row.Id] = row
		}

		return nil
	})

	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("error retrieving user dashboards data: %w", err)
	}

	// Fill the result
	for _, validatorDashboard := range validatorDashboardMap {
		validatorDashboard.GroupCount = validatorDashboardCountMap[validatorDashboard.Id].GroupCount
		validatorDashboard.ValidatorCount = validatorDashboardCountMap[validatorDashboard.Id].ValidatorCount

		result.ValidatorDashboards = append(result.ValidatorDashboards, *validatorDashboard)
	}

	// Get the account dashboards
	err = d.alloyReader.SelectContext(ctx, &result.AccountDashboards, `
		SELECT
			id,
			name
		FROM users_acc_dashboards
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// return number of active / archived dashboards
func (d *DataAccessService) GetUserValidatorDashboardCount(ctx context.Context, userId uint64, active bool) (uint64, error) {
	var count uint64
	err := d.alloyReader.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM users_val_dashboards
		WHERE user_id = $1 AND (($2 AND is_archived IS NULL) OR (NOT $2 AND is_archived IS NOT NULL))
	`, userId, active)

	return count, err
}
