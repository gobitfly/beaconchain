package db

import (
	"database/sql"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

var FrontendReaderDB *sqlx.DB
var FrontendWriterDB *sqlx.DB

func GetAllAppSubscriptions() ([]*types.PremiumData, error) {
	data := []*types.PremiumData{}

	err := FrontendWriterDB.Select(&data,
		"SELECT id, receipt, store, active, expires_at, product_id, user_id, validate_remotely from users_app_subscriptions WHERE validate_remotely = true order by id desc",
	)

	return data, err
}

func UpdateUserSubscriptionProduct(tx *sql.Tx, id uint64, productID string) error {
	var err error
	if tx == nil {
		_, err = FrontendWriterDB.Exec("UPDATE users_app_subscriptions SET product_id = $1 WHERE id = $2;",
			productID, id,
		)
	} else {
		_, err = tx.Exec("UPDATE users_app_subscriptions SET product_id = $1 WHERE id = $2",
			productID, id,
		)
	}

	return err
}

func SetSubscriptionToExpired(tx *sql.Tx, id uint64) error {
	var err error
	query := "UPDATE users_app_subscriptions SET validate_remotely = false, reject_reason = 'expired' WHERE id = $1;"
	if tx == nil {
		_, err = FrontendWriterDB.Exec(query,
			id,
		)
	} else {
		_, err = tx.Exec(query,
			id,
		)
	}

	return err
}

func UpdateUserSubscription(tx *sql.Tx, id uint64, valid bool, expiration int64, rejectReason string) error {
	now := time.Now()
	nowTs := now.Unix()
	var err error
	if tx == nil {
		_, err = FrontendWriterDB.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE id = $5;",
			valid, nowTs, expiration, rejectReason, id,
		)
	} else {
		_, err = tx.Exec("UPDATE users_app_subscriptions SET active = $1, updated_at = TO_TIMESTAMP($2), expires_at = TO_TIMESTAMP($3), reject_reason = $4 WHERE id = $5;",
			valid, nowTs, expiration, rejectReason, id,
		)
	}

	return err
}
