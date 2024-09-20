package notification

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

type validatorProposalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slot           uint64
	Status         uint64 // * Can be 0 = scheduled, 1 executed, 2 missed */
	Reward         float64
}

func (n *validatorProposalNotification) GetInfo(includeUrl bool) string {
	var generalPart, suffix string
	vali := strconv.FormatUint(n.ValidatorIndex, 10)
	slot := strconv.FormatUint(n.Slot, 10)
	if includeUrl {
		vali = fmt.Sprintf(`<a href="https://%[1]v/validator/%[2]v">%[2]v</a>`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex)
		slot = fmt.Sprintf(`<a href="https://%[1]v/slot/%[2]v">%[2]v</a>`, utils.Config.Frontend.SiteDomain, n.Slot)
		suffix = getUrlPart(n.ValidatorIndex)
	}

	dashboardAndGroupInfo := ""
	if n.DashboardId != nil {
		dashboardAndGroupInfo = fmt.Sprintf(` of Group <b>%[4]v</b> in Dashboard <a href="https://%[1]v/dashboard/%[6]v">%[5]v</a>`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, n.DashboardGroupName, n.DashboardName, &n.DashboardId)
	}
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`New scheduled block proposal at slot %s for Validator %s%s.`, slot, vali, dashboardAndGroupInfo)
	case 1:
		generalPart = fmt.Sprintf(`Validator %s%s proposed block at slot %s with %v %v execution reward.`, vali, dashboardAndGroupInfo, slot, n.Reward, utils.Config.Frontend.ElCurrency)
	case 2:
		generalPart = fmt.Sprintf(`Validator %s%s missed a block proposal at slot %s.`, vali, dashboardAndGroupInfo, slot)
	case 3:
		generalPart = fmt.Sprintf(`Validator %s%s had an orphaned block proposal at slot %s.`, vali, dashboardAndGroupInfo, slot)
	}
	return generalPart + suffix
}

func (n *validatorProposalNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Block Proposal Scheduled"
	case 1:
		return "New Block Proposal"
	case 2:
		return "Block Proposal Missed"
	case 3:
		return "Block Proposal Missed (Orphaned)"
	}
	return "-"
}

func (n *validatorProposalNotification) GetInfoMarkdown() string {
	var generalPart = ""
	dashboardAndGroupInfo := ""
	if n.DashboardId != nil {
		dashboardAndGroupInfo = fmt.Sprintf(` of Group **%[4]v** in Dashboard [%[5]v](https://%[1]v/dashboard/%[6]v)`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, n.DashboardGroupName, n.DashboardName, &n.DashboardId)
	}
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`New scheduled block proposal at slot [%[3]v](https://%[1]v/slot/%[3]v) for Validator [%[2]v](https://%[1]v/validator/%[2]v)%[4]s.`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, dashboardAndGroupInfo)
	case 1:
		generalPart = fmt.Sprintf(`Validator [%[2]v](https://%[1]v/validator/%[2]v)%[6]s proposed a new block at slot [%[3]v](https://%[1]v/slot/%[3]v) with %[4]v %[5]v execution reward.`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, n.Reward, utils.Config.Frontend.ElCurrency, dashboardAndGroupInfo)
	case 2:
		generalPart = fmt.Sprintf(`Validator [%[2]v](https://%[1]v/validator/%[2]v)%[4]s missed a block proposal at slot [%[3]v](https://%[1]v/slot/%[3]v).`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, dashboardAndGroupInfo)
	case 3:
		generalPart = fmt.Sprintf(`Validator [%[2]v](https://%[1]v/validator/%[2]v)%[4]s had an orphaned block proposal at slot [%[3]v](https://%[1]v/slot/%[3]v).`, utils.Config.Frontend.SiteDomain, n.ValidatorIndex, n.Slot, dashboardAndGroupInfo)
	}

	return generalPart
}

type validatorIsOfflineNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	IsOffline      bool
}

