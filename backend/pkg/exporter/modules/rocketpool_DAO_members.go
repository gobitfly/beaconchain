package modules

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	rpDAOTrustedNode "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type RocketpoolDAOMember struct {
	Address                []byte    `db:"address"`
	ID                     string    `db:"id"`
	URL                    string    `url:"url"`
	JoinedTime             time.Time `db:"joined_time"`
	LastProposalTime       time.Time `db:"last_proposal_time"`
	RPLBondAmount          *big.Int  `db:"rpl_bond_amount"`
	UnbondedValidatorCount uint64    `db:"unbonded_validator_count"`
}

func (rp *RocketpoolExporter) SaveDAOMembers() error {
	if len(rp.DAOMembersByAddress) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-dao-members")
	}(timeStart)

	data := rp.prepareDAOMemberData()
	return rp.saveDAOMembers(data)
}

func (rp *RocketpoolExporter) prepareDAOMemberData() []*RocketpoolDAOMember {
	data := make([]*RocketpoolDAOMember, len(rp.DAOMembersByAddress))
	i := 0
	for _, val := range rp.DAOMembersByAddress {
		data[i] = val
		i++
	}
	return data
}

func (rp *RocketpoolExporter) saveDAOMembers(data []*RocketpoolDAOMember) error {
	nArgs := 8
	valueStringsTpl := generateSQLParamPlaceholders(nArgs)
	batchSize := 1000

	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs, addresses := rp.prepareDAOMemberBatch(data[start:end], valueStringsTpl, nArgs)
		if err := rp.Database.SaveRocketPoolDAOMembers(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_members: %w", err)
		}

		if err := rp.Database.DeleteRocketPoolDAOMembers(addresses); err != nil {
			return fmt.Errorf("error deleting from rocketpool_dao_members: %w", err)
		}
	}

	return nil
}

func (rp *RocketpoolExporter) prepareDAOMemberBatch(data []*RocketpoolDAOMember, valueStringsTpl string, nArgs int) ([]string, []interface{}, [][]byte) {
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*nArgs)
	addresses := make([][]byte, 0, len(data))

	for i, d := range data {
		for j := 0; j < nArgs; j++ {
			valueStringsArgs[j] = i*nArgs + j + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
		valueArgs = append(valueArgs, d.Address)
		valueArgs = append(valueArgs, d.ID)
		valueArgs = append(valueArgs, d.URL)
		valueArgs = append(valueArgs, d.JoinedTime)
		valueArgs = append(valueArgs, d.LastProposalTime)
		valueArgs = append(valueArgs, d.RPLBondAmount.String())
		valueArgs = append(valueArgs, d.UnbondedValidatorCount)
		addresses = append(addresses, d.Address)
	}

	return valueStrings, valueArgs, addresses
}

func (rp *RocketpoolExporter) UpdateDAOMembers() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-dao-members")
	}(timeStart)

	members, err := rpDAOTrustedNode.GetMembers(rp.API, nil)
	if err != nil {
		return err
	}

	return rp.updateDAOMembers(members)
}

func (rp *RocketpoolExporter) updateDAOMembers(members []rpDAOTrustedNode.MemberDetails) error {
	for _, m := range members {
		addrHex := m.Address.Hex()
		if member, exists := rp.DAOMembersByAddress[addrHex]; exists {
			if err := member.Update(rp.API); err != nil {
				return err
			}
			continue
		}

		member, err := NewRocketpoolDAOMember(rp.API, m.Address.Bytes())
		if err != nil {
			return err
		}
		rp.DAOMembersByAddress[addrHex] = member
	}
	return nil
}

func NewRocketpoolDAOMember(rp *rocketpool.RocketPool, addr []byte) (*RocketpoolDAOMember, error) {
	m := &RocketpoolDAOMember{}
	m.Address = addr
	err := m.Update(rp)
	if err != nil {
		return m, err
	}
	return m, nil
}

func (r *RocketpoolDAOMember) Update(rp *rocketpool.RocketPool) error {
	d, err := rpDAOTrustedNode.GetMemberDetails(rp, common.BytesToAddress(r.Address), nil)
	if err != nil {
		return err
	}
	r.ID = d.ID
	r.URL = d.Url
	r.JoinedTime = time.Unix(int64(d.JoinedTime), 0)
	r.LastProposalTime = time.Unix(int64(d.LastProposalTime), 0)
	r.RPLBondAmount = d.RPLBondAmount
	r.UnbondedValidatorCount = d.UnbondedValidatorCount
	return nil
}
