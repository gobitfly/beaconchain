package dataaccess

import (
	"context"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

type SearchRepository interface {
	GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error)
	GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error)
	GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error)
	GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error)
	GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error)
	GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error)
	GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error)
}

func (d *DataAccessService) GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorByIndex(ctx, chainId, index)
}

func (d *DataAccessService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorByPublicKey(ctx, chainId, publicKey)
}

func (d *DataAccessService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByDepositAddress(ctx, chainId, address)
}

func (d *DataAccessService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByDepositEnsName(ctx, chainId, ensName)
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByWithdrawalCredential(ctx, chainId, credential)
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByWithdrawalEnsName(ctx, chainId, ensName)
}

func (d *DataAccessService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	// TODO: @recy21
	return d.dummy.GetSearchValidatorsByGraffiti(ctx, chainId, graffiti)
}