// Overwrite specific methods
func (n *validatorIsOfflineNotification) GetInfo(includeUrl bool) string {
	if n.IsOffline {
		if includeUrl {
			return fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> is offline since epoch <a href="https://%[3]v/epoch/%[2]s">%[2]s</a>).`, n.ValidatorIndex, n.LatestState, utils.Config.Frontend.SiteDomain)
		} else {
			return fmt.Sprintf(`Validator %v is offline since epoch %s.`, n.ValidatorIndex, n.LatestState)
		}
	} else {
		if includeUrl {
			return fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> is back online since epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
		} else {
			return fmt.Sprintf(`Validator %v is back online since epoch %v.`, n.ValidatorIndex, n.Epoch)
		}
	}
}

func (n *validatorIsOfflineNotification) GetTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

func (n *validatorIsOfflineNotification) GetInfoMarkdown() string {
	if n.IsOffline {
		return fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) is offline since epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
	} else {
		return fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) is back online since epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
	}
}

type validatorAttestationNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex     uint64
	ValidatorPublicKey string
	Status             uint64 // * Can be 0 = scheduled | missed, 1 executed
}

func (n *validatorAttestationNotification) GetInfo(includeUrl bool) string {
	var generalPart = ""
	if includeUrl {
		if n.DashboardId == nil { // leagcy notifications
			switch n.Status {
			case 0:
				generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> missed an attestation in epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
			case 1:
				generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> submitted a successful attestation for epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
			}
		} else { // dashboard based notifications
			switch n.Status {
			case 0:
				generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> of Group <b>%[4]v</b> in Dashboard <a href="https://%[3]v/dashboard/%[6]v">%[5]v</a> missed an attestation in epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain, n.DashboardGroupName, n.DashboardName, n.DashboardId)
			case 1:
				generalPart = fmt.Sprintf(`Validator <a href="https://%[3]v/validator/%[1]v">%[1]v</a> of Group <b>%[4]v</b> in Dashboard <a href="https://%[3]v/dashboard/%[6]v">%[5]v</a> submitted a successful attestation for epoch <a href="https://%[3]v/epoch/%[2]v">%[2]v</a>.`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain, n.DashboardGroupName, n.DashboardName, n.DashboardId)
			}
		}
		// return generalPart + getUrlPart(n.ValidatorIndex)
	} else {
		if n.DashboardId == nil { // leagcy notifications
			switch n.Status {
			case 0:
				generalPart = fmt.Sprintf(`Validator %v missed an attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
			case 1:
				generalPart = fmt.Sprintf(`Validator %v submitted a successful attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
			}
		} else { // dashboard based notifications
			switch n.Status {
			case 0:
				generalPart = fmt.Sprintf(`Validator %v of Group %v in Dashboard %v missed an attestation in epoch %v.`, n.ValidatorIndex, n.DashboardGroupName, n.DashboardName, n.Epoch)
			case 1:
				generalPart = fmt.Sprintf(`Validator %v of Group %v in Dashboard %v submitted a successful attestation in epoch %v.`, n.ValidatorIndex, n.DashboardGroupName, n.DashboardName, n.Epoch)
			}
		}
	}
	return generalPart
}

func (n *validatorAttestationNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

func (n *validatorAttestationNotification) GetInfoMarkdown() string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) missed an attestation in epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
	case 1:
		generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) submitted a successful attestation in epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
	}

	if n.DashboardId == nil { // leagcy notifications
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) missed an attestation in epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
		case 1:
			generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) submitted a successful attestation in epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain)
		}
	} else { // dashboard based notifications
		switch n.Status {
		case 0:
			generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) of Group **%[4]v** in Dashboard [%[4]v](https://%[3]v/dashboard/%[6]v) missed an attestation in epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain, n.DashboardGroupName, n.DashboardName, n.DashboardId)
		case 1:
			generalPart = fmt.Sprintf(`Validator [%[1]v](https://%[3]v/validator/%[1]v) of Group **%[4]v** in Dashboard [%[4]v](https://%[3]v/dashboard/%[6]v) submitted a successful attestation for epoch [%[2]v](https://%[3]v/epoch/%[2]v).`, n.ValidatorIndex, n.Epoch, utils.Config.Frontend.SiteDomain, n.DashboardGroupName, n.DashboardName, n.DashboardId)
		}
	}
	return generalPart
}

type validatorGotSlashedNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slasher        uint64
	Reason         string
}

func (n *validatorGotSlashedNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`Validator %v has been slashed at epoch %v by validator %v for %s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason)
	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorGotSlashedNotification) GetTitle() string {
	return "Validator got Slashed"
}

func (n *validatorGotSlashedNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`Validator [%[1]v](https://%[5]v/validator/%[1]v) has been slashed at epoch [%[2]v](https://%[5]v/epoch/%[2]v) by validator [%[3]v](https://%[5]v/validator/%[3]v) for %[4]s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason, utils.Config.Frontend.SiteDomain)
	return generalPart
}

type validatorWithdrawalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Epoch          uint64
	Slot           uint64
	Amount         uint64
	Address        []byte
}

func (n *validatorWithdrawalNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`An automatic withdrawal of %v has been processed for validator %v.`, utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false), n.ValidatorIndex)
	if includeUrl {
		return generalPart + getUrlPart(n.ValidatorIndex)
	}
	return generalPart
}

func (n *validatorWithdrawalNotification) GetTitle() string {
	return "Withdrawal Processed"
}

