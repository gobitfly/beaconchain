package database

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/databasetest"
)

func TestRemote(t *testing.T) {
	tables := map[string][]string{"testTable": {""}}
	btClient, admin := databasetest.NewBigTable(t)
	bt, err := NewBigTableWithClient(context.Background(), btClient, admin, tables)
	if err != nil {
		t.Fatal(err)
	}
	db := Wrap(bt, "testTable")

	remote := NewRemote(db)
	server := httptest.NewServer(remote.Routes())
	defer server.Close()

	client := NewRemoteClient(server.URL)

	items := map[string]Item{
		"foo": {},
		"bar": {},
	}
	for key, item := range items {
		if err := db.Add(key, item, false); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Read", func(t *testing.T) {
		res, err := client.Read("")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(res), len(items); got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("GetRow", func(t *testing.T) {
		row, err := client.GetRow("foo")
		if err != nil {
			t.Fatal(err)
		}
		if row == nil {
			t.Fatal("row is nil")
		}

		if _, err := client.GetRow("key does not exist"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound got %v", err)
		}
	})

	t.Run("GetRowsRange", func(t *testing.T) {
		rows, err := client.GetRowsRange("foo", "bar")
		if err != nil {
			t.Fatal(err)
		}
		if rows == nil {
			t.Fatal("rows is nil")
		}

		if _, err := client.GetRowsRange("0", "1"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound got %v", err)
		}
	})

	t.Run("GetRowsWithKeys", func(t *testing.T) {
		rows, err := client.GetRowsWithKeys([]string{"foo", "bar"})
		if err != nil {
			t.Fatal(err)
		}
		if rows == nil {
			t.Fatal("rows is nil")
		}
	})
}
