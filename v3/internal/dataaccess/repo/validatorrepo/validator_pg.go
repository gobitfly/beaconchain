package validatorrepo

import (
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
)

type DBRepository struct {
	repo.ChainReader
}
