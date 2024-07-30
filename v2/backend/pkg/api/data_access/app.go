// Copyright (C) 2024 Bitfly GmbH
//
// This file is part of Beaconchain.
//
// Beaconchain is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain.  If not, see <https://www.gnu.org/licenses/>.

package dataaccess

import (
	"database/sql"
	"fmt"
	"time"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/userservice"
	"github.com/pkg/errors"
)

type AppRepository interface {
	GetUserIdByRefreshToken(claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error)
	MigrateMobileSession(oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error
	AddUserDevice(userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error
	GetAppDataFromRedirectUri(callback string) (*t.OAuthAppData, error)
	AddMobileNotificationToken(userID uint64, deviceID, notifyToken string) error
	GetAppSubscriptionCount(userID uint64) (uint64, error)
	AddMobilePurchase(tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error
}

// GetUserIdByRefreshToken basically used to confirm the claimed user id with the refresh token. Returns the userId if successful
func (d *DataAccessService) GetUserIdByRefreshToken(claimUserID, claimAppID, claimDeviceID uint64, hashedRefreshToken string) (uint64, error) {
	if hashedRefreshToken == "" { // sanity
		return 0, errors.New("empty refresh token")
	}
	var userID uint64
	err := d.userWriter.Get(&userID,
		`SELECT user_id FROM users_devices WHERE user_id = $1 AND 
			refresh_token = $2 AND app_id = $3 AND id = $4 AND active = true`, claimUserID, hashedRefreshToken, claimAppID, claimDeviceID)
	if errors.Is(err, sql.ErrNoRows) {
		return userID, fmt.Errorf("%w: user not found via refresh token", ErrNotFound)
	}
	return userID, err
}

func (d *DataAccessService) MigrateMobileSession(oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName string) error {
	result, err := d.userWriter.Exec("UPDATE users_devices SET refresh_token = $2, device_identifier = $3, device_name = $4 WHERE refresh_token = $1", oldHashedRefreshToken, newHashedRefreshToken, deviceID, deviceName)
	if err != nil {
		return errors.Wrap(err, "Error updating refresh token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "Error getting rows affected")
	}

	if rowsAffected != 1 {
		return errors.New(fmt.Sprintf("illegal number of rows affected, expected 1 got %d", rowsAffected))
	}

	return err
}

func (d *DataAccessService) GetAppDataFromRedirectUri(callback string) (*t.OAuthAppData, error) {
	data := t.OAuthAppData{}
	err := d.userWriter.Get(&data, "SELECT id, app_name, redirect_uri, active, owner_id FROM oauth_apps WHERE active = true AND redirect_uri = $1", callback)
	return &data, err
}

func (d *DataAccessService) AddUserDevice(userID uint64, hashedRefreshToken string, deviceID, deviceName string, appID uint64) error {
	_, err := d.userWriter.Exec("INSERT INTO users_devices (user_id, refresh_token, device_identifier, device_name, app_id, created_ts) VALUES($1, $2, $3, $4, $5, 'NOW()') ON CONFLICT DO NOTHING",
		userID, hashedRefreshToken, deviceID, deviceName, appID,
	)
	return err
}

func (d *DataAccessService) AddMobileNotificationToken(userID uint64, deviceID, notifyToken string) error {
	_, err := d.userWriter.Exec("UPDATE users_devices SET notification_token = $1 WHERE user_id = $2 AND device_identifier = $3;",
		notifyToken, userID, deviceID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: user mobile device not found", ErrNotFound)
	}
	return err
}

func (d *DataAccessService) GetAppSubscriptionCount(userID uint64) (uint64, error) {
	var count uint64
	err := d.userReader.Get(&count, "SELECT COUNT(receipt) FROM users_app_subscriptions WHERE user_id = $1", userID)
	return count, err
}

func (d *DataAccessService) AddMobilePurchase(tx *sql.Tx, userID uint64, paymentDetails t.MobileSubscription, verifyResponse *userservice.VerifyResponse, extSubscriptionId string) error {
	now := time.Now()
	nowTs := now.Unix()
	receiptHash := utils.HashAndEncode(verifyResponse.Receipt)

	query := `INSERT INTO users_app_subscriptions 
				(user_id, product_id, price_micros, currency, created_at, updated_at, validate_remotely, active, store, receipt, expires_at, reject_reason, receipt_hash, subscription_id) 
				VALUES($1, $2, $3, $4, TO_TIMESTAMP($5), TO_TIMESTAMP($6), $7, $8, $9, $10, TO_TIMESTAMP($11), $12, $13, $14) 
			  ON CONFLICT(receipt_hash) DO UPDATE SET product_id = $2, active = $7, updated_at = TO_TIMESTAMP($5);`
	var err error
	if tx == nil {
		_, err = d.userWriter.Exec(query,
			userID, verifyResponse.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, verifyResponse.Valid, verifyResponse.Valid, paymentDetails.Transaction.Type, verifyResponse.Receipt, verifyResponse.ExpirationDate, verifyResponse.RejectReason, receiptHash, extSubscriptionId,
		)
	} else {
		_, err = tx.Exec(query,
			userID, verifyResponse.ProductID, paymentDetails.PriceMicros, paymentDetails.Currency, nowTs, nowTs, verifyResponse.Valid, verifyResponse.Valid, paymentDetails.Transaction.Type, verifyResponse.Receipt, verifyResponse.ExpirationDate, verifyResponse.RejectReason, receiptHash, extSubscriptionId,
		)
	}

	return err
}
