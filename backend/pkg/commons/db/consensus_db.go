package db

import (
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/jmoiron/sqlx"
)

type ConsensusDB struct {
	WriterDb *sqlx.DB
	ReaderDb *sqlx.DB
}

type ConsensusDBI interface {
	GetRelays() ([]types.Relay, error)
	UpdateRelay(tagID, endpoint string) error
	UpdateRelayLastExportTry(tagID, endpoint string) error
	UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error
	GetFirstRelayBlock(tagID string) (types.RelayBlock, error)
	GetLastRelayBlock(tagID string) (types.RelayBlock, error)
	SaveBlockTagsAndRelays(tagID string, payload types.BidTrace) error
}
