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

func formatValidatorLink(format types.NotificationFormat, validatorIndex interface{}) string {
	switch format {
	case types.NotifciationFormatHtml:
		return fmt.Sprintf(`<a href="https://%s/validator/%v">%v</a>`, utils.Config.Frontend.SiteDomain, validatorIndex, validatorIndex)
	case types.NotifciationFormatText:
		return fmt.Sprintf(`%v`, validatorIndex)
	case types.NotifciationFormatMarkdown:
		return fmt.Sprintf(`[%d](https://%s/validator/%v)`, validatorIndex, utils.Config.Frontend.SiteDomain, validatorIndex)
	}
	return ""
}
func formatEpochLink(format types.NotificationFormat, epoch interface{}) string {
	switch format {
	case types.NotifciationFormatHtml:
		return fmt.Sprintf(`<a href="https://%s/epoch/%v">%v</a>`, utils.Config.Frontend.SiteDomain, epoch, epoch)
	case types.NotifciationFormatText:
		return fmt.Sprintf(`%v`, epoch)
	case types.NotifciationFormatMarkdown:
		return fmt.Sprintf(`[%v](https://%s/epoch/%v)`, epoch, utils.Config.Frontend.SiteDomain, epoch)
	}
	return ""
}
func formatSlotLink(format types.NotificationFormat, slot interface{}) string {
	switch format {
	case types.NotifciationFormatHtml:
		return fmt.Sprintf(`<a href="https://%s/slot/%v">%v</a>`, utils.Config.Frontend.SiteDomain, slot, slot)
	case types.NotifciationFormatText:
		return fmt.Sprintf(`%v`, slot)
	case types.NotifciationFormatMarkdown:
		return fmt.Sprintf(`[%v](https://%s/slot/%v)`, slot, utils.Config.Frontend.SiteDomain, slot)
	}
	return ""
}

func formatDashboardAndGroupLink(format types.NotificationFormat, n types.Notification) string {
	dashboardAndGroupInfo := ""
	if n.GetDashboardId() != nil {
		switch format {
		case types.NotifciationFormatHtml:
			dashboardAndGroupInfo = fmt.Sprintf(` of Group <b>%[2]v</b> in Dashboard <a href="https://%[1]v/dashboard/%[4]v">%[3]v</a>`, utils.Config.Frontend.SiteDomain, n.GetDashboardGroupName(), n.GetDashboardName(), *n.GetDashboardId())
		case types.NotifciationFormatText:
			dashboardAndGroupInfo = fmt.Sprintf(` of Group %[1]v in Dashboard %[2]v`, n.GetDashboardGroupName(), n.GetDashboardName())
		case types.NotifciationFormatMarkdown:
			dashboardAndGroupInfo = fmt.Sprintf(` of Group **%[1]v** in Dashboard [%[2]v](https://%[3]v/dashboard/%[4]v)`, n.GetDashboardGroupName(), n.GetDashboardName(), utils.Config.Frontend.SiteDomain, *n.GetDashboardId())
		}
	}
	return dashboardAndGroupInfo
}

type ValidatorProposalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slot           uint64
	Block          uint64
	Status         uint64 // * Can be 0 = scheduled, 1 executed, 2 missed */
	Reward         float64
}

func (n *ValidatorProposalNotification) GetEntitiyId() string {
	return fmt.Sprintf("%d", n.ValidatorIndex)
}

func (n *ValidatorProposalNotification) GetInfo(format types.NotificationFormat) string {
	vali := formatValidatorLink(format, n.ValidatorIndex)
	slot := formatSlotLink(format, n.Slot)
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)

	switch n.Status {
	case 0:
		return fmt.Sprintf(`New scheduled block proposal at slot %s for Validator %s%s.`, slot, vali, dashboardAndGroupInfo)
	case 1:
		return fmt.Sprintf(`Validator %s%s proposed block at slot %s with %v %v execution reward.`, vali, dashboardAndGroupInfo, slot, n.Reward, utils.Config.Frontend.ElCurrency)
	case 2:
		return fmt.Sprintf(`Validator %s%s missed a block proposal at slot %s.`, vali, dashboardAndGroupInfo, slot)
	case 3:
		return fmt.Sprintf(`Validator %s%s had an orphaned block proposal at slot %s.`, vali, dashboardAndGroupInfo, slot)
	default:
		return "-"
	}
}

