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
	UpdateRelays(tagID, endpoint string) error
	UpdateRelayLastExportTry(tagID, endpoint string) error
	UpdateRelayExportFailureCount(exportFailureCount uint64, tagID, endpoint string) error
	GetFirstRelayBlock(tagID string) (types.RelayBlock, error)
	GetLastRelayBlock(tagID string) (types.RelayBlock, error)
	SaveBlocksTags(tagID string, slot uint64, blockHash []byte) error
	SaveBlocksRelays(tagID string, slot uint64, payloadValue types.WeiString, blockHash, builderPubkey, proposerPubkey, proposerFeeRecipient []byte) error
}
