package types

import (
	"github.com/gobitfly/beaconchain/pkg/api/enums"
)

// everything that goes in this file is for the data access layer only
// it won't be converted to typescript or used in the frontend

const DefaultGroupId = 0
const AllGroups = -1

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
	GetDirection() enums.SortOrder
}

type GenericCursor struct {
	Direction enums.SortOrder `json:"d"`
	Valid     bool            `json:"-"`
}

func (c GenericCursor) IsCursor() bool {
	return true
}

func (c GenericCursor) IsValid() bool {
	return c.Valid
}

func (c GenericCursor) GetDirection() enums.SortOrder {
	return c.Direction
}

type CLDepositsCursor struct {
	GenericCursor
	Slot      int64
	SlotIndex int64
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