func (n *ValidatorProposalNotification) GetLegacyInfo() string {
	var generalPart, suffix string
	vali := strconv.FormatUint(n.ValidatorIndex, 10)
	slot := strconv.FormatUint(n.Slot, 10)

	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`New scheduled block proposal at slot %s for Validator %s.`, slot, vali)
	case 1:
		generalPart = fmt.Sprintf(`Validator %s proposed block at slot %s with %v %v execution reward.`, vali, slot, n.Reward, utils.Config.Frontend.ElCurrency)
	case 2:
		generalPart = fmt.Sprintf(`Validator %s missed a block proposal at slot %s.`, vali, slot)
	case 3:
		generalPart = fmt.Sprintf(`Validator %s had an orphaned block proposal at slot %s.`, vali, slot)
	}
	return generalPart + suffix
}

func (n *ValidatorProposalNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Block Proposal Scheduled"
	case 1:
		return "New Block Proposal"
	case 2:
		return "Block Proposal Missed"
	case 3:
		return "Block Proposal Missed (Orphaned)"
	default:
		return "-"
	}
}

func (n *ValidatorProposalNotification) GetLegacyTitle() string {
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

type ValidatorIsOfflineNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	IsOffline      bool
}

func (n *ValidatorIsOfflineNotification) GetEntitiyId() string {
	return fmt.Sprintf("%d", n.ValidatorIndex)
}

// Overwrite specific methods
func (n *ValidatorIsOfflineNotification) GetInfo(format types.NotificationFormat) string {
	vali := formatValidatorLink(format, n.ValidatorIndex)
	epoch := ""
	if n.IsOffline {
		epoch = formatEpochLink(format, n.LatestState)
	} else {
		epoch = formatEpochLink(format, n.Epoch)
	}
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)

	if n.IsOffline {
		return fmt.Sprintf(`Validator %v%v is offline since epoch %s.`, vali, dashboardAndGroupInfo, epoch)
	} else {
		return fmt.Sprintf(`Validator %v%v is back online since epoch %v.`, vali, dashboardAndGroupInfo, epoch)
	}
}

func (n *ValidatorIsOfflineNotification) GetTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

func (n *ValidatorIsOfflineNotification) GetLegacyInfo() string {
	if n.IsOffline {
		return fmt.Sprintf(`Validator %v is offline since epoch %s.`, n.ValidatorIndex, n.LatestState)
	} else {
		return fmt.Sprintf(`Validator %v is back online since epoch %v.`, n.ValidatorIndex, n.Epoch)
	}
}

func (n *ValidatorIsOfflineNotification) GetLegacyTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

// type validatorGroupIsOfflineNotification struct {
// 	types.NotificationBaseImpl

// 	IsOffline bool
// }

// func (n *validatorGroupIsOfflineNotification) GetEntitiyId() string {
// 	return fmt.Sprintf("%s - %s", n.GetDashboardName(), n.GetDashboardGroupName())
// }

// // Overwrite specific methods
// func (n *validatorGroupIsOfflineNotification) GetInfo(format types.NotificationFormat) string {
// 	epoch := ""
// 	if n.IsOffline {
// 		epoch = formatEpochLink(format, n.LatestState)
// 	} else {
// 		epoch = formatEpochLink(format, n.Epoch)
// 	}

// 	if n.IsOffline {
// 		return fmt.Sprintf(`Group %s is offline since epoch %s.`, n.DashboardGroupName, epoch)
// 	} else {
// 		return fmt.Sprintf(`Group %s is back online since epoch %v.`, n.DashboardGroupName, epoch)
// 	}
// }

// func (n *validatorGroupIsOfflineNotification) GetTitle() string {
// 	if n.IsOffline {
// 		return "Group is offline"
// 	} else {
// 		return "Group is back online"
// 	}
// }

// func (n *validatorGroupIsOfflineNotification) GetLegacyInfo() string {
// 	if n.IsOffline {
// 		return fmt.Sprintf(`Group %s is offline since epoch %s.`, n.DashboardGroupName, n.LatestState)
// 	} else {
// 		return fmt.Sprintf(`Group %s is back online since epoch %v.`, n.DashboardGroupName, n.Epoch)
// 	}
// }

// func (n *validatorGroupIsOfflineNotification) GetLegacyTitle() string {
// 	if n.IsOffline {
// 		return "Group is offline"
// 	} else {
// 		return "Group is back online"
// 	}
// }

type ValidatorAttestationNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex     uint64
	ValidatorPublicKey string
	Status             uint64 // * Can be 0 = scheduled | missed, 1 executed
}

func (n *ValidatorAttestationNotification) GetEntitiyId() string {
	return fmt.Sprintf("%d", n.ValidatorIndex)
}

