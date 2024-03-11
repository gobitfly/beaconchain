package types

import "github.com/gobitfly/beaconchain/pkg/api/enums"

// everything that goes in this file is for the data access layer only
// it won't be converted to typescript or used in the frontend

type Sort[T enums.Enum] struct {
	Column T
	Desc   bool
}

type VDBValidator struct {
	Index   uint64
	Version uint64
}

type VDBIdPrimary int
type VDBIdPublic string
type VDBIdValidatorSet []VDBValidator

type DashboardInfo struct {
	Id     VDBIdPrimary `db:"id"` // this must be the bigint id
	UserId uint64       `db:"user_id"`
}
