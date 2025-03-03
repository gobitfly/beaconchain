package db2

import (
	"crypto/sha256"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

func TestENSStore(t *testing.T) {
	store := NewENSStore(databasetest.NewPostgres(t))

	if err := store.SetENS(validEns); err != nil {
		t.Fatal(err)
	}

	t.Run("GetAllENS", func(t *testing.T) {
		list, err := store.GetAllENS()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := list, []ENS{validEns}; reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("GetENSNameFromHash", func(t *testing.T) {
		ens, err := store.GetENSNameFromHash(validEns.NameHash)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := ens, validEns.Name; got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("GetNamesForAddress", func(t *testing.T) {
		ens, err := store.GetNamesForAddress(validEns.Address)
		if err != nil {
			t.Fatal(err)
		}
		if len(ens) != 1 {
			t.Fatalf("got %d, want 1", len(ens))
		}
		if got, want := ens[0], validEns.Name; got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("DeleteENS", func(t *testing.T) {
		if err := store.DeleteENS(validEns.Name); err != nil {
			t.Fatal(err)
		}
		res, err := store.GetENSNameFromHash(validEns.NameHash)
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 0 {
			t.Errorf("ens was not deleted")
		}
	})
}

var validEns = ENS{
	NameHash:  toByte32(sha256.New().Sum([]byte("testEns"))),
	Name:      "testEns",
	Address:   common.HexToAddress("0x000000000000000000000000000000000000abba"),
	IsPrimary: true,
	Expires:   time.Now().Add(time.Hour),
}
