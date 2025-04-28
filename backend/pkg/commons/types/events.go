package types

const ConsensusLayerEventVersion = 6

type ConsensusLayerEvent struct {
	EventName  ConsensusLayerEventName `db:"event_name"`
	EventId    string                  `db:"id"`
	BlockRoot  []byte                  `db:"block_root"`
	Slot       int64                   `db:"slot"`
	EventIndex int64                   `db:"event_index"`
	RawData    []byte                  `db:"data"`
}

type ConsensusLayerEventName string

const (
	EpochProcessedEventName         ConsensusLayerEventName = "EpochProcessedEvent"
	DepositQueuedEventName          ConsensusLayerEventName = "DepositQueuedEvent"
	DepositProcessedEventName       ConsensusLayerEventName = "DepositProcessedEvent"
	DepositRejectedEventName        ConsensusLayerEventName = "DepositRejectedEvent"
	DepositPostponeEventName        ConsensusLayerEventName = "DepositPostponedEvent"
	WithdrawalQueuedEventName       ConsensusLayerEventName = "WithdrawalQueuedEvent"
	WithdrawalProcessedEventName    ConsensusLayerEventName = "WithdrawalProcessedEvent"
	WithdrawalRejectedEventName     ConsensusLayerEventName = "WithdrawalRejectedEvent"
	ConsolidationQueuedEventName    ConsensusLayerEventName = "ConsolidationQueuedEvent"
	ConsolidationProcessedEventName ConsensusLayerEventName = "ConsolidationProcessedEvent"
	ConsolidationRejectedEventName  ConsensusLayerEventName = "ConsolidationRejectedEvent"
	SwitchToCompoundingEventName    ConsensusLayerEventName = "SwitchToCompoundingEvent"
	StartupEventName                ConsensusLayerEventName = "StartupEvent"
	RemovedExcessBalanceEventName   ConsensusLayerEventName = "RemovedExcessBalance"
	ExitRequestProcessedEventName   ConsensusLayerEventName = "ExitRequestProcessedEvent"
)

type ConsensusLayerEventFilter struct {
	Slot      uint64
	BlockRoot []byte
}

type EventBase struct {
	Slot       uint64 `json:"slot"`
	BlockRoot  []byte `json:"block_root"`
	EventIndex int    `json:"event_index"`
	Version    int    `json:"version"`
}

type EpochProcessedEvent struct {
	EventBase
}

type SwitchToCompoundingEvent struct {
	EventBase
	Address []byte `json:"address"`
	Pubkey  []byte `json:"pubkey"`
}

type ConsolidationQueuedEvent struct {
	EventBase
	SourceAddress []byte `json:"source_address"`
	SourcePubkey  []byte `json:"source_pubkey"`
	TargetPubkey  []byte `json:"target_pubkey"`
}

type ConsolidationRejectedEvent struct {
	EventBase
	SourceAddress []byte `json:"source_address"`
	SourcePubkey  []byte `json:"source_pubkey"`
	TargetPubkey  []byte `json:"target_pubkey"`
	Reason        string `json:"reason"`
	PreQueue      bool   `json:"pre_queue"`
}
type ConsolidationProcessedEvent struct {
	EventBase
	SourcePubkey []byte `json:"source_pubkey"`
	TargetPubkey []byte `json:"target_pubkey"`
	Amount       uint64 `json:"amount"`
}

type DepositQueuedEvent struct {
	EventBase
	Pubkey                []byte `json:"pubkey"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials"`
	Signature             []byte `json:"signature"`
}

type DepositProcessedEvent struct {
	EventBase
	Pubkey                []byte `json:"pubkey"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials"`
	Signature             []byte `json:"signature"`
	SignatureValid        bool   `json:"signature_valid"`
}

type DepositRejectedEvent struct {
	EventBase
	Pubkey                []byte `json:"pubkey"`
	Reason                string `json:"reason"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials"`
	Signature             []byte `json:"signature"`
	PreQueue              bool   `json:"pre_queue"`
}

type DepositPostponedEvent struct {
	EventBase
	Pubkey                []byte `json:"pubkey"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials"`
	Signature             []byte `json:"signature"`
}
type StartupEvent struct {
	EventBase
}

type RemovedExcessBalance struct {
	EventBase
	Pubkey []byte `json:"pubkey"`
	Amount uint64 `json:"amount"`
}
type WithdrawalQueuedEvent struct {
	EventBase
	Pubkey            []byte `json:"pubkey"`
	Amount            uint64 `json:"amount"`
	WithdrawableEpoch uint64 `json:"withdrawable_epoch"`
}

