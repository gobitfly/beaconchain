package executionlayer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/exp/maps"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

func TestENSImporter(t *testing.T) {
	store := db2.NewENSStore(databasetest.NewPostgres(t))

	tests := []struct {
		name         string
		store        []db2.ENS
		update       db2.ENSLog
		expected     []db2.ENS
		ensContracts fakeEnsContracts
	}{
		{
			name: "node update",
			store: []db2.ENS{
				{
					NameHash: mustNameHash("foobar.eth"),
					Name:     "foobar.eth",
				},
			},
			update: db2.ENSLog{Node: toPointer(mustNameHash("foobar.eth"))},
			expected: []db2.ENS{
				{
					Address: common.HexToAddress("0x0c79845b706290cfb2b91aa526bc984f6a6c524c"),
				},
			},
			ensContracts: fakeEnsContracts{
				expiration:     time.Now().Add(time.Hour),
				resolve:        common.HexToAddress("0x0C79845b706290cFb2B91aa526bc984f6A6C524c"),
				reverseResolve: "foobar.eth",
			},
		},
		{
			name: "owner update",
			store: []db2.ENS{
				{
					NameHash:  mustNameHash("foo.eth"),
					Name:      "foo.eth",
					Address:   common.HexToAddress("0xfD669c4802dd1DBB40830431773a71Bb910A40F9"),
					IsPrimary: true,
					Expires:   time.Now().Add(time.Hour),
				},
			},
			update: db2.ENSLog{Owner: toPointer(common.HexToAddress("0xfD669c4802dd1DBB40830431773a71Bb910A40F9"))},
			expected: []db2.ENS{
				{
					Name:    "foo.eth",
					Address: common.HexToAddress("0x0C79845b706290cFb2B91aa526bc984f6A6C524c"),
				},
			},
			ensContracts: fakeEnsContracts{
				expiration:        time.Now().Add(time.Hour),
				resolve:           common.HexToAddress("0x0C79845b706290cFb2B91aa526bc984f6A6C524c"),
				reverseResolveErr: fmt.Errorf("not a resolver"),
				reverseResolve:    "foo.eth",
			},
		},
		{
			name: "name update",
			store: []db2.ENS{
				{
					NameHash:  mustNameHash("foo.eth"),
					Name:      "foo.eth",
					IsPrimary: true,
					Expires:   time.Now().Add(time.Hour),
				},
			},
			update: db2.ENSLog{Name: toPointer("foo.eth")},
			expected: []db2.ENS{
				{
					Name:    "foo.eth",
					Address: common.HexToAddress("0x0C79845b706290cFb2B91aa526bc984f6A6C524c"),
				},
			},
			ensContracts: fakeEnsContracts{
				expiration:     time.Now().Add(time.Hour),
				resolve:        common.HexToAddress("0x0C79845b706290cFb2B91aa526bc984f6A6C524c"),
				reverseResolve: "foo.eth",
			},
		},
		{
			name: "remove ens entry after on chain deletion",
			store: []db2.ENS{
				{
					Name:      "foo.eth",
					IsPrimary: true,
					Expires:   time.Now().Add(time.Hour),
				},
			},
			update: db2.ENSLog{Name: toPointer("foo.eth")},
			ensContracts: fakeEnsContracts{
				expiration: time.Now().Add(time.Hour),
				resolveErr: fmt.Errorf("unregistered name"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ens := range tt.store {
				if err := store.SetENS(ens); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = store.DeleteENS(ens.Name)
				})
			}
			updates := newFakeENSUpdateStore([]db2.ENSLog{tt.update})
			importer := NewENSImporter(updates, store, &tt.ensContracts)
			if _, err := importer.Import("idontcare", 100); err != nil {
				t.Fatal(err)
			}
			if len(updates.updates) != 0 {
				t.Fatal("updates store is not empty")
			}
			all, err := store.GetAllENS()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := len(all), len(tt.expected); got != want {
				t.Fatalf("got %d entries, expected %d", got, want)
			}
			for i, expected := range tt.expected {
				if expected.Name != "" {
					if got, want := all[i].Name, expected.Name; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
				if expected.Address.Cmp(common.Address{}) != 0 {
					if got, want := all[i].Address, expected.Address; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}
			}
		})
	}
}

type fakeENSUpdateStore struct {
	updates map[string]db2.ENSLog
}

func newFakeENSUpdateStore(logs []db2.ENSLog) fakeENSUpdateStore {
	updates := make(map[string]db2.ENSLog)
	for _, log := range logs {
		updates[logID(log)] = log
	}
	return fakeENSUpdateStore{updates: updates}
}

