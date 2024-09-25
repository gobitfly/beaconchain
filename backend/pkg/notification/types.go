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

type validatorProposalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slot           uint64
	Status         uint64 // * Can be 0 = scheduled, 1 executed, 2 missed */
	Reward         float64
}

func (n *validatorProposalNotification) GetInfo(format types.NotificationFormat) string {
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

func (n *validatorProposalNotification) GetLegacyInfo() string {
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
	default:
		return "-"
	}
}

func (n *validatorProposalNotification) GetLegacyTitle() string {
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

type validatorIsOfflineNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	IsOffline      bool
}

// Overwrite specific methods
func (n *validatorIsOfflineNotification) GetInfo(format types.NotificationFormat) string {
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

func (n *validatorIsOfflineNotification) GetTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

func (n *validatorIsOfflineNotification) GetLegacyInfo() string {
	if n.IsOffline {
		return fmt.Sprintf(`Validator %v is offline since epoch %s.`, n.ValidatorIndex, n.LatestState)
	} else {
		return fmt.Sprintf(`Validator %v is back online since epoch %v.`, n.ValidatorIndex, n.Epoch)
	}
}

func (n *validatorIsOfflineNotification) GetLegacyTitle() string {
	if n.IsOffline {
		return "Validator is Offline"
	} else {
		return "Validator Back Online"
	}
}

type validatorAttestationNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex     uint64
	ValidatorPublicKey string
	Status             uint64 // * Can be 0 = scheduled | missed, 1 executed
}

func (n *validatorAttestationNotification) GetInfo(format types.NotificationFormat) string {
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

func (n *validatorAttestationNotification) GetTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

func (n *validatorAttestationNotification) GetLegacyInfo() string {
	var generalPart = ""
	switch n.Status {
	case 0:
		generalPart = fmt.Sprintf(`Validator %v missed an attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
	case 1:
		generalPart = fmt.Sprintf(`Validator %v submitted a successful attestation in epoch %v.`, n.ValidatorIndex, n.Epoch)
	}
	return generalPart
}

func (n *validatorAttestationNotification) GetLegacyTitle() string {
	switch n.Status {
	case 0:
		return "Attestation Missed"
	case 1:
		return "Attestation Submitted"
	}
	return "-"
}

type validatorGotSlashedNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Slasher        uint64
	Reason         string
}

func (n *validatorGotSlashedNotification) GetInfo(format types.NotificationFormat) string {
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)
	vali := formatValidatorLink(format, n.ValidatorIndex)
	epoch := formatEpochLink(format, n.Epoch)
	slasher := formatValidatorLink(format, n.Slasher)

	return fmt.Sprintf(`Validator %v%v has been slashed at epoch %v by validator %v for %s.`, vali, dashboardAndGroupInfo, epoch, slasher, n.Reason)
}

func (n *validatorGotSlashedNotification) GetTitle() string {
	return "Validator got Slashed"
}

func (n *validatorGotSlashedNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`Validator %v has been slashed at epoch %v by validator %v for %s.`, n.ValidatorIndex, n.Epoch, n.Slasher, n.Reason)
	return generalPart
}

func (n *validatorGotSlashedNotification) GetLegacyTitle() string {
	return "Validator got Slashed"
}

type validatorWithdrawalNotification struct {
	types.NotificationBaseImpl

	ValidatorIndex uint64
	Epoch          uint64
	Slot           uint64
	Amount         uint64
	Address        []byte
}

func (n *validatorWithdrawalNotification) GetInfo(format types.NotificationFormat) string {
	dashboardAndGroupInfo := formatDashboardAndGroupLink(format, n)
	vali := formatValidatorLink(format, n.ValidatorIndex)
	amount := utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false)
	generalPart := fmt.Sprintf(`An automatic withdrawal of %s has been processed for validator %s%s.`, amount, vali, dashboardAndGroupInfo)

	return generalPart
}

func (n *validatorWithdrawalNotification) GetTitle() string {
	return "Withdrawal Processed"
}

func (n *validatorWithdrawalNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`An automatic withdrawal of %v has been processed for validator %v.`, utils.FormatClCurrencyString(n.Amount, utils.Config.Frontend.MainCurrency, 6, true, false, false), n.ValidatorIndex)
	return generalPart
}

func (n *validatorWithdrawalNotification) GetLegacyTitle() string {
	return "Withdrawal Processed"
}

type ethClientNotification struct {
	types.NotificationBaseImpl

	EthClient string
}

func (n *ethClientNotification) GetInfo(format types.NotificationFormat) string {

	switch format {
	case types.NotifciationFormatHtml:
		generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
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
	case types.NotifciationFormatText:
		return fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	case types.NotifciationFormatMarkdown:
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
	return ""
}

func (n *ethClientNotification) GetTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
}

func (n *ethClientNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`A new version for %s is available.`, n.EthClient)
	return generalPart
}

func (n *ethClientNotification) GetLegacyTitle() string {
	return fmt.Sprintf("New %s update", n.EthClient)
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

func (n *monitorMachineNotification) GetInfo(format types.NotificationFormat) string {
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

func (n *monitorMachineNotification) GetTitle() string {
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

func (n *monitorMachineNotification) GetLegacyInfo() string {
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

func (n *monitorMachineNotification) GetLegacyTitle() string {
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

func (n *monitorMachineNotification) GetEventFilter() string {
	return n.MachineName
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

func (n *taxReportNotification) GetInfo(format types.NotificationFormat) string {
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *taxReportNotification) GetTitle() string {
	return "Income Report"
}

func (n *taxReportNotification) GetLegacyInfo() string {
	generalPart := `Please find attached the income history of your selected validators.`
	return generalPart
}

func (n *taxReportNotification) GetLegacyTitle() string {
	return "Income Report"
}

func (n *taxReportNotification) GetEventFilter() string {
	return n.EventFilter
}

type networkNotification struct {
	types.NotificationBaseImpl
}

func (n *networkNotification) GetInfo(format types.NotificationFormat) string {
	switch format {
	case types.NotifciationFormatHtml, types.NotifciationFormatText:
		return fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	case types.NotifciationFormatMarkdown:
		return fmt.Sprintf(`Network experienced finality issues. [Learn more](https://%v/charts/network_liveness)`, utils.Config.Frontend.SiteDomain)
	}
	return ""
}

func (n *networkNotification) GetTitle() string {
	return "Beaconchain Network Issues"
}

func (n *networkNotification) GetLegacyInfo() string {
	generalPart := fmt.Sprintf(`Network experienced finality issues. Learn more at https://%v/charts/network_liveness`, utils.Config.Frontend.SiteDomain)
	return generalPart
}

func (n *networkNotification) GetLegacyTitle() string {
	return "Beaconchain Network Issues"
}

type rocketpoolNotification struct {
	types.NotificationBaseImpl
	ExtraData string
}

func (n *rocketpoolNotification) GetInfo(format types.NotificationFormat) string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `A new reward round has started. You can now claim your rewards from the previous round.`
	case types.RocketpoolCollateralMaxReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.RocketpoolCollateralMinReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
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
	}
	return ""
}

func (n *rocketpoolNotification) GetLegacyInfo() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return fmt.Sprintf(`The current RPL commission rate of %v has reached your configured threshold.`, n.ExtraData)
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `A new reward round has started. You can now claim your rewards from the previous round.`
	case types.RocketpoolCollateralMaxReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	case types.RocketpoolCollateralMinReached:
		return fmt.Sprintf(`Your RPL collateral has reached your configured threshold at %v%%.`, n.ExtraData)
	}

	return ""
}

