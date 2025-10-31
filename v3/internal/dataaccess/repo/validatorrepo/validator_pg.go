package validatorrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/lib/pq"
)

type DBRepository struct {
	repo.ChainReader
}

func (r *DBRepository) GetBalances(ctx context.Context, chain domain.Chain, time time.Time, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorBalance, error) {
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
				goqu.I("b.validator_index").Eq(goqu.I("v.validator_index")),
			),
		).
		Where(
			goqu.And(
				goqu.I("b.epoch_timestamp").Eq(time),
				goqu.I("b.validator_index").In(
					goqu.Select(goqu.I("validator_index")).From(goqu.T("v")),
				),
			),
		).
		Order(goqu.I("b.validator_index").Asc()).
		Limit(uint(pageSize))
	if cursor != nil {
		query = query.Where(goqu.I("b.validator_index").Gt(cursor.Index))
	}
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

func (r *DBRepository) GetHeadBalances(ctx context.Context, chain domain.Chain, time time.Time, indices []domain.ValidatorIndex) ([]domain.ValidatorBalance, error) {
	query := goqu.Dialect("postgres").
		From(goqu.T("legacy_validator_epoch_balances").As("b")).
		Select(
			goqu.I("b.validator_index").As("validatorindex"),
			goqu.I("b.balance").As("currentbalance"),
			goqu.I("b.effective_balance").As("effectivebalance"),
		).
		Where(
			goqu.And(
				goqu.I("b.epoch_timestamp").Eq(time),
				goqu.I("b.validator_index").In(indices),
			),
		)

	return repo.RunQueryRows[[]domain.ValidatorBalance](ctx, r.SelectClickhouse(chain), query)
}

func (r *DBRepository) GetOverview(ctx context.Context, chain domain.Chain, selector domain.ValidatorsSelector, cursor *domain.ValidatorIndexCursor, pageSize int) ([]domain.ValidatorOverview, error) {
	query := goqu.Dialect("postgres").
		From(goqu.T("validators").As("v")).
		Select(
			goqu.I("v.validatorindex").As("validatorindex"),
			goqu.I("v.pubkey").As("validatorpublickey"),
			goqu.I("v.slashed").As("slashed"),
			goqu.Case().
				When(goqu.I("v.status").Like("%_online"), goqu.L("TRUE")).
				When(goqu.I("v.status").Like("%_offline"), goqu.L("FALSE")).
				Else(goqu.L("NULL")).
				As("online"),
			goqu.I("v.withdrawalcredentials").As("withdrawalcredential"),
			// return nil instead of max int64 for these epoch fields
			goqu.L("NULLIF(v.activationeligibilityepoch, ~(1::BIGINT << 63))").As("activationeligibilityepoch"),
			goqu.L("NULLIF(v.activationepoch, ~(1::BIGINT << 63))").As("activationepoch"),
			goqu.L("NULLIF(v.exitepoch, ~(1::BIGINT << 63))").As("exitepoch"),
			goqu.L("NULLIF(v.withdrawableepoch, ~(1::BIGINT << 63))").As("withdrawableepoch"),
		).
		Order(goqu.I("v.validatorindex").Asc()).
		Limit(uint(pageSize))
	if cursor != nil {
		query = query.Where(
			goqu.I("v.validatorindex").Gt(cursor.Index),
		)
	}

	// selector logic here could be extracted if more methods need it
	switch {
	case selector.DepositAddress != nil:
		depositQuery := goqu.Dialect("postgres").
			From(goqu.T("eth1_deposits")).
			SelectDistinct("publickey").
			Where(goqu.I("from_address").Eq(*selector.DepositAddress))
		query = query.Join(depositQuery.As("d"), goqu.On(
			goqu.I("v.pubkey").Eq(goqu.I("d.publickey")),
		))
	case selector.WithdrawalAddress != nil:
		query = query.Where(goqu.And(
			goqu.L("SUBSTRING(v.withdrawalcredentials FROM 1 FOR 1)").Neq([]byte{0x00}), // ensure is not a 0x00 prefix
			goqu.L("SUBSTRING(v.withdrawalcredentials FROM 13 FOR 20)").Eq(*selector.WithdrawalAddress),
		))
	case selector.WithdrawalCredential != nil:
		query = query.Where(goqu.I("v.withdrawalcredentials").Eq(*selector.WithdrawalCredential))
	case selector.Identifiers != nil:
		query = query.Where(
			goqu.Or(
				goqu.L("v.validatorindex = ANY(?)", pq.Array(selector.Identifiers.Indices)),
				goqu.L("v.pubkey = ANY(?)", pq.Array(selector.Identifiers.PublicKeys)),
			),
		)
	default:
		return nil, fmt.Errorf("invalid selector")
	}

	return repo.RunQueryRows[[]domain.ValidatorOverview](ctx, r.SelectPostgres(chain), query)
}
