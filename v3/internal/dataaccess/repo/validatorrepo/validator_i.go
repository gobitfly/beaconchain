package validatorrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type Repository interface {
	GetBalances(ctx context.Context, chain domain.Chain, timestamp int, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorBalance, error)
}
