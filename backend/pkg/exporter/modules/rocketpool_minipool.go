package modules

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rpTypes "github.com/rocket-pool/rocketpool-go/types"
	"golang.org/x/sync/errgroup"
)

type MinipoolPerformanceFile struct {
	Index               uint64                                               `json:"index"`
	Network             string                                               `json:"network"`
	MinipoolPerformance map[common.Address]*SmoothingPoolMinipoolPerformance `json:"minipoolPerformance"`
}

// Minipool stats
type SmoothingPoolMinipoolPerformance struct {
	Pubkey                  string   `json:"pubkey"`
	SuccessfulAttestations  uint64   `json:"successfulAttestations"`
	MissedAttestations      uint64   `json:"missedAttestations"`
	ParticipationRate       float64  `json:"participationRate"`
	MissingAttestationSlots []uint64 `json:"missingAttestationSlots"`
	EthEarned               float64  `json:"ethEarned"`
}

func (rp *RocketpoolExporter) SaveMinipools() error {
	if len(rp.MinipoolsByAddress) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-minipools")
	}(timeStart)

	data := rp.prepareMinipoolData()
	if err := rp.saveMinipoolData(data); err != nil {
		return err
	}

	// updating index column after writing minipools for cheaper access later
	return db.UpdateRocketPoolMiniPools()
}

func (rp *RocketpoolExporter) prepareMinipoolData() []*RocketpoolMinipool {
	data := make([]*RocketpoolMinipool, len(rp.MinipoolsByAddress))
	i := 0
	for _, pool := range rp.MinipoolsByAddress {
		data[i] = pool
		i++
	}

	return data
}

func (rp *RocketpoolExporter) saveMinipoolData(data []*RocketpoolMinipool) error {
	nArgs := 14
	valueStringsTpl := createValueStringsTemplate(nArgs)
	batchSize := 1000
	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs := rp.prepareMinipoolBatch(data[start:end], valueStringsTpl, nArgs)
		if err := db.SaveRocketPoolMiniPools(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into rocketpool_minipools: %w", err)
		}
	}

	return nil
}

func (rp *RocketpoolExporter) prepareMinipoolBatch(data []*RocketpoolMinipool, valueStringsTpl string, nArgs int) ([]string, []interface{}) {
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*nArgs)

	for i, d := range data {
		for j := 0; j < nArgs; j++ {
			valueStringsArgs[j] = i*nArgs + j + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
		valueArgs = append(valueArgs, d.Address)
		valueArgs = append(valueArgs, d.Pubkey)
		valueArgs = append(valueArgs, d.Status)
		valueArgs = append(valueArgs, d.StatusTime)
		valueArgs = append(valueArgs, d.NodeAddress)
		valueArgs = append(valueArgs, d.NodeFee)
		valueArgs = append(valueArgs, d.DepositType)
		valueArgs = append(valueArgs, d.PenaltyCount)
		valueArgs = append(valueArgs, d.NodeDepositBalance.String())
		valueArgs = append(valueArgs, d.NodeRefundBalance.String())
		valueArgs = append(valueArgs, d.UserDepositBalance.String())
		valueArgs = append(valueArgs, d.IsVacant)
		valueArgs = append(valueArgs, d.Version)
	}

	return valueStrings, valueArgs
}

func (rp *RocketpoolExporter) UpdateMinipools() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.InfoWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-minipools")
	}(timeStart)

	minipoolAddresses, err := minipool.GetMinipoolAddresses(rp.API, nil)
	if err != nil {
		return err
	}

	atlasDeployed, err := IsAtlasDeployed(rp.API)
	if err != nil {
		return err
	}

	return rp.updateMinipools(minipoolAddresses, atlasDeployed)
}

