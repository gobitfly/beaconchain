package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
)

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
	Id     VDBIdPrimary // this must be the bigint id
	UserId uint64
}

type PostgresOffsetColumn struct {
	ColumnName string `json:"n"`
	Value      int64  `json:"v"`
}

type PostgresCursor struct {
	Direction enums.SortOrder        `json:"d"`
	Offsets   []PostgresOffsetColumn `json:"o"`
}

func (p PostgresCursor) ToString() (*string, error) {
	bin, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PostgresCursor as json: %w", err)
	}
	encoded_str := base64.RawURLEncoding.EncodeToString(bin)
	return &encoded_str, nil
}

func (PostgresCursor) FromString(str string) (*PostgresCursor, error) {
	bin, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode string using base64: %w", err)
	}

	p := PostgresCursor{}
	err = json.Unmarshal(bin, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decoded base64 string: %w", err)
	}

	return &p, nil
}