func (n *ValidatorAttestationNotification) GetInfo(format types.NotificationFormat) string {
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)
	vali := formatValidatorLink(format, n.ValidatorIndex)
	epoch := formatEpochLink(format, n.Epoch)

	switch format {
	case types.NotifciationFormatHtml:
		switch n.Status {
		case 0:
			return fmt.Sprintf(`Validator %s%s missed an attestation in epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		case 1:
			return fmt.Sprintf(`Validator %s%s submitted a successful attestation for epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		default:
			return "-"
		}
	case types.NotifciationFormatText:
		switch n.Status {
		case 0:
			return fmt.Sprintf(`Validator %s%s missed an attestation in epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		case 1:
			return fmt.Sprintf(`Validator %s%s submitted a successful attestation for epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		default:
			return "-"
		}
	case types.NotifciationFormatMarkdown:
		switch n.Status {
		case 0:
			return fmt.Sprintf(`Validator %s%s missed an attestation in epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		case 1:
			return fmt.Sprintf(`Validator %s%s submitted a successful attestation for epoch %s.`, vali, dashboardAndGroupInfo, epoch)
		default:
			return "-"
		}
	}
	return ""
}

func (n *ValidatorAttestationNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

func (n *ValidatorAttestationNotification) GetLegacyInfo() string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`Validator %v missed an attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
	case 1:
		generalPart = fmt.Sprintf(`Validator %v submitted a successful attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
	}
	return generalPart
}

func (n *ValidatorAttestationNotification) GetLegacyTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

type ValidatorGotSlashedNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slasher        uint64
	Reason         string
}

func (n *ValidatorGotSlashedNotification) GetEntitiyId() string {
	return fmt.Sprintf("%d", n.ValidatorIndex)
}

func (n *ValidatorGotSlashedNotification) GetInfo(format types.NotificationFormat) string {
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)
	vali := formatValidatorLink(format, n.ValidatorIndex)
	epoch := formatEpochLink(format, n.Epoch)
	slasher := formatValidatorLink(format, n.Slasher)

	return fmt.Sprintf(`Validator %v%v has been slashed at epoch %v by validator %v for %s.`, vali, dashboardAndGroupInfo, epoch, slasher, n.Reason)
}

func (n *ValidatorGotSlashedNotification) GetTitle() string {
	return "Validator got Slashed"
}

func (n *ValidatorGotSlashedNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`Validator %v has been slashed at epoch %v by validator %v for %s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason)
	return generalPart
}

func (n *ValidatorGotSlashedNotification) GetLegacyTitle() string {
	return "Validator got Slashed"
}

type ValidatorWithdrawalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Epoch          uint64
	Slot           uint64
	Amount         uint64
	Address        []byte
}

func (n *ValidatorWithdrawalNotification) GetEntitiyId() string {
	return fmt.Sprintf("%v", n.ValidatorIndex)
}

func (n *ValidatorWithdrawalNotification) GetInfo(format types.NotificationFormat) string {
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)
	vali := formatValidatorLink(format, n.ValidatorIndex)
	amount := utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false)
	generalPart := fmt.Sprintf(`An automatic withdrawal of %s has been processed for validator %s%s.`, amount, vali, dashboardAndGroupInfo)

	return generalPart
}

func (n *ValidatorWithdrawalNotification) GetTitle() string {
	return "Withdrawal Processed"
}

func (n *ValidatorWithdrawalNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`An automatic withdrawal of %v has been processed for validator %v.`, utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false), n.ValidatorIndex)
	return generalPart
}

func (n *ValidatorWithdrawalNotification) GetLegacyTitle() string {
	return "Withdrawal Processed"
}

type EthClientNotification struct {
	types.NotificationBaseImpl

	EthClient string
}

func (n *EthClientNotification) GetEntitiyId() string {
	return n.EthClient
}

func (n *EthClientNotification) GetInfo(format types.NotificationFormat) string {
	clientUrls := map[string]string{
		"Geth":       "https://github.com/ethereum/go-ethereum/releases",
		"Nethermind": "https://github.com/NethermindEth/nethermind/releases",
		"Teku":       "https://github.com/ConsenSys/teku/releases",
		"Prysm":      "https://github.com/prysmaticlabs/prysm/releases",
		"Nimbus":     "https://github.com/status-im/nimbus-eth2/releases",
		"Lighthouse": "https://github.com/sigp/lighthouse/releases",
		"Erigon":     "https://github.com/erigontech/erigon/releases",
		"Rocketpool": "https://github.com/rocket-pool/smartnode-install/releases",
		"MEV-Boost":  "https://github.com/flashbots/mev-boost/releases",
		"Lodestar":   "https://github.com/chainsafe/lodestar/releases",
	}
	defaultUrl := "https://beaconcha.in/ethClients"

	switch format {
	case types.NotifciationFormatHtml:
		generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
		url := clientUrls[n.EthClient]
		if url == "" {
			url = defaultUrl
		}
		return generalPart + " " + url
	case types.NotifciationFormatText:
		return fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	case types.NotifciationFormatMarkdown:
		url := clientUrls[n.EthClient]
		if url == "" {
			url = defaultUrl
		}

		generalPart := fmt.Sprintf(`A new version for [%s](%s) is available.`, n.EthClient, url)

		return generalPart
	}
	return ""
}

