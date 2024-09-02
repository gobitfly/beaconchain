package dataaccess

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
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
	// TODO: implement handling of chainid
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	if int(index) < len(validatorMapping.ValidatorPubkeys) {
		return &t.SearchValidator{
			Index:     index,
			PublicKey: hexutil.MustDecode(validatorMapping.ValidatorPubkeys[index]),
		}, nil
	}

	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error) {
	// TODO: implement handling of chainid
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}

	b := hexutil.Encode(publicKey)
	if index, found := validatorMapping.ValidatorIndices[b]; found {
		return &t.SearchValidator{
			Index:     index,
			PublicKey: publicKey,
		}, nil
	}

	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByDepositAddress{
		Address: address,
	}
	err := db.ReaderDb.GetContext(ctx, &ret.Count, "select count(validatorindex) from validators where pubkey in (select publickey from eth1_deposits where from_address = $1);", address)
	if err != nil {
		return nil, err
	}
	if ret.Count == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}

func (d *DataAccessService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositEnsName, error) {
	// TODO: implement handling of chainid
	// TODO: finalize ens implementation first
	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrwalCredential, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByWithdrwalCredential{
		WithdrawalCredential: credential,
	}
	err := db.ReaderDb.GetContext(ctx, &ret.Count, "select count(validatorindex) from validators where withdrawalcredentials = $1;", credential)
	if err != nil {
		return nil, err
	}
	if ret.Count == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithrawalEnsName, error) {
	// TODO: implement handling of chainid
	// TODO: finalize ens implementation first
	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByGraffiti{
		Graffiti: graffiti,
	}
	err := db.ReaderDb.GetContext(ctx, &ret.Count, "select count(distinct proposer) from blocks where graffiti_text = $1;", graffiti)
	if err != nil {
		return nil, err
	}
	if ret.Count == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}
