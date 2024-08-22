package utils

import (
	"sync/atomic"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/google/uuid"
)

// uuid that you can get  - gets set to a random value on startup/first read
var _uuid atomic.Value

// GetUUID returns the uuid
func GetUUID() string {
	if v := _uuid.Load(); v != nil {
		return v.(string)
	}

	tmp_uuid := uuid.NewString()
	if _uuid.CompareAndSwap(nil, tmp_uuid) {
		log.Infof("uuid set to %s", tmp_uuid)
	}
	return _uuid.Load().(string)
}
