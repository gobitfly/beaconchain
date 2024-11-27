package userservice

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/lib/pq"
)

func StripeEmailUpdater() {
	for {
		// fetch all users with pending stripe email updates
		var pendingUsers []struct {
			Email            string `db:"email"`
			StripeCustomerId string `db:"stripe_customer_id"`
		}
		err := db.FrontendReaderDB.Select(&pendingUsers, "SELECT email, stripe_customer_id FROM users WHERE stripe_email_pending AND stripe_customer_id <> ''")
		if err != nil {
			log.Error(err, "error getting pending users for stripe email update service", 0)
			time.Sleep(time.Second * 10)
			continue
		}

		// update stripe customer email
		var updatedUsers []string
		for _, user := range pendingUsers {
			err := updateStripeCustomerEmail(user.StripeCustomerId, user.Email)
			if err != nil {
				log.Error(err, "error updating stripe customer email", 0, map[string]interface{}{"email": user.Email, "stripe_customer_id": user.StripeCustomerId})
			} else {
				updatedUsers = append(updatedUsers, user.Email)
			}
			time.Sleep(time.Millisecond * 200)
		}

		// set stripe_email_pending flag to false for all users that were updated
		if len(updatedUsers) > 0 {
			_, err := db.FrontendWriterDB.Exec("UPDATE users SET stripe_email_pending = false WHERE email = ANY($1)", pq.Array(updatedUsers))
			if err != nil {
				log.Error(err, "error setting stripe_email_pending flag false for users, stripe email was updated", 0, map[string]interface{}{"emails": updatedUsers})
				time.Sleep(time.Second * 10)
				continue
			}
		}

		services.ReportStatus("stripe_email_updater", "Running", nil)

		time.Sleep(time.Minute)
	}
}

func updateStripeCustomerEmail(stripeCustomerId, newEmail string) error {
	// see https://stripe.com/docs/api/customers/update
	apiEndpoint := "https://api.stripe.com/v1/customers/" + stripeCustomerId

	data := url.Values{}
	data.Set("email", newEmail)
	req, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating email change request for stripe: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", utils.Config.Frontend.Stripe.SecretKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.Client{Timeout: time.Second * 10}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to stripe: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error updating email in stripe, also could not read body: %w", err)
		}
		return fmt.Errorf("error updating email in stripe: %w; body: %v", err, string(body))
	}
	return nil
}
