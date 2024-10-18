package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"math/big"
	"strings"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EventName string
type EventFilter string

type NotificationsPerUserId map[UserId]NotificationsPerDashboard
type NotificationsPerDashboard map[DashboardId]NotificationsPerDashboardGroup
type NotificationsPerDashboardGroup map[DashboardGroupId]NotificationsPerEventName
type NotificationsPerEventName map[EventName]NotificationsPerEventFilter
type NotificationsPerEventFilter map[EventFilter]Notification

func (npui NotificationsPerUserId) AddNotification(n Notification) {
	if n.GetUserId() == 0 {
		log.Fatal(fmt.Errorf("Notification user id is 0"), fmt.Sprintf("Notification: %v", n), 1)
	}
	if n.GetEventName() == "" {
		log.Fatal(fmt.Errorf("Notification event name is empty"), fmt.Sprintf("Notification: %v", n), 1)
	}

	dashboardId := DashboardId(0)
	dashboardGroupId := DashboardGroupId(0)
	if n.GetDashboardId() != nil {
		dashboardId = DashboardId(*n.GetDashboardId())
		dashboardGroupId = DashboardGroupId(*n.GetDashboardGroupId())
	}

	// next check is disabled as there are events that do not require a filter (rocketpool, network events)
	// if n.GetEventFilter() == "" {
	// 	log.Fatal(fmt.Errorf("Notification event filter is empty"), fmt.Sprintf("Notification: %v", n), 0)
	// }

	if _, ok := npui[n.GetUserId()]; !ok {
		npui[n.GetUserId()] = make(NotificationsPerDashboard)
	}
	if _, ok := npui[n.GetUserId()][dashboardId]; !ok {
		npui[n.GetUserId()][dashboardId] = make(NotificationsPerDashboardGroup)
	}
	if _, ok := npui[n.GetUserId()][dashboardId][dashboardGroupId]; !ok {
		npui[n.GetUserId()][dashboardId][dashboardGroupId] = make(NotificationsPerEventName)
	}
	if _, ok := npui[n.GetUserId()][dashboardId][dashboardGroupId][n.GetEventName()]; !ok {
		npui[n.GetUserId()][dashboardId][dashboardGroupId][n.GetEventName()] = make(NotificationsPerEventFilter)
	}
	npui[n.GetUserId()][dashboardId][dashboardGroupId][n.GetEventName()][EventFilter(n.GetEventFilter())] = n
}

const (
	ValidatorMissedProposalEventName   EventName = "validator_proposal_missed"
	ValidatorExecutedProposalEventName EventName = "validator_proposal_submitted"

	ValidatorDidSlashEventName       EventName = "validator_did_slash"
	ValidatorGroupIsOfflineEventName EventName = "validator_group_is_offline"

	ValidatorReceivedDepositEventName               EventName = "validator_received_deposit"
	NetworkSlashingEventName                        EventName = "network_slashing"
	NetworkValidatorActivationQueueFullEventName    EventName = "network_validator_activation_queue_full"
	NetworkValidatorActivationQueueNotFullEventName EventName = "network_validator_activation_queue_not_full"
	NetworkValidatorExitQueueFullEventName          EventName = "network_validator_exit_queue_full"
	NetworkValidatorExitQueueNotFullEventName       EventName = "network_validator_exit_queue_not_full"
	NetworkLivenessIncreasedEventName               EventName = "network_liveness_increased"
	TaxReportEventName                              EventName = "user_tax_report"
	SyncCommitteeSoonEventName                      EventName = "validator_synccommittee_soon"
	//nolint:misspell
	RocketpoolCommissionThresholdEventName EventName = "rocketpool_commision_threshold"

	// Validator dashboard events
	ValidatorIsOfflineEventName             EventName = "validator_is_offline"
	ValidatorIsOnlineEventName              EventName = "validator_is_online"
	GroupIsOfflineEventName                 EventName = "group_is_offline"
	ValidatorMissedAttestationEventName     EventName = "validator_attestation_missed"
	ValidatorProposalEventName              EventName = "validator_proposal"
	ValidatorUpcomingProposalEventName      EventName = "validator_proposal_upcoming"
	SyncCommitteeSoon                       EventName = "validator_synccommittee_soon"
	ValidatorReceivedWithdrawalEventName    EventName = "validator_withdrawal"
	ValidatorGotSlashedEventName            EventName = "validator_got_slashed"
	RocketpoolCollateralMinReachedEventName EventName = "rocketpool_colleteral_min" //nolint:misspell
	RocketpoolCollateralMaxReachedEventName EventName = "rocketpool_colleteral_max" //nolint:misspell

	// Account dashboard events
	IncomingTransactionEventName  EventName = "incoming_transaction"
	OutgoingTransactionEventName  EventName = "outgoing_transaction"
	ERC20TokenTransferEventName   EventName = "erc20_token_transfer"   // #nosec G101
	ERC721TokenTransferEventName  EventName = "erc721_token_transfer"  // #nosec G101
	ERC1155TokenTransferEventName EventName = "erc1155_token_transfer" // #nosec G101

	// Machine events
	MonitoringMachineOfflineEventName        EventName = "monitoring_machine_offline"
	MonitoringMachineDiskAlmostFullEventName EventName = "monitoring_hdd_almostfull"
	MonitoringMachineCpuLoadEventName        EventName = "monitoring_cpu_load"
	MonitoringMachineMemoryUsageEventName    EventName = "monitoring_memory_usage"

	// Client events
	EthClientUpdateEventName EventName = "eth_client_update"

	// Network events
	RocketpoolNewClaimRoundStartedEventName    EventName = "rocketpool_new_claimround"
	NetworkGasAboveThresholdEventName          EventName = "network_gas_above_threshold"
	NetworkGasBelowThresholdEventName          EventName = "network_gas_below_threshold"
	NetworkParticipationRateThresholdEventName EventName = "network_participation_rate_threshold"
)