func (n *EthClientNotification) GetTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
}

func (n *EthClientNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	return generalPart
}

func (n *EthClientNotification) GetLegacyTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
}

type MachineEvents struct {
	SubscriptionID  uint64         `db:"id"`
	UserID          types.UserId   `db:"user_id"`
	MachineName     string         `db:"machine"`
	UnsubscribeHash sql.NullString `db:"unsubscribe_hash"`
	EventThreshold  float64        `db:"event_threshold"`
}

type MonitorMachineNotification struct {
	types.NotificationBaseImpl

	MachineName    string
	EventThreshold float64
}

func (n *MonitorMachineNotification) GetEntitiyId() string {
	return n.MachineName
}

func (n *MonitorMachineNotification) GetInfo(format types.NotificationFormat) string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return fmt.Sprintf(`Your staking machine "%v" is running low on storage space.`, n.MachineName)
	case types.MonitoringMachineOfflineEventName:
		return fmt.Sprintf(`Your staking machine "%v" might be offline. It has not been seen for a couple minutes now.`, n.MachineName)
	case types.MonitoringMachineCpuLoadEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured CPU usage threshold.`, n.MachineName)
	case types.MonitoringMachineMemoryUsageEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured RAM threshold.`, n.MachineName)
	}
	return ""
}

func (n *MonitorMachineNotification) GetTitle() string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return "Storage Warning"
	case types.MonitoringMachineOfflineEventName:
		return "Staking Machine Offline"
	case types.MonitoringMachineCpuLoadEventName:
		return "High CPU Load"
	case types.MonitoringMachineMemoryUsageEventName:
		return "Memory Warning"
	}
	return ""
}

func (n *MonitorMachineNotification) GetLegacyInfo() string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return fmt.Sprintf(`Your staking machine "%v" is running low on storage space.`, n.MachineName)
	case types.MonitoringMachineOfflineEventName:
		return fmt.Sprintf(`Your staking machine "%v" might be offline. It has not been seen for a couple minutes now.`, n.MachineName)
	case types.MonitoringMachineCpuLoadEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured CPU usage threshold.`, n.MachineName)
	case types.MonitoringMachineMemoryUsageEventName:
		return fmt.Sprintf(`Your staking machine "%v" has reached your configured RAM threshold.`, n.MachineName)
	}
	return ""
}

func (n *MonitorMachineNotification) GetLegacyTitle() string {
	switch n.EventName {
	case types.MonitoringMachineDiskAlmostFullEventName:
		return "Storage Warning"
	case types.MonitoringMachineOfflineEventName:
		return "Staking Machine Offline"
	case types.MonitoringMachineCpuLoadEventName:
		return "High CPU Load"
	case types.MonitoringMachineMemoryUsageEventName:
		return "Memory Warning"
	}
	return ""
}

func (n *MonitorMachineNotification) GetEventFilter() string {
	return n.MachineName
}

type TaxReportNotification struct {
	types.NotificationBaseImpl
}

func (n *TaxReportNotification) GetEntitiyId() string {
	return ""
}

func (n *TaxReportNotification) GetEmailAttachment() *types.EmailAttachment {
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

func (n *TaxReportNotification) GetInfo(format types.NotificationFormat) string {
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *TaxReportNotification) GetTitle() string {
	return "Income Report"
}

func (n *TaxReportNotification) GetLegacyInfo() string {
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *TaxReportNotification) GetLegacyTitle() string {
	return "Income Report"
}

func (n *TaxReportNotification) GetEventFilter() string {
	return n.EventFilter
}

type NetworkNotification struct {
	types.NotificationBaseImpl
}

func (n *NetworkNotification) GetEntitiyId() string {
	return ""
}

func (n *NetworkNotification) GetInfo(format types.NotificationFormat) string {
	switch format {
	case types.NotifciationFormatHtml, types.NotifciationFormatText:
		return fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	case types.NotifciationFormatMarkdown:
		return fmt.Sprintf(`Network experienced finality issues. [Learn more](https://%v/charts/network_liveness)`, utils.Config.Frontend.SiteDomain)
	}
	return ""
}