type WithdrawalProcessedEvent struct {
	EventBase
	Pubkey         []byte `json:"pubkey"`
	Amount         uint64 `json:"amount"`
	OriginalAmount uint64 `json:"original_amount"`
}
type WithdrawalRejectedEvent struct {
	EventBase
	Pubkey   []byte `json:"pubkey"`
	Reason   string `json:"reason"`
	Amount   uint64 `json:"amount"`
	PreQueue bool   `json:"pre_queue"`
}

type ExitRequestProcessedEvent struct {
	EventBase
	Pubkey []byte `json:"pubkey"`
}

type DepositRequestType string

const (
	DepositRequestAccountType      DepositRequestType = "account"
	DepositRequestSystemExcessType DepositRequestType = "system_excess"
	DepositRequestGenesisType      DepositRequestType = "genesis"
)

type GenericEventStatus string

const (
	GenericEventStatusQueued    GenericEventStatus = "queued"
	GenericEventStatusProcessed GenericEventStatus = "completed"
	GenericEventStatusRejected  GenericEventStatus = "rejected"
	GenericEventStatusPostponed GenericEventStatus = "postponed"
)

type DepositRequestDBRow struct {
	Id               int64  `db:"id"`
	ExecutionLayerID *int64 `db:"eth1_id"`

	SlotQueued      *int64 `db:"slot_queued"`
	IndexQueued     *int64 `db:"index_queued"`
	BlockQueuedRoot []byte `db:"block_queued_root"`

	SlotProcessed      *int64 `db:"slot_processed"`
	IndexProcessed     *int64 `db:"index_processed"`
	BlockProcessedRoot []byte `db:"block_processed_root"`

	Type         DepositRequestType `db:"type"`
	Status       GenericEventStatus `db:"status"`
	RejectReason *string            `db:"reject_reason"`

	Pubkey                []byte `db:"pubkey"`
	WithdrawalCredentials []byte `db:"withdrawal_credentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

type ConsolidationRequestDBRow struct {
	Id               int64  `db:"id"`
	ExecutionLayerID *int64 `db:"eth1_id"`

	SlotQueued      *int64 `db:"slot_queued"`
	IndexQueued     *int64 `db:"index_queued"`
	BlockQueuedRoot []byte `db:"block_queued_root"`

	SlotProcessed      *int64 `db:"slot_processed"`
	IndexProcessed     *int64 `db:"index_processed"`
	BlockProcessedRoot []byte `db:"block_processed_root"`

	Status             GenericEventStatus `db:"status"`
	RejectReason       *string            `db:"reject_reason"`
	AmountConsolidated *uint64            `db:"amount_consolidated"`
	SourcePubkey       []byte             `db:"source_pubkey"`
	TargetPubkey       []byte             `db:"target_pubkey"`
}

type WithdrawalRequestDBRow struct {
	Id               int64  `db:"id"`
	ExecutionLayerID *int64 `db:"eth1_id"`

	SlotQueued      *int64 `db:"slot_queued"`
	IndexQueued     *int64 `db:"index_queued"`
	BlockQueuedRoot []byte `db:"block_queued_root"`

	SlotProcessed      *int64 `db:"slot_processed"`
	IndexProcessed     *int64 `db:"index_processed"`
	BlockProcessedRoot []byte `db:"block_processed_root"`

	Status          GenericEventStatus `db:"status"`
	RejectReason    *string            `db:"reject_reason"`
	ValidatorPubkey []byte             `db:"validator_pubkey"`
	Amount          uint64             `db:"amount"`
}

type SwitchToCompoundingRequestDBRow struct {
	Id               int64  `db:"id"`
	ExecutionLayerID *int64 `db:"eth1_id"`

	SlotProcessed      *int64 `db:"slot_processed"`
	IndexProcessed     *int64 `db:"index_processed"`
	BlockProcessedRoot []byte `db:"block_processed_root"`

	Status          GenericEventStatus `db:"status"`
	RejectReason    *string            `db:"reject_reason"`
	ValidatorPubkey []byte             `db:"validator_pubkey"`
}

type EpochBlockRoots struct {
	Epoch uint64
	// between n*slotsPerEpoch and (n+1)*slotsPerEpoch-1. missed slots are not included
	//
	// note, there is a special case for the last slot of the epoch, it will always be set to a block root (or the function will return an error)
	//
	// this is because events will be emitted in the last slot of the epoch during process_epoch, with the block root set to whatever the last proposed block was.
	// in the ideal case this is simply the block of the slot itself, but it can be a different slot as well, and when the epoch itself has no proposed slots,
	// it can be from a different epoch altogether as well.
	SlotBlockRoots map[uint64][]byte
	// the special case the description of SlotBlockRoots describes. this exist to make it easier to access. used to verify EpochProcessed events
	EpochBlockRoot []byte
}