var EventSortOrder = []EventName{
	ValidatorGotSlashedEventName,
	ValidatorDidSlashEventName,
	ValidatorMissedProposalEventName,
	ValidatorExecutedProposalEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineMemoryUsageEventName,
	SyncCommitteeSoonEventName,
	ValidatorIsOfflineEventName,
	ValidatorIsOnlineEventName,
	ValidatorReceivedWithdrawalEventName,
	NetworkLivenessIncreasedEventName,
	EthClientUpdateEventName,
	TaxReportEventName,
	RocketpoolCommissionThresholdEventName,
	RocketpoolNewClaimRoundStartedEventName,
	RocketpoolCollateralMinReachedEventName,
	RocketpoolCollateralMaxReachedEventName,
	ValidatorMissedAttestationEventName,
}

var MachineEvents = []EventName{
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineMemoryUsageEventName,
}

var UserIndexEvents = []EventName{
	EthClientUpdateEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineMemoryUsageEventName,
}

var UserIndexEventsMap = map[EventName]struct{}{
	EthClientUpdateEventName:                 {},
	MonitoringMachineCpuLoadEventName:        {},
	MonitoringMachineOfflineEventName:        {},
	MonitoringMachineDiskAlmostFullEventName: {},
	MonitoringMachineMemoryUsageEventName:    {},
}

var MachineEventsMap = map[EventName]struct{}{
	MonitoringMachineCpuLoadEventName:        {},
	MonitoringMachineOfflineEventName:        {},
	MonitoringMachineDiskAlmostFullEventName: {},
	MonitoringMachineMemoryUsageEventName:    {},
}

var LegacyEventLabel map[EventName]string = map[EventName]string{
	ValidatorMissedProposalEventName:         "Your validator(s) missed a proposal",
	ValidatorExecutedProposalEventName:       "Your validator(s) submitted a proposal",
	ValidatorMissedAttestationEventName:      "Your validator(s) missed an attestation",
	ValidatorGotSlashedEventName:             "Your validator(s) got slashed",
	ValidatorDidSlashEventName:               "Your validator(s) slashed another validator",
	ValidatorIsOfflineEventName:              "Your validator(s) went offline",
	ValidatorIsOnlineEventName:               "Your validator(s) came back online",
	ValidatorReceivedWithdrawalEventName:     "A withdrawal was initiated for your validators",
	NetworkLivenessIncreasedEventName:        "The network is experiencing liveness issues",
	EthClientUpdateEventName:                 "An Ethereum client has a new update available",
	MonitoringMachineOfflineEventName:        "Your machine(s) might be offline",
	MonitoringMachineDiskAlmostFullEventName: "Your machine(s) disk space is running low",
	MonitoringMachineCpuLoadEventName:        "Your machine(s) has a high CPU load",
	MonitoringMachineMemoryUsageEventName:    "Your machine(s) has a high memory load",
	TaxReportEventName:                       "You have an available tax report",
	RocketpoolCommissionThresholdEventName:   "Your configured Rocket Pool commission threshold is reached",
	RocketpoolNewClaimRoundStartedEventName:  "Your Rocket Pool claim from last round is available",
	RocketpoolCollateralMinReachedEventName:  "You reached the Rocket Pool min RPL collateral",
	RocketpoolCollateralMaxReachedEventName:  "You reached the Rocket Pool max RPL collateral",
	SyncCommitteeSoonEventName:               "Your validator(s) will soon be part of the sync committee",
}