func (rp *RocketpoolExporter) updateMinipools(minipoolAddresses []common.Address, atlasDeployed bool) error {
	for _, a := range minipoolAddresses {
		addrHex := a.Hex()
		if mp, exists := rp.MinipoolsByAddress[addrHex]; exists {
			if err := mp.Update(rp.API, atlasDeployed); err != nil {
				return err
			}
			continue
		}
		mp, err := NewRocketpoolMinipool(rp.API, a.Bytes(), atlasDeployed)
		if err != nil {
			return err
		}
		rp.MinipoolsByAddress[addrHex] = mp
	}
	return nil
}

type RocketpoolMinipool struct {
	Address            []byte    `db:"address"`
	Pubkey             []byte    `db:"pubkey"`
	NodeAddress        []byte    `db:"node_address"`
	NodeFee            float64   `db:"node_fee"`
	DepositType        string    `db:"deposit_type"`
	Status             string    `db:"status"`
	StatusTime         time.Time `db:"status_time"`
	PenaltyCount       uint64    `db:"penalty_count"`
	NodeDepositBalance *big.Int  `db:"node_deposit_balance"`
	NodeRefundBalance  *big.Int  `db:"node_refund_balance"`
	UserDepositBalance *big.Int  `db:"user_deposit_balance"`
	IsVacant           bool      `db:"is_vacant"`
	Version            uint8     `db:"version"`
}

func NewRocketpoolMinipool(rp *rocketpool.RocketPool, addr []byte, atlasDeployed bool) (*RocketpoolMinipool, error) {
	pubk, err := minipool.GetMinipoolPubkey(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(addr), nil)
	if err != nil {
		return nil, err
	}
	nodeAddr, err := mp.GetNodeAddress(nil)
	if err != nil {
		return nil, err
	}

	rpm := &RocketpoolMinipool{
		Address:     addr,
		Pubkey:      pubk.Bytes(),
		NodeAddress: nodeAddr.Bytes(),
	}
	err = rpm.Update(rp, atlasDeployed)
	if err != nil {
		return nil, err
	}
	return rpm, nil
}

func (r *RocketpoolMinipool) Update(rp *rocketpool.RocketPool, atlasDeployed bool) error {
	mp, err := minipool.NewMinipool(rp, common.BytesToAddress(r.Address), nil)
	if err != nil {
		return err
	}

	var wg errgroup.Group
	var status rpTypes.MinipoolStatus
	var statusTime time.Time
	var penaltyCount uint64
	var nodeFee float64

	var nodeDepositBalance, nodeRefundBalance, userDepositBalance *big.Int = leb16, big.NewInt(0), leb16
	var version uint8
	var statusDetail minipool.StatusDetails = minipool.StatusDetails{
		IsVacant: false,
	}
	var depositType rpTypes.MinipoolDeposit

	// Node fee can change on conversion starting with Atlas
	wg.Go(func() error {
		var err error
		nodeFee, err = mp.GetNodeFee(nil)
		return err
	})

	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		statusTime, err = mp.GetStatusTime(nil)
		return err
	})
	wg.Go(func() error {
		var err error
		penaltyCount, err = minipool.GetMinipoolPenaltyCount(rp, common.BytesToAddress(r.Address), nil)
		return err
	})

	if atlasDeployed {
		wg.Go(func() error {
			var err error
			nodeDepositBalance, err = mp.GetNodeDepositBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			userDepositBalance, err = mp.GetUserDepositBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			nodeRefundBalance, err = mp.GetNodeRefundBalance(nil)
			return err
		})

		wg.Go(func() error {
			var err error
			statusDetail, err = mp.GetStatusDetails(nil)
			return err
		})
	}

	wg.Go(func() error {
		var err error
		version = mp.GetVersion()
		return err
	})

	wg.Go(func() error {
		var err error
		depositType, err = mp.GetDepositType(nil)
		return err
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	r.NodeFee = nodeFee
	r.Status = status.String()
	r.StatusTime = statusTime
	r.PenaltyCount = penaltyCount
	r.Version = version
	r.NodeDepositBalance = nodeDepositBalance
	r.NodeRefundBalance = nodeRefundBalance
	r.UserDepositBalance = userDepositBalance
	r.IsVacant = statusDetail.IsVacant
	r.DepositType = depositType.String()
	return nil
}
