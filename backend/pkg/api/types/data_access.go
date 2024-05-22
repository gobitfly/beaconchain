package types

import (
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/shopspring/decimal"
)

// everything that goes in this file is for the data access layer only
// it won't be converted to typescript or used in the frontend

const DefaultGroupId = 0
const AllGroups = -1
const DefaultGroupName = "default"

type Sort[T enums.Enum] struct {
	Column T
	Desc   bool
}

type VDBIdPrimary int
type VDBIdPublic string
type VDBIdValidatorSet []VDBValidator
type VDBId struct {
	Validators VDBIdValidatorSet // if this is nil, then use the id
	Id         VDBIdPrimary
}

type VDBValidator struct {
	Index uint64 `db:"validatorindex"`
}

type DashboardInfo struct {
	Id     VDBIdPrimary `db:"id"` // this must be the bigint id
	UserId uint64       `db:"user_id"`
}

type CursorLike interface {
	IsCursor() bool
	IsValid() bool
	IsReverse() bool
}

type GenericCursor struct {
	Reverse bool `json:"r"`
	Valid   bool `json:"-"`
}

func (c GenericCursor) IsCursor() bool {
	return true
}

func (c GenericCursor) IsValid() bool {
	return c.Valid
}

// note: dont have to check for valid when calling this
func (c GenericCursor) IsReverse() bool {
	return c.Reverse && c.Valid
}

type CLDepositsCursor struct {
	GenericCursor
	Slot      int64
	SlotIndex int64
}

type ELDepositsCursor struct {
	GenericCursor
	BlockNumber int64
	LogIndex    int64
}

type ValidatorsCursor struct {
	GenericCursor

	Index uint64 `json:"vi"`
}

type RewardsCursor struct {
	GenericCursor

	Epoch   uint64
	GroupId int64
}

type ValidatorDutiesCursor struct {
	GenericCursor

	Index  uint64
	Reward decimal.Decimal
}

type WithdrawalsCursor struct {
	GenericCursor

	Slot            uint64
	WithdrawalIndex uint64
	Index           uint64
	Recipient       []byte
	Amount          uint64
}

type User struct {
	Id        uint64 `db:"id"`
	Password  string `db:"password"`
	ProductId string `db:"product_id"`
	UserGroup string `db:"user_group"`
}

type BlocksCursor struct {
	GenericCursor
	Slot int64 // basically the same as Block, Epoch, Age; mandatory, used to index

	// optional, max one of those (for now)
	Proposer int64
	Group    int64
	Status   int64
	Reward   int64
}

type NetworkInfo struct {
	ChainId uint64
	Name    string
}