var EventLabel map[EventName]string = map[EventName]string{
	ValidatorMissedProposalEventName:         "Block proposal missed",
	ValidatorExecutedProposalEventName:       "Block proposal submitted",
	ValidatorMissedAttestationEventName:      "Attestation missed",
	ValidatorGotSlashedEventName:             "Validator slashed",
	ValidatorDidSlashEventName:               "Validator has slashed",
	ValidatorIsOfflineEventName:              "Validator offline",
	ValidatorIsOnlineEventName:               "Validator back online",
	ValidatorReceivedWithdrawalEventName:     "Withdrawal processed",
	NetworkLivenessIncreasedEventName:        "The network is experiencing liveness issues",
	EthClientUpdateEventName:                 "An Ethereum client has a new update available",
	MonitoringMachineOfflineEventName:        "Machine offline",
	MonitoringMachineDiskAlmostFullEventName: "Machine low disk space",
	MonitoringMachineCpuLoadEventName:        "Machine high CPU load",
	MonitoringMachineMemoryUsageEventName:    "Machine high memory load",
	TaxReportEventName:                       "Tax report available",
	RocketpoolCommissionThresholdEventName:   "Rocket pool commission threshold is reached",
	RocketpoolNewClaimRoundStartedEventName:  "Rocket pool claim from last round is available",
	RocketpoolCollateralMinReachedEventName:  "Rocket pool node min RPL collateral reached",
	RocketpoolCollateralMaxReachedEventName:  "Rocket pool node max RPL collateral reached",
	SyncCommitteeSoonEventName:               "Upcoming sync committee",
}

func IsUserIndexed(event EventName) bool {
	_, ok := UserIndexEventsMap[event]
	return ok
}

func IsMachineNotification(event EventName) bool {
	_, ok := MachineEventsMap[event]
	return ok
}

var EventNames = []EventName{
	ValidatorExecutedProposalEventName,
	ValidatorMissedProposalEventName,
	ValidatorMissedAttestationEventName,
	ValidatorGotSlashedEventName,
	ValidatorDidSlashEventName,
	ValidatorIsOfflineEventName,
	ValidatorIsOnlineEventName,
	ValidatorReceivedWithdrawalEventName,
	NetworkLivenessIncreasedEventName,
	EthClientUpdateEventName,
	MonitoringMachineOfflineEventName,
	MonitoringMachineDiskAlmostFullEventName,
	MonitoringMachineCpuLoadEventName,
	MonitoringMachineMemoryUsageEventName,
	TaxReportEventName,
	RocketpoolCommissionThresholdEventName,
	RocketpoolNewClaimRoundStartedEventName,
	RocketpoolCollateralMinReachedEventName,
	RocketpoolCollateralMaxReachedEventName,
	SyncCommitteeSoonEventName,
}

type EventNameDesc struct {
	Desc    string
	Event   EventName
	Info    template.HTML
	Warning template.HTML
}

type MachineMetricSystemUser struct {
	UserID                    UserId
	Machine                   string
	CurrentData               *MachineMetricSystem
	CurrentDataInsertTs       int64
	FiveMinuteOldData         *MachineMetricSystem
	FiveMinuteOldDataInsertTs int64
}

// this is the source of truth for the validator events that are supported by the user/notification page
var AddWatchlistEvents = []EventNameDesc{
	{
		Desc:  "Validator is Offline",
		Event: ValidatorIsOfflineEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation:<br><ul><li>Once you have been offline for 3 epochs</li><li>Every 32 Epochs (~3 hours) during your downtime</li><li>Once you are back online again</li></ul></div>" class="fas fa-question-circle"></i>`),
	},
	{
		Desc:  "Proposals missed",
		Event: ValidatorMissedProposalEventName,
	},
	{
		Desc:  "Proposals submitted",
		Event: ValidatorExecutedProposalEventName,
	},
	{
		Desc:  "Validator got slashed",
		Event: ValidatorGotSlashedEventName,
	},
	{
		Desc:  "Sync committee",
		Event: SyncCommitteeSoonEventName,
	},
	{
		Desc:    "Attestations missed",
		Event:   ValidatorMissedAttestationEventName,
		Warning: template.HTML(`<i data-toggle="tooltip" title="Will trigger every epoch (6.4 minutes) during downtime" class="fas fa-exclamation-circle text-warning"></i>`),
	},
	{
		Desc:  "Withdrawal processed",
		Event: ValidatorReceivedWithdrawalEventName,
		Info:  template.HTML(`<i data-toggle="tooltip" data-html="true" title="<div class='text-left'>Will trigger a notifcation when:<br><ul><li>A partial withdrawal is processed</li><li>Your validator exits and its full balance is withdrawn</li></ul> <div>Requires that your validator has 0x01 credentials</div></div>" class="fas fa-question-circle"></i>`),
	},
}

