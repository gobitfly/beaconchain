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
	Index   uint64
	Version uint64
}

type DashboardInfo struct {
	Id     VDBIdPrimary // this must be the bigint id
	UserId uint64
}

type CursorLike interface {
	isCursor() bool
}

type GenericCursor struct {
	Direction enums.SortOrder `json:"d"`
}

func (b GenericCursor) isCursor() bool {
	return true
}
