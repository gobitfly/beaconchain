package dataaccess

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type ValidatorRepository interface {
	GetValidatorsByDepositAddress(ctx context.Context, depositAddress string) ([]t.VDBValidator, error)
	GetValidatorsByWithdrawalCredentials(ctx context.Context, withdrawalCredentials string) ([]t.VDBValidator, error)
	GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error)
}

func (d *DataAccessService) GetValidatorsByDepositAddress(ctx context.Context, depositAddress string) ([]t.VDBValidator, error) {
	addressParsed, err := hex.DecodeString(strings.TrimPrefix(depositAddress, "0x"))
	if err != nil {
		return nil, err
	}

	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("validators").As("v")).
		SelectDistinct(goqu.I("v.validatorindex")).
		InnerJoin(
			goqu.T("eth1_deposits").As("d"),
			goqu.On(goqu.I("v.pubkey").Eq(goqu.I("d.publickey"))),
		).
		Where(
			goqu.I("d.from_address").Eq(addressParsed),
		)

	query, args, err := validatorsDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	var validators []uint64
	err = d.alloyReader.SelectContext(ctx, &validators, query, args...)
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func (d *DataAccessService) GetValidatorsByWithdrawalCredentials(ctx context.Context, withdrawalCredentials string) ([]t.VDBValidator, error) {
	addressParsed, err := hex.DecodeString(strings.TrimPrefix(withdrawalCredentials, "0x"))
	if err != nil {
		return nil, err
	}

	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("validators").As("v")).
		SelectDistinct(goqu.I("v.validatorindex")).
		Where(
			goqu.I("v.withdrawalcredentials").Eq(addressParsed),
		)

	query, args, err := validatorsDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	var validators []uint64
	err = d.alloyReader.SelectContext(ctx, &validators, query, args...)
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func (d *DataAccessService) GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error) {
	// includes orphaned blocks
	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks").As("b")).
		SelectDistinct(goqu.I("b.proposer")).
		Where(
			goqu.I("b.graffiti_text").Eq(graffiti),
		)

	query, args, err := validatorsDs.Prepared(true).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	var validators []uint64
	err = d.alloyReader.SelectContext(ctx, &validators, query, args...)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