// this is the source of truth for the network events that are supported by the user/notification page
var NetworkNotificationEvents = []EventNameDesc{
	{
		Desc:  "Network Notifications",
		Event: NetworkLivenessIncreasedEventName,
	},
	// {
	// 	Desc:  "Slashing Notifications",
	// 	Event: NetworkSlashingEventName,
	// },
}

func GetDisplayableEventName(event EventName) string {
	return cases.Title(language.English).String(strings.ReplaceAll(string(event), "_", " "))
}

func EventNameFromString(event string) (EventName, error) {
	for _, en := range EventNames {
		if string(en) == event {
			return en, nil
		}
	}
	return "", errors.Errorf("Could not convert event to string. %v is not a known event type", event)
}

type Tag string

const (
	ValidatorTagsWatchlist Tag = "watchlist"
)

type NotificationFormat string

var NotifciationFormatHtml NotificationFormat = "html"
var NotifciationFormatText NotificationFormat = "text"
var NotifciationFormatMarkdown NotificationFormat = "markdown"

type Notification interface {
	GetLatestState() string
	GetSubscriptionID() uint64
	GetEventName() EventName
	GetEpoch() uint64
	GetInfo(format NotificationFormat) string
	GetTitle() string
	GetLegacyInfo() string
	GetLegacyTitle() string
	GetEventFilter() string
	SetEventFilter(filter string)
	GetEmailAttachment() *EmailAttachment
	GetUserId() UserId
	GetDashboardId() *int64
	GetDashboardName() string
	GetDashboardGroupId() *int64
	GetDashboardGroupName() string
	GetEntitiyId() string
}

type NotificationBaseImpl struct {
	LatestState        string
	SubscriptionID     uint64
	EventName          EventName
	Epoch              uint64
	Info               string
	Title              string
	EventFilter        string
	EmailAttachment    *EmailAttachment
	UserID             UserId
	DashboardId        *int64
	DashboardName      string
	DashboardGroupId   *int64
	DashboardGroupName string
}

func (n *NotificationBaseImpl) GetLatestState() string {
	return n.LatestState
}

func (n *NotificationBaseImpl) GetSubscriptionID() uint64 {
	return n.SubscriptionID
}

func (n *NotificationBaseImpl) GetEventName() EventName {
	return n.EventName
}

func (n *NotificationBaseImpl) GetEpoch() uint64 {
	return n.Epoch
}

func (n *NotificationBaseImpl) GetInfo(format NotificationFormat) string {
	return n.Info
}

func (n *NotificationBaseImpl) GetTitle() string {
	return n.Title
}

func (n *NotificationBaseImpl) GetLegacyInfo() string {
	return n.Info
}

func (n *NotificationBaseImpl) GetLegacyTitle() string {
	return n.Title
}

func (n *NotificationBaseImpl) GetEventFilter() string {
	return n.EventFilter
}

func (n *NotificationBaseImpl) SetEventFilter(filter string) {
	n.EventFilter = filter
}

func (n *NotificationBaseImpl) GetEmailAttachment() *EmailAttachment {
	return n.EmailAttachment
}

func (n *NotificationBaseImpl) GetUserId() UserId {
	return n.UserID
}

func (n *NotificationBaseImpl) GetDashboardId() *int64 {
	return n.DashboardId
}

func (n *NotificationBaseImpl) GetDashboardName() string {
	return n.DashboardName
}

func (n *NotificationBaseImpl) GetDashboardGroupId() *int64 {
	return n.DashboardGroupId
}

func (n *NotificationBaseImpl) GetDashboardGroupName() string {
	return n.DashboardGroupName
}

// func UnMarschal

