package dataaccess

import (
	"context"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type SearchRepository interface {
	GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error)
	GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey string) (*t.SearchValidator, error)
	GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address string) (*t.SearchValidatorsByDepositAddress, error)
	GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error)
	GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential string) (*t.SearchValidatorsByWithdrwalCredential, error)
	GetSearchValidatorsByWithdrawalAddress(ctx context.Context, chainId uint64, address string) (*t.SearchValidatorsByDepositAddress, error)
	GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error)
	GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error)
}

func (d *DataAccessService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorByIndex(ctx, chainId, index)
}

func (d *DataAccessService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey string) (*t.SearchValidator, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorByPublicKey(ctx, chainId, publicKey)
}

func (d *DataAccessService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address string) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByDepositAddress(ctx, chainId, address)
}

func (d *DataAccessService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByDepositEnsName(ctx, chainId, ensName)
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential string) (*t.SearchValidatorsByWithdrwalCredential, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, credential)
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalAddress(ctx context.Context, chainId uint64, address string) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByWithdrawalAddress(ctx, chainId, address)
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByWithdrawalEnsName(ctx, chainId, ensName)
}

func (d *DataAccessService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByGraffiti(ctx, chainId, graffiti)
}