func (n *NetworkNotification) GetTitle() string {
	return "Beaconchain Network Issues"
}

func (n *NetworkNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func (n *NetworkNotification) GetLegacyTitle() string {
	return "Beaconchain Network Issues"
}

type RocketpoolNotification struct {
	types.NotificationBaseImpl
	ExtraData string
}

func (n *RocketpoolNotification) GetEntitiyId() string {
	return ""
}

func (n *RocketpoolNotification) GetInfo(format types.NotificationFormat) string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `A new reward round has started. You can now claim your rewards from the previous round.`
	case types.RocketpoolCollateralMaxReachedEventName:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.RocketpoolCollateralMinReachedEventName:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	}

	return ""
}

func (n *RocketpoolNotification) GetTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return `Rocketpool Commission`
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `Rocketpool Claim Available`
	case types.RocketpoolCollateralMaxReachedEventName:
		return `Rocketpool Max Collateral`
	case types.RocketpoolCollateralMinReachedEventName:
		return `Rocketpool Min Collateral`
	}
	return ""
}

func (n *RocketpoolNotification) GetLegacyInfo() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `A new reward round has started. You can now claim your rewards from the previous round.`
	case types.RocketpoolCollateralMaxReachedEventName:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.RocketpoolCollateralMinReachedEventName:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	}

	return ""
}

func (n *RocketpoolNotification) GetLegacyTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return `Rocketpool Commission`
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `Rocketpool Claim Available`
	case types.RocketpoolCollateralMaxReachedEventName:
		return `Rocketpool Max Collateral`
	case types.RocketpoolCollateralMinReachedEventName:
		return `Rocketpool Min Collateral`
	}
	return ""
}

type SyncCommitteeSoonNotification struct {
	types.NotificationBaseImpl
	ValidatorIndex uint64
	StartEpoch     uint64
	EndEpoch       uint64
}

func (n *SyncCommitteeSoonNotification) GetEntitiyId() string {
	return fmt.Sprintf("%d", n.ValidatorIndex)
}

func (n *SyncCommitteeSoonNotification) GetInfo(format types.NotificationFormat) string {
	return getSyncCommitteeSoonInfo(format, map[types.EventFilter]types.Notification{
		types.EventFilter(n.EventFilter): n,
	})
}

func (n *SyncCommitteeSoonNotification) GetTitle() string {
	return `Sync Committee Duty`
}

func (n *SyncCommitteeSoonNotification) GetLegacyInfo() string {
	return getSyncCommitteeSoonLegacyInfo(map[types.EventFilter]types.Notification{
		types.EventFilter(n.EventFilter): n,
	})
}

func (n *SyncCommitteeSoonNotification) GetLegacyTitle() string {
	return `Sync Committee Duty`
}

func getSyncCommitteeSoonLegacyInfo(ns map[types.EventFilter]types.Notification) string {
	validators := []string{}
	var startEpoch, endEpoch string
	var inTime time.Duration

	i := 0
	for _, n := range ns {
		n, ok := n.(*SyncCommitteeSoonNotification)
		if !ok {
			log.Error(nil, "Sync committee notification not of type syncCommitteeSoonNotification", 0)
			return ""
		}

		validators = append(validators, fmt.Sprintf("%d", n.ValidatorIndex))
		if i == 0 {
			// startEpoch, endEpoch and inTime must be the same for all validators
			startEpoch = fmt.Sprintf("%d", n.StartEpoch)
			endEpoch = fmt.Sprintf("%d", n.EndEpoch)

			syncStartEpoch, err := strconv.ParseUint(startEpoch, 10, 64)
			if err != nil {
				inTime = utils.Day
			} else {
				inTime = time.Until(utils.EpochToTime(syncStartEpoch))
			}
			inTime = inTime.Round(time.Second)
		}
		i++
	}

	if len(validators) > 0 {
		validatorsInfo := ""
		if len(validators) == 1 {
			validatorsInfo = fmt.Sprintf(`Your validator %s has been elected to be part of the next sync committee.`, validators[0])
		} else {
			validatorsText := ""
			for i, validator := range validators {
				if i < len(validators)-1 {
					validatorsText += fmt.Sprintf("%s, ", validator)
				} else {
					validatorsText += fmt.Sprintf("and %s", validator)
				}
			}
			validatorsInfo = fmt.Sprintf(`Your validators %s have been elected to be part of the next sync committee.`, validatorsText)
		}
		return fmt.Sprintf(`%s The additional duties start at epoch %s, which is in %s and will last for about a day until epoch %s.`, validatorsInfo, startEpoch, inTime, endEpoch)
	}

	return ""
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
