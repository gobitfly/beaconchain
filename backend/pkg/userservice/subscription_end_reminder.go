package userservice

import (
	"fmt"
	"html/template"
	"time"

	"github.com/doug-martin/goqu/v9"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/mail"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"golang.org/x/sync/errgroup"
)

type GraceSubscription struct {
	Id          string
	ProductName string
	End         time.Time
	Store       t.ProductStore
}

type ExpiredSubsInfo struct {
	Email       string
	PremiumSubs []GraceSubscription
	AddonSubs   []GraceSubscription
	ApiSubs     []GraceSubscription
}

const reminderFrequency = utils.Week
const gracePeriod = utils.Week * 2

func SubscriptionEndReminder() {
	for {
		start := time.Now()

		// get all subscriptions running on grace period which haven't been warned recently
		gracePeriodSubscriptions, err := getPendingGracePeriodSubscriptionsByUser()
		if err != nil {
			log.Error(err, "error getting subscriptions in grace period", 0)
			time.Sleep(time.Second * 10)
			continue
		}

		mailsSent := 0
		var premiumAddonIds, apiIds []string
		for userId, subsInfo := range gracePeriodSubscriptions {
			// send email
			// not checking/increasing daily rate limit here
			email := types.Email{
				Body:                  formatEmail(subsInfo),
				Title:                 "Subscription payment issue",
				SubscriptionManageURL: "beaconcha.in",
			}
			if err = mail.SendHTMLMail(subsInfo.Email, "beaconcha.in - Subscription payment issue", email, []types.EmailAttachment{}); err != nil {
				log.Error(err, "error sending subscription payment issue email", 0, map[string]interface{}{"email": subsInfo.Email, "user_id": userId})
				continue
			}

			// batch update email sent ts later
			for _, sub := range subsInfo.PremiumSubs {
				premiumAddonIds = append(premiumAddonIds, sub.Id)
			}
			for _, sub := range subsInfo.AddonSubs {
				premiumAddonIds = append(premiumAddonIds, sub.Id)
			}
			for _, sub := range subsInfo.ApiSubs {
				apiIds = append(apiIds, sub.Id)
			}
		}

		// update grace Period warning sent timestamp
		if err := updateGraceTs(premiumAddonIds, "users_app_subscriptions"); err != nil {
			log.Error(err, "error updating premium/addon grace period warning email timestamps", 0)
		}
		if err := updateGraceTs(apiIds, "users_stripe_subscriptions"); err != nil {
			log.Error(err, "error updating api grace period warning email timestamps", 0)
		}

		services.ReportStatus("subscription_end_reminder", "Running", nil)

		log.InfoWithFields(log.Fields{"mails sent": mailsSent, "duration": time.Since(start)}, "sending subscription payment issue warnings completed")
		time.Sleep(time.Hour * 4)
	}
}