func (n *validatorWithdrawalNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`An automatic withdrawal of %[2]v has been processed for validator [%[1]v](https://%[6]v/validator/%[1]v) during slot [%[3]v](https://%[6]v/slot/%[3]v). The funds have been sent to: [%[4]v](https://%[6]v/address/0x%[5]x).`, n.ValidatorIndex, utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false), n.Slot, utils.FormatHashRaw(n.Address), n.Address, utils.Config.Frontend.SiteDomain)
	return generalPart
}

type ethClientNotification struct {
	types.NotificationBaseImpl

	EthClient string
}

func (n *ethClientNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	if includeUrl {
		url := ""
		switch n.EthClient {
		case "Geth":
			url = "https://github.com/ethereum/go-ethereum/releases"
		case "Nethermind":
			url = "https://github.com/NethermindEth/nethermind/releases"
		case "Teku":
			url = "https://github.com/ConsenSys/teku/releases"
		case "Prysm":
			url = "https://github.com/prysmaticlabs/prysm/releases"
		case "Nimbus":
			url = "https://github.com/status-im/nimbus-eth2/releases"
		case "Lighthouse":
			url = "https://github.com/sigp/lighthouse/releases"
		case "Erigon":
			url = "https://github.com/erigontech/erigon/releases"
		case "Rocketpool":
			url = "https://github.com/rocket-pool/smartnode-install/releases"
		case "MEV-Boost":
			url = "https://github.com/flashbots/mev-boost/releases"
		case "Lodestar":
			url = "https://github.com/chainsafe/lodestar/releases"
		default:
			url = "https://beaconcha.in/ethClients"
		}

		return generalPart + " " + url
	}
	return generalPart
}

func (n *ethClientNotification) GetTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
}

func (n *ethClientNotification) GetInfoMarkdown() string {
	url := ""
	switch n.EthClient {
	case "Geth":
		url = "https://github.com/ethereum/go-ethereum/releases"
	case "Nethermind":
		url = "https://github.com/NethermindEth/nethermind/releases"
	case "Teku":
		url = "https://github.com/ConsenSys/teku/releases"
	case "Prysm":
		url = "https://github.com/prysmaticlabs/prysm/releases"
	case "Nimbus":
		url = "https://github.com/status-im/nimbus-eth2/releases"
	case "Lighthouse":
		url = "https://github.com/sigp/lighthouse/releases"
	case "Erigon":
		url = "https://github.com/erigontech/erigon/releases"
	case "Rocketpool":
		url = "https://github.com/rocket-pool/smartnode-install/releases"
	case "MEV-Boost":
		url = "https://github.com/flashbots/mev-boost/releases"
	case "Lodestar":
		url = "https://github.com/chainsafe/lodestar/releases"
	default:
		url = "https://beaconcha.in/ethClients"
	}

	generalPart := fmt.Sprintf(`A new version for [%s](%s) is available.`, n.EthClient, url)

	return generalPart
}

type MachineEvents struct {
	SubscriptionID  uint64         `db:"id"`
	UserID          types.UserId   `db:"user_id"`
	MachineName     string         `db:"machine"`
	UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	EventThreshold  float64        `db:"event_threshold"`
}

type monitorMachineNotification struct {
	types.NotificationBaseImpl

	MachineName string
}

func (n *monitorMachineNotification) GetInfo(includeUrl bool) string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return fmt.Sprintf(`Your staking machine "%v" is running low on storage space.`, n.MachineName)
	case types.MonitoringMachineOfflineEventName:
		return fmt.Sprintf(`Your staking machine "%v" might be offline. It has not been seen for a couple minutes now.`, n.MachineName)
	case types.MonitoringMachineCpuLoadEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured CPU usage threshold.`, n.MachineName)
	case types.MonitoringMachineSwitchedToETH1FallbackEventName:
		return fmt.Sprintf(`Your staking machine "%v" has switched to your configured ETH1 fallback`, n.MachineName)
	case types.MonitoringMachineSwitchedToETH2FallbackEventName:
		return fmt.Sprintf(`Your staking machine "%v" has switched to your configured ETH2 fallback`, n.MachineName)
	case types.MonitoringMachineMemoryUsageEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured RAM threshold.`, n.MachineName)
	}
	return ""
}

func (n *monitorMachineNotification) GetTitle() string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return "Storage Warning"
	case types.MonitoringMachineOfflineEventName:
		return "Staking Machine Offline"
	case types.MonitoringMachineCpuLoadEventName:
		return "High CPU Load"
	case types.MonitoringMachineSwitchedToETH1FallbackEventName:
		return "ETH1 Fallback Active"
	case types.MonitoringMachineSwitchedToETH2FallbackEventName:
		return "ETH2 Fallback Active"
	case types.MonitoringMachineMemoryUsageEventName:
		return "Memory Warning"
	}
	return ""
}

