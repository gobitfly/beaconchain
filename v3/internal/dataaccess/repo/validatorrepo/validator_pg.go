package validatorrepo

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/lib/pq"
)

type DBRepository struct {
	repo.ChainReader
}

func (r *DBRepository) GetBalances(ctx context.Context, chain domain.Chain, timestamp int, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorBalance, error) {
	query := goqu.Dialect("postgres").
		From(goqu.T("_final_validator_balance_epoch").As("b")).
		With("v", selectedValidators(selector)).
		Select(
			goqu.I("b.validator_index").As("ValidatorIndex"),
			goqu.I("v.public_key").As("PublicKey"),
			goqu.I("b.current_balance").As("CurrentBalance"),
			goqu.I("b.effective_balance").As("EffectiveBalance"),
		).
		Join(
			goqu.T("v"),
			goqu.On(
				goqu.I("b.validator_index").Eq(goqu.I("v.validatorindex")),
			),
		).
		Where(
			goqu.And(
				goqu.L("b.epoch_timestamp = fromUnixTimestamp(?)", timestamp),
				goqu.I("b.validator_index").In(
					goqu.Select(goqu.I("validatorindex")).From(goqu.T("v")),
				),
			),
		).
		Order(goqu.I("b.validator_index").Asc()).
		Limit(uint(pageSize))
	return repo.RunQueryRows[[]domain.ValidatorBalance](ctx, r.SelectClickhouse(chain), query)
}

// selectedValidators builds a CTE that selects validator indices based on the given selector.
// Intended to be used by all validator-related queries
func selectedValidators(selector domain.ValidatorsSelector) *goqu.SelectDataset {
	baseQuery := goqu.Dialect("postgres").
		From(goqu.T("_lookup_validator FINAL")).
		SelectDistinct(
			goqu.I("validator_index"),
			goqu.I("public_key"),
		)
	if selector.DepositAddress != nil {
		return baseQuery.
			Where(goqu.And(
				goqu.I("selector_type").Eq("deposit_address"),
				goqu.I("selector").Eq(*selector.DepositAddress),
			))
	}
	if selector.WithdrawalAddress != nil {
		return baseQuery.
			Where(goqu.And(
				goqu.I("selector_type").Eq("withdrawal_address"),
				goqu.I("selector").Eq(*selector.WithdrawalAddress),
			))
	}
	if selector.WithdrawalCredential != nil {
		return baseQuery.
			Where(goqu.And(
				goqu.I("selector_type").Eq("withdrawal_credential"),
				goqu.I("selector").Eq(*selector.WithdrawalCredential),
			))
	}
	if selector.Identifiers != nil {
		return baseQuery.
			Where(
				goqu.Or(
					goqu.L("validator_index = ANY(?)", pq.Array(selector.Identifiers.Indices)),
					goqu.L("public_key = ANY(?)", pq.Array(selector.Identifiers.PublicKeys)),
				),
			)
	}
	return nil
}
