package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
)

var currentValidatorMapping *types.RedisCachedValidatorsMapping

var currentMappingMutex = &sync.RWMutex{}

func StartIndexMappingService() {
	for {
		startTime := time.Now()
		err := updateValidatorMapping() // TODO: only update data if something has changed (new head epoch)
		if err != nil {
			log.Error(err, "error updating validator mapping", 0)
		}
		log.Infof("=== validator mapping updated in %s", time.Since(startTime))
		utils.ConstantTimeDelay(startTime, 32*12*time.Second)
	}
}

func updateValidatorMapping() error {
	var validatorMapping *types.RedisCachedValidatorsMapping

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	key := fmt.Sprintf("%d:%s", utils.Config.Chain.ClConfig.DepositChainID, "vm")
	encoded, err := db.PersistentRedisDbClient.Get(ctx, key).Result()
	if err != nil {
		return errors.Wrap(err, "failed to get compressed validator mapping from db")
	}
	log.Infof("reading validator mapping from redis done, took %s", time.Since(start))

	// ungob
	start = time.Now()
	var buf bytes.Buffer
	buf.Write([]byte(encoded))
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&validatorMapping)
	if err != nil {
		return errors.Wrap(err, "error decoding assignments data")
	}
	log.Debugf("decoding validator mapping from gob took %s", time.Since(start))

	currentMappingMutex.Lock()
	currentValidatorMapping = validatorMapping
	currentMappingMutex.Unlock()

	return nil
}