func (s fakeENSUpdateStore) GetENSUpdate(chainID string, batchSize int64) ([]db2.ENSLog, error) {
	return maps.Values(s.updates), nil
}

func (s fakeENSUpdateStore) DeleteENSUpdate(chainID string, logs []db2.ENSLog) error {
	for _, log := range logs {
		delete(s.updates, logID(log))
	}
	return nil
}

func logID(log db2.ENSLog) string {
	var id string
	switch {
	case log.Owner != nil:
		id = log.Owner.Hex()
	case log.Name != nil:
		id = *log.Name
	case log.Node != nil:
		id = hex.EncodeToString(log.Node[:])
	}
	return id
}

func toPointer[T any](source T) *T {
	return &source
}

func mustNameHash(name string) [32]byte {
	hash, _ := ens.NameHash(name)
	return hash
}

type fakeEnsContracts struct {
	expiration        time.Time
	resolve           common.Address
	resolveErr        error
	reverseResolve    string
	reverseResolveErr error
}

func (f *fakeEnsContracts) Expiration(name string) (time.Time, error) {
	return f.expiration, nil
}

func (f *fakeEnsContracts) Resolve(name string) (common.Address, error) {
	return f.resolve, f.resolveErr
}

func (f *fakeEnsContracts) ReverseResolve(address common.Address) (string, error) {
	if f.reverseResolveErr != nil {
		err := f.reverseResolveErr
		f.reverseResolveErr = nil
		return "", err
	}
	return f.reverseResolve, nil
}

// TestENSContracts use stubbed data, where only the owner address and the name were changed
func TestENSContracts(t *testing.T) {
	t.Run("expiration", func(t *testing.T) {
		timestampHex := "67941590"
		backend := stubCallContract{
			resp: map[string][]byte{
				"02571be393cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae": hexutil.MustDecode("0x00000000000000000000000057f1887a8bf19b14fc0df6fd9b2acc9af147ea85"),
				"01ffc9a728ed4f6c00000000000000000000000000000000000000000000000000000000": hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000001"),
				"d6e4fa8641b1a0649752af1b28b3dc29a1556eee781e4a4c3a1f7f53f90fa834de098c4d": hexutil.MustDecode(fmt.Sprintf("0x00000000000000000000000000000000000000000000000000000000%s", timestampHex)),
			},
		}
		ensContracts := NewEnsContracts(backend)
		expiration, err := ensContracts.Expiration("foo.eth")
		if err != nil {
			t.Fatal(err)
		}

		expected := new(big.Int).SetBytes(hexutil.MustDecode("0x" + timestampHex)).Int64()
		if got, want := expiration, time.Unix(expected, 0); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("resolve", func(t *testing.T) {
		expected := common.HexToAddress("0x000000000000000000000000000000000000beef")
		backend := stubCallContract{
			resp: map[string][]byte{
				"02571be3de9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f": hexutil.MustDecode("0x000000000000000000000000000000000000000000000000000000000000beef"),
				"0178b8bfde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f": hexutil.MustDecode("0x000000000000000000000000231b0ee14048e9dccd1d247744d114a4eb5e8e63"),
				"3b3b57deeb4f647bea6caa36333c816d7b46fdcb05f9466ecacc140ea8c66faf15b3d9f1": hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000000"),
				"3b3b57dede9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f": hexutil.MustDecode("0x000000000000000000000000000000000000000000000000000000000000beef"),
			},
		}
		ensContracts := NewEnsContracts(backend)
		address, err := ensContracts.Resolve("foo.eth")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := address.Hex(), expected.Hex(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("reverseResolve", func(t *testing.T) {
		backend := stubCallContract{
			resp: map[string][]byte{
				"0178b8bff4843d9c05bd7abb6a1e82c8b55312ce19dd3720c0eaac900b2223b50f9238ed": hexutil.MustDecode("0x000000000000000000000000231b0ee14048e9dccd1d247744d114a4eb5e8e63"),
				"691f343168d3cf674cfc1dbcea90ac43bb06b3fecc47b8fc849438205d8c638ad936604c": hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000"),
				"691f3431f4843d9c05bd7abb6a1e82c8b55312ce19dd3720c0eaac900b2223b50f9238ed": hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000007666f6f2e65746800000000000000000000000000000000000000000000000000"),
			},
		}
		ensContracts := NewEnsContracts(backend)
		name, err := ensContracts.ReverseResolve(common.HexToAddress("0x000000000000000000000000000000000000beef"))
		if err != nil {
			t.Fatal(err)
		}
		expected := "foo.eth"
		if got, want := name, expected; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

type stubCallContract struct {
	bind.ContractBackend
	resp map[string][]byte
}

func (stub stubCallContract) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return stub.resp[hex.EncodeToString(call.Data)], nil
}
