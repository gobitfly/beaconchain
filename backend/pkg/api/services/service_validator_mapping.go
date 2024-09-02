package services

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/gobitfly/beaconchain/pkg/monitoring/services"
	"github.com/klauspost/pgzip"
	"github.com/pkg/errors"
)

type ValidatorMapping struct {
	ValidatorPubkeys  []string
	ValidatorIndices  map[string]constypes.ValidatorIndex // note: why pointers?
	ValidatorMetadata []*types.CachedValidator            // note: why pointers?
}

var currentValidatorMapping unsafe.Pointer
var _cachedBufferCompressed = new(bytes.Buffer)
var _cachedBufferDecompressed = new(bytes.Buffer)
var _cachedRedisValidatorMapping = new(types.RedisCachedValidatorsMapping)

var lastEpochUpdate = uint64(0)

func (s *Services) startIndexMappingService() {
	var err error
	for {
		startTime := time.Now()
		delay := time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot) * time.Second
		err = nil // clear error
		r := services.NewStatusReport("api_service_validator_mapping", constants.Default, delay)
		r(constants.Running, nil)
		latestEpoch := cache.LatestEpoch.Get()
		if currentValidatorMapping == nil || latestEpoch != lastEpochUpdate {
			err = s.updateValidatorMapping()
		}
		if err != nil {
			log.Error(err, "error updating validator mapping", 0)
			r(constants.Failure, map[string]string{"error": err.Error()})
			delay = 10 * time.Second
		} else {
			log.Infof("=== validator mapping updated in %s", time.Since(startTime))
			r(constants.Success, map[string]string{"took": time.Since(startTime).String(), "latest_epoch": fmt.Sprintf("%d", lastEpochUpdate)})
			lastEpochUpdate = latestEpoch
		}
		utils.ConstantTimeDelay(startTime, delay)
	}
}

func (s *Services) initValidatorMapping() {
	log.Infof("initializing validator mapping")
	lenMapping := len(_cachedRedisValidatorMapping.Mapping)

	c := ValidatorMapping{}
	c.ValidatorIndices = make(map[string]constypes.ValidatorIndex, lenMapping)
	c.ValidatorPubkeys = make([]string, lenMapping)
	c.ValidatorMetadata = _cachedRedisValidatorMapping.Mapping

	for i, v := range _cachedRedisValidatorMapping.Mapping {
		if i == lenMapping {
			break
		}

		b := hexutil.Encode(v.PublicKey)
		j := constypes.ValidatorIndex(i)

		c.ValidatorPubkeys[i] = b
		c.ValidatorIndices[b] = j
	}
	atomic.StorePointer(&currentValidatorMapping, unsafe.Pointer(&c))
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
	_cachedBufferCompressed.Write(compressed)
	defer _cachedBufferCompressed.Reset()
	w, err := pgzip.NewReaderN(_cachedBufferCompressed, 1_000_000, 10)
	if err != nil {
		return errors.Wrap(err, "failed to create pgzip reader")
	}
	defer w.Close()
	_, err = w.WriteTo(_cachedBufferDecompressed)
	defer _cachedBufferDecompressed.Reset()
	if err != nil {
		return errors.Wrap(err, "failed to decompress validator mapping from redis")
	}
	log.Debugf("decompressing validator mapping using pgzip took %s", time.Since(start))

	// ungob
	start = time.Now()
	dec := gob.NewDecoder(_cachedBufferDecompressed)
	err = dec.Decode(&_cachedRedisValidatorMapping)
	if err != nil {
		return errors.Wrap(err, "error decoding assignments data")
	}
	log.Debugf("decoding validator mapping from gob took %s", time.Since(start))

	start = time.Now()
	s.initValidatorMapping() // no more quick update as we are pointer swapping
	log.Debugf("updated Validator Mapping, took %s", time.Since(start))

	return nil
}

// GetCurrentValidatorMapping returns the current validator mapping and a function to release the lock
// Call release lock after you are done with accessing the data, otherwise it will block the validator mapping service from updating
func (s *Services) GetCurrentValidatorMapping() (*ValidatorMapping, error) {
	// in theory the consumer can just check if the pointer is nil, but this is more explicit
	if currentValidatorMapping == nil {
		return nil, fmt.Errorf("%w: validator mapping", ErrWaiting)
	}
	return (*ValidatorMapping)(atomic.LoadPointer(&currentValidatorMapping)), nil
}

func (s *Services) GetPubkeySliceFromIndexSlice(indices []constypes.ValidatorIndex) ([]string, error) {
	res := make([]string, len(indices))
	mapping, err := s.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}
	lastValidatorIndex := constypes.ValidatorIndex(len(mapping.ValidatorPubkeys) - 1)
	for i, index := range indices {
		if index > lastValidatorIndex {
			return nil, fmt.Errorf("validator index outside of mapped range (%d is not within 0-%d)", index, lastValidatorIndex)
		}
		res[i] = mapping.ValidatorPubkeys[index]
	}
	return res, nil
}

func (s *Services) GetIndexSliceFromPubkeySlice(pubkeys []string) ([]constypes.ValidatorIndex, error) {
	res := make([]constypes.ValidatorIndex, len(pubkeys))
	mapping, err := s.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}
	for i, pubkey := range pubkeys {
		p, ok := mapping.ValidatorIndices[pubkey]
		if !ok {
			return nil, fmt.Errorf("pubkey %s not found in mapping", pubkey)
		}
		res[i] = p
	}
	return res, nil
}
