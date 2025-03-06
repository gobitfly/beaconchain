package consensus

import (
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type ConsensusRepository interface {
	GetRelays() ([]types.Relay, error)
	UpdateRelay(tagID, endpoint string) error
	UpdateRelayLastExportTry(tagID, endpoint string) error
	UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error
	GetFirstRelayBlock(tagID string) (types.RelayBlock, error)
	GetLastRelayBlock(tagID string) (types.RelayBlock, error)
	SaveBlockTagsAndRelays(tagID string, payload types.BidTrace) error
}

type consensusRepository struct {
	ReaderDb *sqlx.DB
	WriterDb *sqlx.DB
}

func NewConsensusRepository(reader, writer *sqlx.DB) ConsensusRepository {
	return &consensusRepository{
		ReaderDb: reader,
		WriterDb: writer,
	}
}
