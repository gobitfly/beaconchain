package domain

type Chain string

const (
	ChainUnknown Chain = ""
	ChainMainnet Chain = "mainnet"
	ChainHoodi   Chain = "hoodi"
	ChainSepolia Chain = "sepolia"
)

func (c Chain) IsValid() bool {
	return c != ChainUnknown
}

func (c Chain) GetChainID() int {
	switch c {
	case ChainMainnet:
		return 1
	case ChainHoodi:
		return 560048
	case ChainSepolia:
		return 11155111
	default:
		return 0
	}
}

type Slot struct {
	Slot                     int
	AttestationSlashingCount int32
	ProposerSlashingCount    int32
	AttestationCount         int32
	BlockRoot                []byte

	ProcessedAutoWithdrawals   ClEventDetails
	ProcessedConsolidations    ClEventDetails
	ProcessedDeposits          ClEventDetails
	ProcessedManualWithdrawals ClEventDetails

	QueuedConsolidations ClEventDetails
	QueuedDeposits       ClEventDetails
	QueuedWithdrawals    ClEventDetails

	Graffiti  []byte
	Proposer  Validator
	Status    DutyStatus
	Finalized bool
}

type DutyStatus string

type Validator struct {
	Index  int32
	Pubkey []byte
}

type ClEventDetails struct {
	Count  int64
	Amount int64
}

type LatestState struct {
	Slot          int
	Epoch         int
	ConsensusView ConsensusView
}

type ConsensusView string

const (
	ConsensusViewHead      ConsensusView = "head"
	ConsensusViewJustified ConsensusView = "justified"
	ConsensusViewFinalized ConsensusView = "finalized"
)

type BlockTransaction struct {
	Hash string
	Idx  int // field used for pagination
}

type BlockTransactionCursor struct {
	Idx int
}
