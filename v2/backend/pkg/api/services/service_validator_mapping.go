package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/pkg/errors"
)

type ValidatorMapping struct {
	ValidatorPubkeys  []string
	ValidatorIndices  map[string]*uint64
	ValidatorMetadata []*ValidatorMetadata
}

type ValidatorMetadata struct {
	ActivationEpoch   types.Epoch
	WithdrawableEpoch types.Epoch
}

var currentValidatorMapping *ValidatorMapping
var lastValidatorIndex int

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

func initValidatorMapping(data *types.RedisCachedValidatorsMapping) error {
	log.Infof("initializing validator mapping")
	l := len(data.Mapping)

	c := ValidatorMapping{}
	c.ValidatorIndices = make(map[string]*uint64)
	c.ValidatorPubkeys = make([]string, l)
	c.ValidatorMetadata = make([]*ValidatorMetadata, l)

	for i, v := range data.Mapping {
		if i == l {
			break
		}

		b := hex.EncodeToString(v.PublicKey)
		j := uint64(i)

		c.ValidatorPubkeys[i] = b
		c.ValidatorIndices[b] = &j
		c.ValidatorMetadata[i] = &ValidatorMetadata{
			ActivationEpoch:   v.ActivationEpoch,
			WithdrawableEpoch: v.WithdrawableEpoch,
		}
	}
	currentValidatorMapping = &c
	lastValidatorIndex = l - 1

	return nil
}

func quickUpdateValidatorMapping(data *types.RedisCachedValidatorsMapping) error {
	log.Infof("quick updating validator mapping")

	for i, v := range data.Mapping {
		if i > lastValidatorIndex {
			b := hex.EncodeToString(v.PublicKey)
			j := uint64(i)

			currentValidatorMapping.ValidatorPubkeys = append(currentValidatorMapping.ValidatorPubkeys, b)
			currentValidatorMapping.ValidatorIndices[b] = &j
			currentValidatorMapping.ValidatorMetadata = append(currentValidatorMapping.ValidatorMetadata, &ValidatorMetadata{
				ActivationEpoch:   v.ActivationEpoch,
				WithdrawableEpoch: v.WithdrawableEpoch,
			})

			lastValidatorIndex = i
			continue
		}
		currentValidatorMapping.ValidatorMetadata[i].ActivationEpoch = v.ActivationEpoch
		currentValidatorMapping.ValidatorMetadata[i].WithdrawableEpoch = v.WithdrawableEpoch
	}
	return nil
}

func updateValidatorMapping() error {
	var validatorMapping *types.RedisCachedValidatorsMapping

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	key := fmt.Sprintf("%d:%s", utils.Config.Chain.ClConfig.DepositChainID, "vm")
	encoded, err := db.PersistentRedisDbClient.Get(ctx, key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to get compressed validator mapping from db")
	}
	log.Infof("reading validator mapping from redis done, took %s", time.Since(start))

	// ungob
	start = time.Now()
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&validatorMapping)
	if err != nil {
		return errors.Wrap(err, "error decoding assignments data")
	}
	log.Debugf("decoding validator mapping from gob took %s", time.Since(start))

	currentMappingMutex.Lock()
	start = time.Now()
	if currentValidatorMapping == nil {
		initValidatorMapping(validatorMapping)
	} else {
		quickUpdateValidatorMapping(validatorMapping)
	}
	log.Debugf("updated Validator Mapping, took %s", time.Since(start))
	currentMappingMutex.Unlock()

	return nil
}

// GetCurrentValidatorMapping returns the current validator mapping and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the validator mapping service from updating
func GetCurrentValidatorMapping() (*ValidatorMapping, func(), error) {
	currentMappingMutex.RLock()

	if currentValidatorMapping == nil {
		return nil, currentMappingMutex.RUnlock, errors.New("waiting for validator mapping to be initialized")
	}

	return currentValidatorMapping, currentMappingMutex.RUnlock, nil
}

func GetPubkeysOfValidatorIndexSlice(indices []uint64) ([]string, error) {
	res := make([]string, len(indices))
	mapping, releaseLock, err := GetCurrentValidatorMapping()
	defer releaseLock()
	if err != nil {
		return nil, err
	}
	for i, index := range indices {
		if index > uint64(lastValidatorIndex) {
			return nil, fmt.Errorf("validator index outside of mapped range (%d is not within 0-%d)", index, lastValidatorIndex)
		}
		res[i] = mapping.ValidatorPubkeys[index]
	}
	return res, nil
}

func GetValidatorIndexOfPubkeySlice(pubkeys []string) ([]uint64, error) {
	res := make([]uint64, len(pubkeys))
	mapping, releaseLock, err := GetCurrentValidatorMapping()
	defer releaseLock()
	if err != nil {
		return nil, err
	}
	for i, pubkey := range pubkeys {
		p, ok := mapping.ValidatorIndices[pubkey]
		if !ok {
			return nil, fmt.Errorf("pubkey not found in validator mapping: %s", pubkey)
		}
		res[i] = *p
	}
	return res, nil
}
