package executionlayer

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
)

type ENSUpdateStore interface {
	GetENSUpdate(chainID string, batchSize int64) ([]db2.ENSLog, error)
	DeleteENSUpdate(chainID string, logs []db2.ENSLog) error
}

type ENSStore interface {
	GetENSNameFromHash(nameHash [32]byte) (string, error)
	GetNamesForAddress(address common.Address) ([]string, error)
	SetENS(ens db2.ENS) error
	DeleteENS(name string) error
}

type ENS interface {
	Expiration(name string) (time.Time, error)
	Resolve(name string) (common.Address, error)
	ReverseResolve(address common.Address) (string, error)
}

type ENSImporter struct {
	updates ENSUpdateStore
	store   ENSStore
	ens     ENS
}

func NewENSImporter(updates ENSUpdateStore, store ENSStore, ens ENS) *ENSImporter {
	return &ENSImporter{
		updates: updates,
		store:   store,
		ens:     ens,
	}
}

type EnsImportResult struct {
	updated []string
	deleted []string
}

func (importer ENSImporter) Import(chainID string, readBatchSize int64) (*EnsImportResult, error) {
	updates, err := importer.updates.GetENSUpdate(chainID, readBatchSize)
	if err != nil {
		return nil, err
	}

	batchSize := 100
	total := len(updates)
	checked := newENSChecked()

	var all EnsImportResult
	var mu sync.Mutex
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := updates[i:to]

		g := new(errgroup.Group)
		g.SetLimit(10) // limit load on the node

		for _, ensLog := range batch {
			g.Go(func() error {
				res, err := importer.importENS(ensLog, checked)
				if err != nil {
					return err
				}
				mu.Lock()
				all.updated = append(all.updated, res.updated...)
				all.deleted = append(all.deleted, res.deleted...)
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}

		// after processing a batch of keys we remove them from the update store
		if err := importer.updates.DeleteENSUpdate(chainID, batch); err != nil {
			return nil, err
		}

		// give node some time for other stuff between batches
		time.Sleep(time.Millisecond * 100)
	}

	return &all, nil
}

func (importer ENSImporter) importENS(ensLog db2.ENSLog, checked *ensChecked) (*EnsImportResult, error) {
	var names []string
	if ensLog.Node != nil {
		name, err := importer.store.GetENSNameFromHash(*ensLog.Node)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	if ensLog.Owner != nil {
		addressNames, err := importer.getEnsNamesForAddress(*ensLog.Owner, checked)
		if err != nil {
			return nil, fmt.Errorf("error getting names for new address [%v]: %w", *ensLog.Owner, err)
		}
		names = append(names, addressNames...)
	}

	if ensLog.Name != nil {
		names = append(names, *ensLog.Name)
	}
	var res EnsImportResult
	for _, name := range names {
		deleteName, err := importer.validateEnsName(name, checked)
		if err != nil {
			return nil, fmt.Errorf("error validating new name [%v]: %w", name, err)
		}
		if deleteName {
			res.deleted = append(res.deleted, name)
			if err := importer.store.DeleteENS(name); err != nil {
				return nil, fmt.Errorf("error removing ens name [%v]: %w", name, err)
			}
			continue
		}
		res.updated = append(res.updated, name)
	}
	return &res, nil
}

func (importer ENSImporter) validateEnsName(name string, alreadyChecked *ensChecked) (bool, error) {
	if name == "" || name == ".eth" {
		return false, nil
	}
	// For now only .eth is supported other ens domains use different techniques and require and individual implementation
	if !strings.HasSuffix(name, ".eth") {
		name = fmt.Sprintf("%s.eth", name)
	}
	if alreadyChecked.NameAlreadyChecked(name) {
		return false, nil
	}

	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_validate_ens_name").Observe(time.Since(startTime).Seconds())
	}()

	nameHash, err := ens.NameHash(name)
	if err != nil {
		return true, nil
	}

	address, err := importer.ens.Resolve(name)
	if err != nil {
		if ignoreResolveError(err) {
			return true, nil
		}
		return false, fmt.Errorf("error could not resolve name [%v]: %w", name, err)
	}

	// we need to get the main domain to get the expiration date
	parts := strings.Split(name, ".")
	mainName := strings.Join(parts[len(parts)-2:], ".")

	expires, err := importer.ens.Expiration(mainName)
	if err != nil {
		return false, fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
	}

	isPrimary := false
	reverseName, err := importer.ens.ReverseResolve(address)
	if err != nil && !ignoreReverseResolveError(err) {
		return false, fmt.Errorf("error could not reverse resolve address [%v]: %w", address, err)
	}
	if reverseName == name {
		isPrimary = true
	}

	return false, importer.store.SetENS(db2.ENS{
		NameHash:  nameHash,
		Name:      name,
		Address:   address,
		IsPrimary: isPrimary,
		Expires:   expires,
	})
}

func (importer ENSImporter) getEnsNamesForAddress(address common.Address, alreadyChecked *ensChecked) ([]string, error) {
	if alreadyChecked.AddressAlreadyChecked(address) {
		return nil, nil
	}

	names, err := importer.store.GetNamesForAddress(address)
	if err != nil {
		return nil, err
	}

	var ensNames []string
	for _, name := range names {
		if name != "" {
			ensNames = append(ensNames, name)
		}
		reverseName, err := importer.ens.ReverseResolve(address)
		if err != nil && !ignoreReverseResolveError(err) {
			return nil, fmt.Errorf("error could not reverse resolve address [%v]: %w", address, err)
		}
		if reverseName != name {
			ensNames = append(ensNames, reverseName)
		}
	}
	return ensNames, nil
}

func ignoreResolveError(err error) bool {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unregistered name"):
	case strings.Contains(msg, "no address"):
	case strings.Contains(msg, "no resolver"):
	case strings.Contains(msg, "abi: attempting to unmarshal an empty string while arguments are expected"):
	case strings.Contains(msg, "execution reverted"):
	case strings.Contains(msg, "invalid jump destination"):
	case strings.Contains(msg, "invalid opcode: INVALID"):
	// the given name is not available anymore or resolving it did not work properly => we can remove it from the db (if it is there)
	default:
		return false
	}
	return true
}

