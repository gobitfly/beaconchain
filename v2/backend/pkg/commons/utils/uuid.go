package utils

import (
	"sync/atomic"

	"github.com/bwmarrin/snowflake"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/google/uuid"
)

// uuid that you can get  - gets set to a random value on startup/first read
var _uuid atomic.Value
var _snowflakeGenerator atomic.Value

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

func GetSnowflake() int64 {
	if v := _snowflakeGenerator.Load(); v != nil {
		return v.(*snowflake.Node).Generate().Int64()
	}

	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatal(err, "snowflake generator failed to start", 0)
		return 0
	}
	_snowflakeGenerator.CompareAndSwap(nil, node)
	return _snowflakeGenerator.Load().(*snowflake.Node).Generate().Int64()
}
