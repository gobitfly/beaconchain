package ethereumnetworkrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type Repository interface {
	LatestStateRepository
	GetSlot(ctx context.Context, chain domain.Chain, slot int) (*domain.Slot, error)
}

type LatestStateRepository interface {
	GetLatestState(ctx context.Context, chain domain.Chain, view domain.ConsensusView) (domain.LatestState, error)
}