func (n *rocketpoolNotification) GetLegacyTitle() string {
	switch n.EventName {
	case types.RocketpoolCommissionThresholdEventName:
		return `Rocketpool Commission`
	case types.RocketpoolNewClaimRoundStartedEventName:
		return `Rocketpool Claim Available`
	case types.RocketpoolCollateralMaxReached:
		return `Rocketpool Max Collateral`
	case types.RocketpoolCollateralMinReached:
		return `Rocketpool Min Collateral`
	}
	return ""
}

type syncCommitteeSoonNotification struct {
	types.NotificationBaseImpl
	Validator  uint64
	StartEpoch uint64
	EndEpoch   uint64
}

func (n *syncCommitteeSoonNotification) GetInfo(format types.NotificationFormat) string {
	return getSyncCommitteeSoonInfo(format, map[types.EventFilter]types.Notification{
		types.EventFilter(n.EventFilter): n,
	})
}

func (n *syncCommitteeSoonNotification) GetTitle() string {
	return `Sync Committee Duty`
}

func (n *syncCommitteeSoonNotification) GetLegacyInfo() string {
	return getSyncCommitteeSoonLegacyInfo(map[types.EventFilter]types.Notification{
		types.EventFilter(n.EventFilter): n,
	})
}

func (n *syncCommitteeSoonNotification) GetLegacyTitle() string {
	return `Sync Committee Duty`
}

func getSyncCommitteeSoonLegacyInfo(ns map[types.EventFilter]types.Notification) string {
	validators := []string{}
	var startEpoch, endEpoch string
	var inTime time.Duration

	i := 0
	for _, n := range ns {
		n, ok := n.(*syncCommitteeSoonNotification)
		if !ok {
			log.Error(nil, "Sync committee notification not of type syncCommitteeSoonNotification", 0)
			return ""
		}

		validators = append(validators, fmt.Sprintf("%d", n.Validator))
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
