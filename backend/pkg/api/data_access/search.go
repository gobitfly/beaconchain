package dataaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/common/hexutil"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/pkg/errors"
)

type SearchRepository interface {
	GetSearchValidatorByIndex(ctx context.Context, chainId, index uint64) (*t.SearchValidator, error)
	GetSearchValidatorByPublicKey(ctx context.Context, chainId uint64, publicKey []byte) (*t.SearchValidator, error)
	GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error)
	GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositAddress, error)
	GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrawalCredential, error)
	GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithdrawalCredential, error)
	GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error)
	GetSearchValidatorsByGraffitiHex(ctx context.Context, chainId uint64, graffiti []byte) (*t.SearchValidatorsByGraffiti, error)
	GetSearchAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchAddress, error)
	GetSearchTransaction(ctx context.Context, chainId uint64, transactionHash []byte) (*t.SearchTransaction, error)
	GetSearchBlock(ctx context.Context, chainId uint64, blockNumber uint64) (*t.SearchBlock, error)
	GetSearchEpoch(ctx context.Context, chainId uint64, epoch uint64) (*t.SearchEpoch, error)
	GetSearchToken(ctx context.Context, chainId uint64, address []byte) (*t.SearchToken, error)
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
			PublicKey: validatorMapping.ValidatorPubkeys[index],
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
			PublicKey: b,
		}, nil
	}

	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByDepositAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByDepositAddress{
		DepositAddress: hexutil.Encode(address),
	}
	err := db.ReaderDb.GetContext(ctx, &ret.Count, `
		select count(validatorindex) from validators where pubkey in (select publickey from eth1_deposits where from_address = $1);`, address)
	if err != nil {
		return nil, err
	}
	if ret.Count == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}

func (d *DataAccessService) GetSearchValidatorsByDepositEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByDepositAddress, error) {
	// TODO: implement handling of chainid
	// TODO: finalize ens implementation first
	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByWithdrawalCredential(ctx context.Context, chainId uint64, credential []byte) (*t.SearchValidatorsByWithdrawalCredential, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByWithdrawalCredential{
		WithdrawalCredential: hexutil.Encode(credential),
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

func (d *DataAccessService) GetSearchValidatorsByWithdrawalEnsName(ctx context.Context, chainId uint64, ensName string) (*t.SearchValidatorsByWithdrawalCredential, error) {
	// TODO: implement handling of chainid
	// TODO: finalize ens implementation first
	return nil, ErrNotFound
}

func (d *DataAccessService) GetSearchValidatorsByGraffiti(ctx context.Context, chainId uint64, graffiti string) (*t.SearchValidatorsByGraffiti, error) {
	// TODO: implement handling of chainid
	graffitiHex := [32]byte{}
	copy(graffitiHex[:], graffiti)
	ret := &t.SearchValidatorsByGraffiti{
		Graffiti: graffiti,
		Hex:      hexutil.Encode(graffitiHex[:]),
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

func (d *DataAccessService) GetSearchValidatorsByGraffitiHex(ctx context.Context, chainId uint64, graffiti []byte) (*t.SearchValidatorsByGraffiti, error) {
	// TODO: implement handling of chainid
	ret := &t.SearchValidatorsByGraffiti{
		Graffiti: strings.TrimRight(string(graffiti), "\u0000"),
		Hex:      hexutil.Encode(graffiti),
	}
	err := db.ReaderDb.GetContext(ctx, &ret.Count, "select count(distinct proposer) from blocks where graffiti = $1;", graffiti)
	if err != nil {
		return nil, err
	}
	if ret.Count == 0 {
		return nil, ErrNotFound
	}
	return ret, nil
}

func (d *DataAccessService) GetSearchAddress(ctx context.Context, chainId uint64, address []byte) (*t.SearchAddress, error) {
	eth1AddressSearchItem, err := d.bigtable.SearchForAddress(address, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search for address %s: %w", hexutil.Encode(address), err)
	}
	if len(eth1AddressSearchItem) == 0 || eth1AddressSearchItem[0] == nil {
		return nil, ErrNotFound
	}
	foundAddress := *eth1AddressSearchItem[0]
	return &t.SearchAddress{
		Address: t.Address{
			Hash: t.Hash("0x" + foundAddress.Address),
			// TODO: implement finding contract status / name
		},
	}, nil
}

func (d *DataAccessService) GetSearchTransaction(ctx context.Context, chainId uint64, transactionHash []byte) (*t.SearchTransaction, error) {
	tx, err := db.BigtableClient.GetIndexedEth1Transaction(transactionHash)
	if err != nil {
		return nil, fmt.Errorf("failed to search for transaction %s: %w", hexutil.Encode(transactionHash), err)
	}
	if tx == nil {
		return nil, ErrNotFound
	}
	return &t.SearchTransaction{
		TransactionHash: t.Hash(hexutil.Encode(tx.Hash)),
	}, nil
}

func (d *DataAccessService) GetSearchBlock(ctx context.Context, chainId uint64, blockNumber uint64) (*t.SearchBlock, error) {
	block, err := db.BigtableClient.GetBlockFromBlocksTable(blockNumber)
	if err != nil {
		if err == db.ErrBlockNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to search for block %d: %w", blockNumber, err)
	}
	if block == nil {
		return nil, fmt.Errorf("nil block returned for block number %d", blockNumber)
	}
	return &t.SearchBlock{
		BlockNumber: block.Number,
	}, nil
}

func (d *DataAccessService) GetSearchEpoch(ctx context.Context, chainId uint64, epoch uint64) (*t.SearchEpoch, error) {
	ds := goqu.Dialect("postgres").
		From("epochs").
		Select(goqu.I("epoch")).
		Where(goqu.I("epoch").Eq(epoch))

	foundEpoch, err := runQuery[uint64](ctx, d.readerDb, ds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t.SearchEpoch{
		Epoch: foundEpoch,
	}, nil
}

func (d *DataAccessService) GetSearchToken(ctx context.Context, chainId uint64, address []byte) (*t.SearchToken, error) {
	// TODO: find erc721 and erc1155 tokens
	// currently only erc20 tokens supported
	eth1AddressSearchItem, err := d.bigtable.SearchForAddress(address, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search for address %s: %w", hexutil.Encode(address), err)
	}
	if len(eth1AddressSearchItem) == 0 || eth1AddressSearchItem[0] == nil {
		return nil, ErrNotFound
	}
	foundAddress := *eth1AddressSearchItem[0]
	if foundAddress.Token == "" {
		return nil, ErrNotFound
	}
	return &t.SearchToken{
		Address: t.Address{
			Hash:       t.Hash("0x" + foundAddress.Address),
			IsContract: true,
			Label:      foundAddress.Name,
		},
		Token: foundAddress.Token,
	}, nil
}