func (n *monitorMachineNotification) GetEventFilter() string {
	return n.MachineName
}

func (n *monitorMachineNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
}

type taxReportNotification struct {
	types.NotificationBaseImpl
}

func (n *taxReportNotification) GetEmailAttachment() *types.EmailAttachment {
	tNow := time.Now()
	lastDay := time.Date(tNow.Year(), tNow.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstDay := lastDay.AddDate(0, -1, 0)

	q, err := url.ParseQuery(n.EventFilter)

	if err != nil {
		log.Warnf("Failed to parse rewards report eventfilter: %v", err)
		return nil
	}

	currency := q.Get("currency")

	validators := []uint64{}
	valSlice := strings.Split(q.Get("validators"), ",")
	if len(valSlice) > 0 {
		for _, val := range valSlice {
			v, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				continue
			}
			validators = append(validators, v)
		}
	} else {
		log.Warnf("Validators Not found in rewards report eventfilter")
		return nil
	}

	pdf := services.GetPdfReport(validators, currency, uint64(firstDay.Unix()), uint64(lastDay.Unix()))

	return &types.EmailAttachment{Attachment: pdf, Name: fmt.Sprintf("income_history_%v_%v.pdf", firstDay.Format("20060102"), lastDay.Format("20060102"))}
}

func (n *taxReportNotification) GetInfo(includeUrl bool) string {
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *taxReportNotification) GetTitle() string {
	return "Income Report"
}

func (n *taxReportNotification) GetEventFilter() string {
	return n.EventFilter
}

func (n *taxReportNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
}

type networkNotification struct {
	types.NotificationBaseImpl
}

func (n *networkNotification) GetInfo(includeUrl bool) string {
	generalPart := fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func (n *networkNotification) GetTitle() string {
	return "Beaconchain Network Issues"
}

func (n *networkNotification) GetInfoMarkdown() string {
	generalPart := fmt.Sprintf(`Network experienced finality issues ([view chart](https://%v/charts/network_liveness)).`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

type rocketpoolNotification struct {
	types.NotificationBaseImpl
	ExtraData string
}

func (n *rocketpoolNotification) GetInfo(includeUrl bool) string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `A new reward round has started. You can now claim your rewards from the previous round.`
	case types.RocketpoolCollateralMaxReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.RocketpoolCollateralMinReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.SyncCommitteeSoon:
		return getSyncCommitteeSoonInfo(map[types.EventFilter]types.Notification{
			types.EventFilter(n.EventFilter): n,
		})
	}

	return ""
}

func (n *rocketpoolNotification) GetTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return `Rocketpool Commission`
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `Rocketpool Claim Available`
	case types.RocketpoolCollateralMaxReached:
		return `Rocketpool Max Collateral`
	case types.RocketpoolCollateralMinReached:
		return `Rocketpool Min Collateral`
	case types.SyncCommitteeSoon:
		return `Sync Committee Duty`
	}
	return ""
}

func (n *rocketpoolNotification) GetInfoMarkdown() string {
	return n.GetInfo(false)
}

type BigFloat big.Float

func (b *BigFloat) Value() (driver.Value, error) {
	if b != nil {
		return (*big.Float)(b).String(), nil
	}
	return nil, nil
}

func (b *BigFloat) Scan(value interface{}) error {
	if value == nil {
		return errors.New("can not cast nil to BigFloat")
	}

	switch t := value.(type) {
	case float64:
		(*big.Float)(b).SetFloat64(value.(float64))
	case []uint8:
		_, ok := (*big.Float)(b).SetString(string(value.([]uint8)))
		if !ok {
			return fmt.Errorf("failed to load value to []uint8: %v", value)
		}
	case string:
		_, ok := (*big.Float)(b).SetString(value.(string))
		if !ok {
			return fmt.Errorf("failed to load value to []uint8: %v", value)
		}
	default:
		return fmt.Errorf("could not scan type %T into BigFloat", t)
	}

	return nil
}

func (b *BigFloat) bigFloat() *big.Float {
	return (*big.Float)(b)
}
func bigFloat(x float64) *big.Float {
	return new(big.Float).SetFloat64(x)
}

type WebhookQueue struct {
	NotificationID uint64         `db:"id"`
	Url            string         `db:"url"`
	Retries        uint64         `db:"retries"`
	LastSent       time.Time      `db:"last_retry"`
	Destination    sql.NullString `db:"destination"`
	Payload        []byte         `db:"payload"`
	LastTry        time.Time      `db:"last_try"`
}

func getUrlPart(validatorIndex uint64) string {
	return fmt.Sprintf(` For more information visit: <a href='https://%s/validator/%v'>https://%s/validator/%v</a>.`, utils.Config.Frontend.SiteDomain, validatorIndex, utils.Config.Frontend.SiteDomain, validatorIndex)
}
