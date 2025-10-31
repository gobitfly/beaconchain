package validatorrepo

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type Repository interface {
	GetOverview(ctx context.Context, chain domain.Chain, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorOverview, error)
	GetBalances(ctx context.Context, chain domain.Chain, time time.Time, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorBalance, error)
	GetHeadBalances(ctx context.Context, chain domain.Chain, time time.Time, validatorIndices []domain.ValidatorIndex) ([]domain.ValidatorBalance, error)
}
