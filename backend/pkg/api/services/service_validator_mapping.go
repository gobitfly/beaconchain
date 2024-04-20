package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/klauspost/pgzip"
	"github.com/pkg/errors"
)

type ValidatorMapping struct {
	ValidatorPubkeys  []string
	ValidatorIndices  map[string]*uint64
	ValidatorMetadata []*types.CachedValidator
}

var currentValidatorMapping *ValidatorMapping
var lastValidatorIndex int
var _cachedBufferCompressed = new(bytes.Buffer)
var _cachedBufferDecompressed = new(bytes.Buffer)
var _cachedValidatorMapping = new(types.RedisCachedValidatorsMapping)

var currentMappingMutex = &sync.RWMutex{}

func (s *Services) startIndexMappingService() {
	for {
		startTime := time.Now()
		err := s.updateValidatorMapping() // TODO: only update data if something has changed (new head epoch)
		delay := time.Duration(utils.Config.Chain.ClConfig.SlotsPerEpoch*utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second
		if err != nil {
			log.Error(err, "error updating validator mapping", 0)
			delay = 10 * time.Second
		} else {
			log.Infof("=== validator mapping updated in %s", time.Since(startTime))
		}
		utils.ConstantTimeDelay(startTime, delay)
	}
}

func (s *Services) initValidatorMapping() {
	log.Infof("initializing validator mapping")
	lenMapping := len(_cachedValidatorMapping.Mapping)

	c := ValidatorMapping{}
	c.ValidatorIndices = make(map[string]*uint64, lenMapping)
	c.ValidatorPubkeys = make([]string, lenMapping)
	c.ValidatorMetadata = _cachedValidatorMapping.Mapping

	for i, v := range _cachedValidatorMapping.Mapping {
		if i == lenMapping {
			break
		}

		b := hexutil.Encode(v.PublicKey)
		j := uint64(i)

		c.ValidatorPubkeys[i] = b
		c.ValidatorIndices[b] = &j
	}
	currentValidatorMapping = &c
	lastValidatorIndex = lenMapping - 1
}

func (s *Services) quickUpdateValidatorMapping() {
	log.Infof("quick updating validator mapping")
	// update metadata by overwriting it
	currentValidatorMapping.ValidatorMetadata = _cachedValidatorMapping.Mapping

	newLastValidatorIndex := len(_cachedValidatorMapping.Mapping) - 1

	if newLastValidatorIndex <= lastValidatorIndex {
		log.Debugf("no new validators to add to mapping")
		return
	}
	// update mappings
	for i := lastValidatorIndex + 1; i <= newLastValidatorIndex; i++ {
		v := _cachedValidatorMapping.Mapping[i]
		b := hexutil.Encode(v.PublicKey)
		j := uint64(i)

		currentValidatorMapping.ValidatorPubkeys = append(currentValidatorMapping.ValidatorPubkeys, b)
		currentValidatorMapping.ValidatorIndices[b] = &j
	}
	lastValidatorIndex = newLastValidatorIndex
}

func (s *Services) updateValidatorMapping() error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	key := fmt.Sprintf("%d:%s", utils.Config.Chain.ClConfig.DepositChainID, "vm")
	compressed, err := s.persistentRedisDbClient.Get(ctx, key).Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to get compressed validator mapping from db")
	}
	log.Debugf("reading validator mapping from redis done, took %s", time.Since(start))

	// decompress
	start = time.Now()
	_cachedBufferCompressed.Reset()
	_cachedBufferCompressed.Write(compressed)
	w, err := pgzip.NewReaderN(_cachedBufferCompressed, 1_000_000, 10)
	if err != nil {
		return errors.Wrap(err, "failed to create pgzip reader")
	}
	defer w.Close()
	_cachedBufferDecompressed.Reset()
	_, err = w.WriteTo(_cachedBufferDecompressed)
	if err != nil {
		return errors.Wrap(err, "failed to decompress validator mapping from redis")
	}
	log.Debugf("decompressing validator mapping using pgzip took %s", time.Since(start))

	// ungob
	start = time.Now()
	dec := gob.NewDecoder(_cachedBufferDecompressed)
	err = dec.Decode(&_cachedValidatorMapping)
	if err != nil {
		return errors.Wrap(err, "error decoding assignments data")
	}
	log.Debugf("decoding validator mapping from gob took %s", time.Since(start))

	currentMappingMutex.Lock()
	start = time.Now()
	if currentValidatorMapping == nil {
		s.initValidatorMapping()
	} else {
		s.quickUpdateValidatorMapping()
	}
	log.Debugf("updated Validator Mapping, took %s", time.Since(start))
	currentMappingMutex.Unlock()

	// free up memory
	_cachedBufferCompressed.Reset()
	_cachedBufferDecompressed.Reset()

	return nil
}

// GetCurrentValidatorMapping returns the current validator mapping and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the validator mapping service from updating
func (s *Services) GetCurrentValidatorMapping() (*ValidatorMapping, func(), error) {
	currentMappingMutex.RLock()

	if currentValidatorMapping == nil {
		return nil, currentMappingMutex.RUnlock, errors.New("waiting for validator mapping to be initialized")
	}

	return currentValidatorMapping, currentMappingMutex.RUnlock, nil
}

func (s *Services) GetPubkeysOfValidatorIndexSlice(indices []uint64) ([]string, error) {
	res := make([]string, len(indices))
	mapping, releaseLock, err := s.GetCurrentValidatorMapping()
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

func (s *Services) GetValidatorIndexOfPubkeySlice(pubkeys []string) ([]uint64, error) {
	res := make([]uint64, len(pubkeys))
	mapping, releaseLock, err := s.GetCurrentValidatorMapping()
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