func getPendingGracePeriodSubscriptionsByUser() (map[uint64]*ExpiredSubsInfo, error) {
	var err error
	type expiredSubscriptionResult struct {
		Email          string         `db:"email"`
		UserId         uint64         `db:"user_id"`
		SubscriptionId string         `db:"subscription_id"`
		Store          t.ProductStore `db:"store"`
		ProductId      string         `db:"product_id"`
		End            time.Time      `db:"end"`
	}
	baseDs := goqu.Dialect("postgres").
		Select(
			goqu.I("u.email"),
			goqu.I("u.id").As("user_id"),
		).
		From(goqu.T("users").As("u")).
		Where(
			goqu.I("active").Eq(true),
			goqu.Or(
				goqu.L("payment_issues_mail_ts").IsNull(),
				goqu.L("payment_issues_mail_ts").Lt(time.Now().Add(-reminderFrequency)),
			),
		)

	execQuery := func(ds *goqu.SelectDataset, res *[]expiredSubscriptionResult) error {
		query, args, err := ds.Prepared(true).ToSQL()
		if err != nil {
			return err
		}
		err = db.FrontendReaderDB.Select(res, query, args...)
		return err
	}

	wg := errgroup.Group{}
	var premiumResults, apiResults []expiredSubscriptionResult

	// get app (also called premium or mobile) and add-on subscriptions, stored in users_app_subscriptions
	wg.Go(func() error {
		expiredAppSubscriptionssDs := baseDs.
			SelectAppend(
				goqu.L("uas.id").As("subscription_id"),
				goqu.I("uas.product_id"),
				goqu.L("uas.expires_at").As("end"),
				goqu.L("store"),
			).
			LeftJoin(
				goqu.T("users_app_subscriptions").As("uas"),
				goqu.On(
					goqu.I("uas.user_id").Eq(goqu.I("u.id")),
				),
			).
			Where(
				goqu.L("uas.expires_at").Lt(time.Now()),
				goqu.L("EXTRACT(epoch FROM uas.expires_at)").Gt(0),
			)
		return execQuery(expiredAppSubscriptionssDs, &premiumResults)
	})

	// get api subscriptions, stored in users_stripe_subscriptions
	wg.Go(func() error {
		expiredApiSubscriptionssDs := baseDs.
			SelectAppend(
				goqu.L("uss.subscription_id"),
				goqu.I("uss.price_id").As("product_id"),
				goqu.L("to_timestamp((uss.payload->>'current_period_end')::bigint)").As("end"),
				goqu.L("'stripe'").As("store"),
			).
			LeftJoin(
				goqu.T("users_stripe_subscriptions").As("uss"),
				goqu.On(
					goqu.I("u.stripe_customer_id").Eq(goqu.I("uss.customer_id")),
					goqu.I("uss.purchase_group").Eq(utils.GROUP_API),
				),
			).
			Where(
				goqu.L("to_timestamp((uss.payload->>'current_period_end')::bigint)").Lt(time.Now()),
				goqu.L("(uss.payload->>'current_period_end')::bigint").Gt(0),
			)
		err = execQuery(expiredApiSubscriptionssDs, &apiResults)
		if err != nil {
			return err
		}
		for i, res := range apiResults {
			productId := utils.PriceIdToProductId(res.ProductId)
			if productId == "" {
				log.Error(nil, "unmapped stripe subscription price id", 0, map[string]interface{}{"price_id": res.ProductId})
			}
			apiResults[i].ProductId = productId
		}
		return err
	})

	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	subsByUser := make(map[uint64]*ExpiredSubsInfo)
	for _, subResult := range append(premiumResults, apiResults...) {
		switch subResult.Store {
		case t.ProductStoreStripe, t.ProductStoreIosAppstore, t.ProductStoreAndroidPlaystore:
		default:
			// ethpool and custom are not supported
			log.Error(nil, "unsupported subscription store", 0, map[string]interface{}{"store": subResult.Store})
			continue
		}
		productName := utils.EffectiveProductName(subResult.ProductId)
		if productName == "" {
			log.Error(nil, "unmapped subscription product id", 0, map[string]interface{}{"product_id": subResult.ProductId})
			continue
		}

		var subsInfo *ExpiredSubsInfo
		if _, exists := subsByUser[subResult.UserId]; !exists {
			subsInfo = &ExpiredSubsInfo{}
		} else {
			subsInfo = subsByUser[subResult.UserId]
		}

		subsInfo.Email = subResult.Email
		sub := GraceSubscription{
			ProductName: productName,
			End:         subResult.End,
			Id:          subResult.SubscriptionId,
			Store:       subResult.Store,
		}

		switch utils.GetPurchaseGroup(subResult.ProductId) {
		case utils.GROUP_API:
			subsInfo.ApiSubs = append(subsInfo.ApiSubs, sub)
		case utils.GROUP_MOBILE:
			subsInfo.PremiumSubs = append(subsInfo.PremiumSubs, sub)
		case utils.GROUP_ADDON:
			subsInfo.AddonSubs = append(subsInfo.AddonSubs, sub)
		default:
			log.Error(nil, "unmapped subscription product group", 0, map[string]interface{}{"product_id": subResult.ProductId})
			continue
		}
		subsByUser[subResult.UserId] = subsInfo
	}
	return subsByUser, nil
}

func formatEmail(subsInfo *ExpiredSubsInfo) template.HTML {
	var content template.HTML

	content += template.HTML("We had issues processing your subscription payments. You are currently granted a grace period so you can renew your subscription(s). Failure to do so in time could result in your validator dashboards getting archived or permanently deleted!<br><br><br>The following products are affected:<br>")

	formatSubs := func(subs []GraceSubscription, category string) {
		//nolint:gosec // enum string
		tempContent := template.HTML(fmt.Sprintf("<u>%s Subscriptions:</u><br>", category))
		for _, sub := range subs {
			var store string
			switch sub.Store {
			case t.ProductStoreStripe:
				store = fmt.Sprintf(`<a href="%s/pricing">Manage</a>`, utils.Config.Frontend.SiteDomain)
			case t.ProductStoreAndroidPlaystore:
				store = "check Google Play Store"
			case t.ProductStoreIosAppstore:
				store = "check Apple App Store"
			}
			//nolint:gosec
			tempContent += template.HTML(fmt.Sprintf("&emsp;%s (%s, expires on %s)<br>", sub.ProductName, store, sub.End.Add(gracePeriod).Format("Mon Jan 2 2006")))
		}
		content += "<br>"
	}
	formatSubs(subsInfo.PremiumSubs, "Premium")
	formatSubs(subsInfo.AddonSubs, "Premium Add-On")
	formatSubs(subsInfo.ApiSubs, "API")
	return content
}

func updateGraceTs(ids []string, table string) error {
	if len(ids) == 0 {
		return nil
	}
	idColumn := goqu.I("id")
	if table == "users_stripe_subscriptions" {
		idColumn = goqu.I("subscription_id")
	}
	ds := goqu.Dialect("postgres").
		Update(table).
		Set(goqu.Record{"payment_issues_mail_ts": time.Now()}).
		Where(idColumn.In(ids))

	query, args, err := ds.Prepared(true).ToSQL()
	if err != nil {
		return err
	}
	_, err = db.FrontendWriterDB.Exec(query, args...)
	return err
}
