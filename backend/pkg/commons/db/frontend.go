package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
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

	fields := goqu.Record{
		"active":        valid,
		"updated_at":    nowTs,
		"reject_reason": rejectReason,
	}
	if expiration != 0 {
		fields["expires_at"] = expiration
	}

	ds := goqu.Dialect("postgres").
		Update("users_app_subscriptions").
		Set(fields).
		Where(goqu.I("id").Eq(id))

	qry, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}

	if tx == nil {
		_, err = FrontendWriterDB.Exec(qry, args)
	} else {
		_, err = tx.Exec(qry, args)
	}

	return err
}
