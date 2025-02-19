package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	go_ens "github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"
)

type EnsCheckedDictionary struct {
	mux     sync.Mutex
	address map[common.Address]bool
	name    map[string]bool
}

func (bigtable *Bigtable) ImportEnsUpdates(client *ethclient.Client, readBatchSize int64) error {
	key := fmt.Sprintf("%s:ENS:V", bigtable.chainId)

	ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	defer done()

	rowRange := gcp_bigtable.PrefixRange(key)
	keys := []string{}

	err := bigtable.tableData.ReadRows(ctx, rowRange, func(row gcp_bigtable.Row) bool {
		row_ := row[DEFAULT_FAMILY][0]
		keys = append(keys, row_.Row)
		return true
	}, gcp_bigtable.LimitRows(readBatchSize)) // limit to max 1000 entries to avoid blocking the import of new blocks
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		log.Warnf("No ENS entries to validate")
		return nil
	}

	log.Infof("Validating %v ENS entries", len(keys))
	alreadyChecked := EnsCheckedDictionary{
		address: make(map[common.Address]bool),
		name:    make(map[string]bool),
	}

	mutDelete := gcp_bigtable.NewMutation()
	mutDelete.DeleteRow()

	batchSize := 100
	total := len(keys)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := keys[i:to]
		log.Infof("Batching ENS entries %v:%v of %v", i, to, total)

		g := new(errgroup.Group)
		g.SetLimit(10) // limit load on the node
		mutsDelete := &types.BulkMutations{
			Keys: make([]string, 0, 1),
			Muts: make([]*gcp_bigtable.Mutation, 0, 1),
		}

		for _, k := range batch {
			key := k
			var name string
			var address *common.Address
			split := strings.Split(key, ":")
			value := split[4]

			switch split[3] {
			case "H":
				// if we have a hash we look if we find a name in the db. If not we can ignore it.
				nameHash, err := hex.DecodeString(value)
				if err != nil {
					log.Error(err, fmt.Errorf("name hash could not be decoded: %v", value), 0)
				} else {
					err := ReaderDb.Get(&name, `
					SELECT
						ens_name
					FROM ens
					WHERE name_hash = $1
					`, nameHash[:])
					if err != nil && err != sql.ErrNoRows {
						return err
					}
				}
			case "A":
				addressHash, err := hex.DecodeString(value)
				if err != nil {
					log.Error(err, fmt.Errorf("address hash could not be decoded: %v", value), 0)
				} else {
					add := common.BytesToAddress(addressHash)
					address = &add
				}
			case "N":
				name = value
			}

			g.Go(func() error {
				if name != "" {
					err := validateEnsName(client, name, &alreadyChecked)
					if err != nil {
						return fmt.Errorf("error validating new name [%v]: %w", name, err)
					}
				} else if address != nil {
					err := validateEnsAddress(client, *address, &alreadyChecked)
					if err != nil {
						return fmt.Errorf("error validating new address [%v]: %w", address, err)
					}
				}
				return nil
			})

			mutsDelete.Keys = append(mutsDelete.Keys, key)
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// After processing a batch of keys we remove them from bigtable
		err = bigtable.WriteBulk(mutsDelete, bigtable.tableData, DEFAULT_BATCH_INSERTS)
		if err != nil {
			return err
		}

		// give node some time for other stuff between batches
		time.Sleep(time.Millisecond * 100)
	}

	log.Warnf("Import of ENS updates completed")
	return nil
}