type TaggedValidators struct {
	UserID             uint64 `db:"user_id"`
	Tag                string `db:"tag"`
	ValidatorPublickey []byte `db:"validator_publickey"`
	Validator          *Validator
	Events             []EventName `db:"events"`
}

type MinimalTaggedValidators struct {
	PubKey string
	Index  uint64
}

type OAuthAppData struct {
	ID          uint64 `db:"id"`
	Owner       uint64 `db:"owner_id"`
	AppName     string `db:"app_name"`
	RedirectURI string `db:"redirect_uri"`
	Active      bool   `db:"active"`
}

type OAuthCodeData struct {
	AppID  uint64 `db:"app_id"`
	UserID uint64 `db:"user_id"`
}

type MobileSettingsData struct {
	NotifyToken string `json:"notify_token"`
}

type MobileSubscription struct {
	ProductID   string                               `json:"id"`
	PriceMicros uint64                               `json:"priceMicros"`
	Currency    string                               `json:"currency"`
	Transaction MobileSubscriptionTransactionGeneric `json:"transaction"`
	Valid       bool                                 `json:"valid"`
}

type MobileSubscriptionTransactionGeneric struct {
	Type    string `json:"type"`
	Receipt string `json:"receipt"`
	ID      string `json:"id"`
}

type PremiumData struct {
	ID               uint64    `db:"id"`
	Receipt          string    `db:"receipt"`
	Store            string    `db:"store"`
	Active           bool      `db:"active"`
	ValidateRemotely bool      `db:"validate_remotely"`
	ProductID        string    `db:"product_id"`
	UserID           uint64    `db:"user_id"`
	ExpiresAt        time.Time `db:"expires_at"`
}

type UserWithPremium struct {
	ID      uint64         `db:"id"`
	Product sql.NullString `db:"product_id"`
}

type TransitEmail struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime        `db:"delivered"`
	Channel string              `db:"channel"`
	Content TransitEmailContent `db:"content"`
}

type TransitEmailContent struct {
	Address     string            `json:"address,omitempty"`
	Subject     string            `json:"subject,omitempty"`
	Email       Email             `json:"email,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	UserId      UserId            `json:"userId,omitempty"`
	CreatedTs   time.Time         `json:"-"`
}

func (e *TransitEmailContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitEmailContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitWebhook struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime          `db:"delivered"`
	Channel string                `db:"channel"`
	Content TransitWebhookContent `db:"content"`
}

type TransitWebhookContent struct {
	Webhook UserWebhook
	Event   WebhookEvent `json:"event"`
	UserId  UserId       `json:"userId"`
}

type WebhookEvent struct {
	Network     string `json:"network,omitempty"`
	Name        string `json:"event,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Epoch       uint64 `json:"epoch,omitempty"`
	Target      string `json:"target,omitempty"`
}

func (e *TransitWebhookContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitWebhookContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitDiscord struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime          `db:"delivered"`
	Channel string                `db:"channel"`
	Content TransitDiscordContent `db:"content"`
}

type TransitDiscordContent struct {
	Webhook        UserWebhook
	DiscordRequest DiscordReq `json:"discordRequest"`
	UserId         UserId     `json:"userId"`
}