func ignoreReverseResolveError(err error) bool {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not a resolver"):
	case strings.Contains(msg, "no resolution"):
	case strings.Contains(msg, "execution reverted"):
	case strings.Contains(msg, "name is not valid"):
	default:
		return false
	}
	return true
}

type ensChecked struct {
	mux     sync.Mutex
	address map[common.Address]bool
	name    map[string]bool
}

func newENSChecked() *ensChecked {
	return &ensChecked{
		address: make(map[common.Address]bool),
		name:    make(map[string]bool),
	}
}

func (e *ensChecked) NameAlreadyChecked(name string) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.name[name] {
		return true
	}
	e.name[name] = true
	return false
}

func (e *ensChecked) AddressAlreadyChecked(address common.Address) bool {
	e.mux.Lock()
	defer e.mux.Unlock()
	if e.address[address] {
		return true
	}
	e.address[address] = true
	return false
}

type EnsContracts struct {
	client bind.ContractBackend
}

func NewEnsContracts(client bind.ContractBackend) EnsContracts {
	return EnsContracts{
		client: client,
	}
}

func (e EnsContracts) Expiration(name string) (time.Time, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ens_get_expiration").Observe(time.Since(startTime).Seconds())
	}()

	normName, err := ens.NormaliseDomain(name)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NormaliseDomain: %w", err)
	}
	domain := ens.Domain(normName)
	label, err := ens.DomainPart(normName, 1)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.DomainPart: %w", err)
	}
	uqName, err := ens.UnqualifiedName(label, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.UnqualifiedName: %w", err)
	}
	labelHash, err := ens.LabelHash(uqName)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.LabelHash: %w", err)
	}
	id := new(big.Int).SetBytes(labelHash[:])
	registrar, err := ens.NewBaseRegistrar(e.client, domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error calling go_ens.NewBaseRegistrar: %w", err)
	}
	ts, err := registrar.Contract.NameExpires(&bind.CallOpts{}, id)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts.Int64(), 0), nil
}

func (e EnsContracts) Resolve(name string) (common.Address, error) {
	return ens.Resolve(e.client, name)
}

func (e EnsContracts) ReverseResolve(address common.Address) (string, error) {
	return ens.ReverseResolve(e.client, address)
}