func validateEnsAddress(client *ethclient.Client, address common.Address, alreadyChecked *EnsCheckedDictionary) error {
	alreadyChecked.mux.Lock()
	if alreadyChecked.address[address] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.address[address] = true
	alreadyChecked.mux.Unlock()

	names := []string{}
	err := ReaderDb.Select(&names, `SELECT ens_name FROM ens WHERE address = $1 AND is_primary_name AND valid_to >= now()`, address.Bytes())
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	for _, name := range names {
		if name != "" {
			err = validateEnsName(client, name, alreadyChecked)
			if err != nil {
				return err
			}
		}
		reverseName, err := go_ens.ReverseResolve(client, address)
		if err != nil {
			if err.Error() == "not a resolver" ||
				err.Error() == "no resolution" ||
				err.Error() == "execution reverted" ||
				strings.HasPrefix(err.Error(), "name is not valid") {
				// log.Warnf("reverse resolving address [%v] resulted in a skippable error [%s], skipping it", address, err.Error())
			} else {
				return fmt.Errorf("error could not reverse resolve address [%v]: %w", address, err)
			}
		}

		if reverseName != name {
			err = validateEnsName(client, reverseName, alreadyChecked)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEnsName(client *ethclient.Client, name string, alreadyChecked *EnsCheckedDictionary) error {
	if name == "" || name == ".eth" {
		return nil
	}
	// For now only .eth is supported other ens domains use different techniques and require and individual implementation
	if !strings.HasSuffix(name, ".eth") {
		name = fmt.Sprintf("%s.eth", name)
	}
	alreadyChecked.mux.Lock()
	if alreadyChecked.name[name] {
		alreadyChecked.mux.Unlock()
		return nil
	}
	alreadyChecked.name[name] = true
	alreadyChecked.mux.Unlock()

	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_validate_ens_name").Observe(time.Since(startTime).Seconds())
	}()

	nameHash, err := go_ens.NameHash(name)
	if err != nil {
		// log.Warnf("error could not hash name [%v]: %v -> removing ens entry", name, err)
		err = removeEnsName(name)
		if err != nil {
			return fmt.Errorf("error removing ens name [%v]: %w", name, err)
		}
		return nil
	}

	addr, err := go_ens.Resolve(client, name)
	if err != nil {
		if err.Error() == "unregistered name" ||
			err.Error() == "no address" ||
			err.Error() == "no resolver" ||
			err.Error() == "abi: attempting to unmarshal an empty string while arguments are expected" ||
			strings.Contains(err.Error(), "execution reverted") ||
			err.Error() == "invalid jump destination" ||
			err.Error() == "invalid opcode: INVALID" {
			// the given name is not available anymore or resolving it did not work properly => we can remove it from the db (if it is there)
			// log.Warnf("could not resolve name [%v]: %v -> removing ens entry", name, err)
			err = removeEnsName(name)
			if err != nil {
				return fmt.Errorf("error removing ens name after resolve failed [%v]: %w", name, err)
			}
			return nil
		}
		return fmt.Errorf("error could not resolve name [%v]: %w", name, err)
	}

	// we need to get the main domain to get the expiration date
	parts := strings.Split(name, ".")
	mainName := strings.Join(parts[len(parts)-2:], ".")

	expires, err := GetEnsExpiration(client, mainName)
	if err != nil {
		return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
	}

	isPrimary := false
	reverseName, err := go_ens.ReverseResolve(client, addr)
	if err != nil {
		if err.Error() == "not a resolver" || err.Error() == "no resolution" || err.Error() == "execution reverted" {
			// log.Warnf("reverse resolving address [%v] for name [%v] resulted in an error [%s], marking entry as not primary", addr, name, err.Error())
		} else {
			return fmt.Errorf("error could not reverse resolve address [%v]: %w", addr, err)
		}
	}
	if reverseName == name {
		isPrimary = true
	}

	_, err = WriterDb.Exec(`
	INSERT INTO ens (
		name_hash, 
		ens_name, 
		address,
		is_primary_name, 
		valid_to)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT 
		(name_hash) 
	DO UPDATE SET 
		ens_name = excluded.ens_name,
		address = excluded.address,
		is_primary_name = excluded.is_primary_name,
		valid_to = excluded.valid_to
	`, nameHash[:], name, addr.Bytes(), isPrimary, expires)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
			log.Warnf("could not insert ens name [%v]: %v", name, err)
			return nil
		}
		return fmt.Errorf("error writing ens data for name [%v]: %w", name, err)
	}

	// log.InfoWithFields(log.Fields{
	// 	"name":        name,
	// 	"address":     addr,
	// 	"expires":     expires,
	// 	"reverseName": reverseName,
	// }, "validated ens name")
	return nil
}

func GetEnsExpiration(client *ethclient.Client, name string) (time.Time, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_get_expiration").Observe(time.Since(startTime).Seconds())
	}()

	normName, err := go_ens.NormaliseDomain(name)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NormaliseDomain: %w", err)
	}
	domain := go_ens.Domain(normName)
	label, err := go_ens.DomainPart(normName, 1)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.DomainPart: %w", err)
	}
	registrar, err := go_ens.NewBaseRegistrar(client, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NewBaseRegistrar: %w", err)
	}
	uqName, err := go_ens.UnqualifiedName(label, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.UnqualifiedName: %w", err)
	}
	labelHash, err := go_ens.LabelHash(uqName)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.LabelHash: %w", err)
	}
	id := new(big.Int).SetBytes(labelHash[:])
	ts, err := registrar.Contract.NameExpires(&bind.CallOpts{}, id)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts.Int64(), 0), nil
}

func GetAddressForEnsName(name string) (address *common.Address, err error) {
	addressBytes := []byte{}
	err = ReaderDb.Get(&addressBytes, `
	SELECT address
	FROM ens
	WHERE
		ens_name = $1 AND
		valid_to >= now()
	`, name)
	if err == nil && addressBytes != nil {
		add := common.BytesToAddress(addressBytes)
		address = &add
	}
	return address, err
}

// pass invalid time to get latest data
func GetEnsNameForAddress(address common.Address, validUntil time.Time) (name string, err error) {
	if validUntil.IsZero() {
		validUntil = time.Now()
	}
	err = ReaderDb.Get(&name, `
	SELECT ens_name
	FROM ens
	WHERE
		address = $1 AND
		is_primary_name AND
		valid_to >= $2
	;`, address.Bytes(), validUntil)
	return name, err
}

func GetEnsNamesForAddresses(addressMap map[string]string) error {
	if len(addressMap) == 0 {
		return nil
	}
	type pair struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}
	dbAddresses := []pair{}
	addresses := make([][]byte, 0, len(addressMap))
	for address := range addressMap {
		add, err := hexutil.Decode(address)
		if err != nil {
			return err
		}
		addresses = append(addresses, add)
	}

	err := ReaderDb.Select(&dbAddresses, `
	SELECT address, ens_name
	FROM ens
	WHERE
		address = ANY($1) AND
		is_primary_name AND
		valid_to >= now()
	;`, addresses)
	if err != nil {
		return err
	}
	for _, foundling := range dbAddresses {
		addressMap[hexutil.Encode(foundling.Address)] = foundling.EnsName
	}
	return nil
}

func removeEnsName(name string) error {
	_, err := WriterDb.Exec(`
	DELETE FROM ens
	WHERE
		ens_name = $1
	;`, name)
	if err != nil && strings.Contains(fmt.Sprintf("%v", err), "invalid byte sequence") {
		log.Warnf("could not delete ens name [%v]: %v", name, err)
		return nil
	} else if err != nil {
		return fmt.Errorf("error deleting ens name [%v]: %v", name, err)
	}
	log.Infof("Ens name removed from db: %v", name)
	return nil
}