func (e *TransitDiscordContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitDiscordContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type TransitPush struct {
	Id      uint64       `db:"id,omitempty"`
	Created sql.NullTime `db:"created"`
	Sent    sql.NullTime `db:"sent"`
	// Delivered sql.NullTime       `db:"delivered"`
	Channel string             `db:"channel"`
	Content TransitPushContent `db:"content"`
}

type TransitPushContent struct {
	Messages []*messaging.Message
	UserId   UserId `json:"userId"`
}

func (e *TransitPushContent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a TransitPushContent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type EmailAttachment struct {
	Attachment []byte `json:"attachment"`
	Name       string `json:"name"`
}

type Email struct {
	Title                 string
	Body                  template.HTML
	SubscriptionManageURL template.HTML
}

type UserWebhook struct {
	ID          uint64         `db:"id" json:"id"`
	UserID      uint64         `db:"user_id" json:"-"`
	Url         string         `db:"url" json:"url"`
	Retries     uint64         `db:"retries" json:"retries"`
	LastSent    sql.NullTime   `db:"last_sent" json:"lastRetry"`
	Response    sql.NullString `db:"response" json:"response"`
	Request     sql.NullString `db:"request" json:"request"`
	Destination sql.NullString `db:"destination" json:"destination"`
	EventNames  pq.StringArray `db:"event_names" json:"-"`
}

type UserWebhookSubscriptions struct {
	ID             uint64 `db:"id"`
	UserID         uint64 `db:"user_id"`
	WebhookID      uint64 `db:"webhook_id"`
	SubscriptionID uint64 `db:"subscription_id"`
}

type NotificationChannel string

var NotificationChannelLabels map[NotificationChannel]template.HTML = map[NotificationChannel]template.HTML{
	EmailNotificationChannel:          "Email Notification",
	PushNotificationChannel:           "Push Notification",
	WebhookNotificationChannel:        `Webhook Notification (<a href="/user/webhooks">configure</a>)`,
	WebhookDiscordNotificationChannel: "Discord Notification",
}

const (
	EmailNotificationChannel          NotificationChannel = "email"
	PushNotificationChannel           NotificationChannel = "push"
	WebhookNotificationChannel        NotificationChannel = "webhook"
	WebhookDiscordNotificationChannel NotificationChannel = "webhook_discord"
)

var NotificationChannels = []NotificationChannel{
	EmailNotificationChannel,
	PushNotificationChannel,
	WebhookNotificationChannel,
	WebhookDiscordNotificationChannel,
}

func GetNotificationChannel(channel string) (NotificationChannel, error) {
	for _, ch := range NotificationChannels {
		if string(ch) == channel {
			return ch, nil
		}
	}
	return "", errors.Errorf("Could not convert channel from string to NotificationChannel type. %v is not a known channel type", channel)
}

type ErrorResponse struct {
	Status string // e.g. "200 OK"
	Body   string
}

func (e *ErrorResponse) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a ErrorResponse) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type EnsSearchPageData = struct {
	Error  string
	Search string
	Result *EnsDomainResponse
}

type GasNowPageData struct {
	Code int `json:"code"`
	Data struct {
		Rapid     *big.Int `json:"rapid"`
		Fast      *big.Int `json:"fast"`
		Standard  *big.Int `json:"standard"`
		Slow      *big.Int `json:"slow"`
		Timestamp int64    `json:"timestamp"`
		Price     float64  `json:"price,omitempty"`
		PriceUSD  float64  `json:"priceUSD"`
		Currency  string   `json:"currency,omitempty"`
	} `json:"data"`
}

type Eth1AddressSearchItem struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Token   string `json:"token"`
}

type RawMempoolResponse struct {
	Pending map[string]map[string]*RawMempoolTransaction `json:"pending"`
	Queued  map[string]map[string]*RawMempoolTransaction `json:"queued"`
	BaseFee map[string]map[string]*RawMempoolTransaction `json:"baseFee"`

	TxsByHash map[common.Hash]*RawMempoolTransaction
}

func (mempool RawMempoolResponse) FindTxByHash(txHashString string) *RawMempoolTransaction {
	return mempool.TxsByHash[common.HexToHash(txHashString)]
}

type RawMempoolTransaction struct {
	Hash             common.Hash     `json:"hash"`
	From             *common.Address `json:"from"`
	To               *common.Address `json:"to"`
	Value            *hexutil.Big    `json:"value"`
	Gas              *hexutil.Big    `json:"gas"`
	GasFeeCap        *hexutil.Big    `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Nonce            *hexutil.Big    `json:"nonce"`
	Input            *string         `json:"input"`
	TransactionIndex *hexutil.Big    `json:"transactionIndex"`
}

type MempoolTxPageData struct {
	RawMempoolTransaction
	TargetIsContract   bool
	IsContractCreation bool
}

type SyncCommitteesStats struct {
	ParticipatedSlots uint64 `db:"participated_sync" json:"participatedSlots"`
	MissedSlots       uint64 `db:"missed_sync" json:"missedSlots"`
	OrphanedSlots     uint64 `db:"orphaned_sync" json:"-"`
	ScheduledSlots    uint64 `json:"scheduledSlots"`
}

type SignatureType string

const (
	MethodSignature SignatureType = "method"
	EventSignature  SignatureType = "event"
)

type SignatureImportStatus struct {
	LatestTimestamp *string `json:"latestTimestamp"`
	NextPage        *string `json:"nextPage"`
	HasFinished     bool    `json:"hasFinished"`
}

type Signature struct {
	Id        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text_signature"`
	Hex       string `json:"hex_signature"`
	Bytes     string `json:"bytes_signature"`
}

type SearchValidatorsByEth1Result []struct {
	Eth1Address      string        `db:"from_address_text" json:"eth1_address"`
	ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
	Count            uint64        `db:"count" json:"-"`
}
