package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	rpDAO "github.com/rocket-pool/rocketpool-go/dao"
	rpDAOTrustedNode "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type RocketpoolDAOProposalMemberVotes struct {
	ProposalID uint64 `db:"id"`
	Address    []byte `db:"member_address"`
	Voted      bool   `db:"voted"`
	Supported  bool   `db:"supported"`
}

type RocketpoolDAOProposal struct {
	ID              uint64    `db:"id"`
	DAO             string    `db:"dao"`
	ProposerAddress []byte    `db:"proposer_address"`
	Message         string    `db:"message"`
	CreatedTime     time.Time `db:"created_time"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
	ExpiryTime      time.Time `db:"expiry_time"`
	VotesRequired   float64   `db:"votes_required"`
	VotesFor        float64   `db:"votes_for"`
	VotesAgainst    float64   `db:"votes_against"`
	MemberVoted     bool      `db:"member_voted"`
	MemberSupported bool      `db:"member_supported"`
	IsCancelled     bool      `db:"is_cancelled"`
	IsExecuted      bool      `db:"is_executed"`
	Payload         []byte    `db:"payload"`
	State           string    `db:"state"`
	MemberVotes     []RocketpoolDAOProposalMemberVotes
}

func NewRocketpoolDAOProposal(rp *rocketpool.RocketPool, pid uint64) (*RocketpoolDAOProposal, error) {
	p := &RocketpoolDAOProposal{ID: pid}
	err := p.Update(rp)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *RocketpoolDAOProposal) Update(rp *rocketpool.RocketPool) error { // update
	pd, err := rpDAO.GetProposalDetails(rp, r.ID, nil)
	if err != nil {
		return err
	}
	r.ID = pd.ID
	r.DAO = pd.DAO
	r.ProposerAddress = pd.ProposerAddress.Bytes()
	r.Message = pd.Message
	r.CreatedTime = time.Unix(int64(pd.CreatedTime), 0)
	r.StartTime = time.Unix(int64(pd.StartTime), 0)
	r.EndTime = time.Unix(int64(pd.EndTime), 0)
	r.ExpiryTime = time.Unix(int64(pd.ExpiryTime), 0)
	r.VotesRequired = pd.VotesRequired
	r.VotesFor = pd.VotesFor
	r.VotesAgainst = pd.VotesAgainst
	r.MemberVoted = pd.MemberVoted
	r.MemberSupported = pd.MemberSupported
	r.IsCancelled = pd.IsCancelled
	r.IsExecuted = pd.IsExecuted
	r.Payload = pd.Payload
	r.State = pd.State.String()

	// Update member votes
	r.MemberVotes = []RocketpoolDAOProposalMemberVotes{}
	members, err := rpDAOTrustedNode.GetMembers(rp, nil)
	if err != nil {
		return err
	}
	for _, m := range members {
		memberVoted, err := rpDAO.GetProposalMemberVoted(rp, r.ID, m.Address, nil)
		if err != nil {
			return err
		}

		memberSupported, err := rpDAO.GetProposalMemberSupported(rp, r.ID, m.Address, nil)
		if err != nil {
			return err
		}

		r.MemberVotes = append(r.MemberVotes, RocketpoolDAOProposalMemberVotes{
			ProposalID: r.ID,
			Address:    m.Address.Bytes(),
			Voted:      memberVoted,
			Supported:  memberSupported,
		})
	}

	return nil
}

func (rp *RocketpoolExporter) SaveDAOProposals() error {
	if len(rp.DAOProposalsByID) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-dao-proposals")
	}(timeStart)

	data := rp.prepareDAOProposalData()
	return rp.saveDAOProposals(data)
}

func (rp *RocketpoolExporter) prepareDAOProposalData() []*RocketpoolDAOProposal {
	data := make([]*RocketpoolDAOProposal, len(rp.DAOProposalsByID))
	i := 0
	for _, val := range rp.DAOProposalsByID {
		data[i] = val
		i++
	}
	return data
}

func (rp *RocketpoolExporter) saveDAOProposals(data []*RocketpoolDAOProposal) error {
	nArgs := 18
	valueStringsTpl := createValueStringsTemplate(nArgs)
	batchSize := 1000

	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs := rp.prepareDAOProposalBatch(data[start:end], valueStringsTpl, nArgs)
		if err := db.SaveRocketPoolDAOProposals(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_proposals: %w", err)
		}
	}
	return nil
}

func (rp *RocketpoolExporter) prepareDAOProposalBatch(data []*RocketpoolDAOProposal, valueStringsTpl string, nArgs int) ([]string, []interface{}) {
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*nArgs)

	for i, d := range data {
		for j := 0; j < nArgs; j++ {
			valueStringsArgs[j] = i*nArgs + j + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
		valueArgs = append(valueArgs, d.ID)
		valueArgs = append(valueArgs, d.DAO)
		valueArgs = append(valueArgs, d.ProposerAddress)
		valueArgs = append(valueArgs, d.Message)
		valueArgs = append(valueArgs, d.CreatedTime)
		valueArgs = append(valueArgs, d.StartTime)
		valueArgs = append(valueArgs, d.EndTime)
		valueArgs = append(valueArgs, d.ExpiryTime)
		valueArgs = append(valueArgs, d.VotesRequired)
		valueArgs = append(valueArgs, d.VotesFor)
		valueArgs = append(valueArgs, d.VotesAgainst)
		valueArgs = append(valueArgs, d.MemberVoted)
		valueArgs = append(valueArgs, d.MemberSupported)
		valueArgs = append(valueArgs, d.IsCancelled)
		valueArgs = append(valueArgs, d.IsExecuted)
		valueArgs = append(valueArgs, d.Payload)
		valueArgs = append(valueArgs, d.State)
	}

	return valueStrings, valueArgs
}

func (rp *RocketpoolExporter) SaveDAOProposalsMemberVotes() error {
	if len(rp.DAOProposalsByID) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-dao-proposals-member-votes")
	}(timeStart)

	data := rp.prepareDAOProposalMemberVotesData()
	return rp.saveDAOProposalMemberVotes(data)
}

func (rp *RocketpoolExporter) prepareDAOProposalMemberVotesData() []RocketpoolDAOProposalMemberVotes {
	data := []RocketpoolDAOProposalMemberVotes{}
	for _, val := range rp.DAOProposalsByID {
		data = append(data, val.MemberVotes...)
	}
	return data
}

func (rp *RocketpoolExporter) saveDAOProposalMemberVotes(data []RocketpoolDAOProposalMemberVotes) error {
	nArgs := 5
	valueStringsTpl := createValueStringsTemplate(nArgs)
	batchSize := 1000

	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs := rp.prepareDAOProposalMemberVotesBatch(data[start:end], valueStringsTpl, nArgs)
		if err := db.SaveRocketPoolDAOProposalVotes(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into rocketpool_dao_proposals_member_votes: %w", err)
		}
	}
	return nil
}

func (rp *RocketpoolExporter) prepareDAOProposalMemberVotesBatch(data []RocketpoolDAOProposalMemberVotes, valueStringsTpl string, nArgs int) ([]string, []interface{}) {
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*nArgs)

	for i, d := range data {
		for j := 0; j < nArgs; j++ {
			valueStringsArgs[j] = i*nArgs + j + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
		valueArgs = append(valueArgs, d.ProposalID)
		valueArgs = append(valueArgs, d.Address)
		valueArgs = append(valueArgs, d.Voted)
		valueArgs = append(valueArgs, d.Supported)
	}

	return valueStrings, valueArgs
}

func (rp *RocketpoolExporter) UpdateDAOProposals() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-dao-proposals")
	}(timeStart)

	pc, err := rpDAO.GetProposalCount(rp.API, nil)
	if err != nil {
		return err
	}

	return rp.updateDAOProposals(pc)
}

func (rp *RocketpoolExporter) updateDAOProposals(proposalCount uint64) error {
	for i := uint64(0); i < proposalCount; i++ {
		p, err := NewRocketpoolDAOProposal(rp.API, i+1)
		if err != nil {
			return err
		}
		rp.DAOProposalsByID[i] = p
	}

	return nil
}
