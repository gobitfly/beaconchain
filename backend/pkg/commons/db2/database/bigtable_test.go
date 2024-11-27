package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"cloud.google.com/go/bigtable/bttest"
	"golang.org/x/exp/maps"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/databasetest"
)

func TestNewBigTable(t *testing.T) {
	srv, err := bttest.NewServer("localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	testTable, testFamily := "table", "family"
	bt, err := NewBigTable("testProject", "testInstance", map[string][]string{testTable: {testFamily}}, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	tableInfo, err := bt.admin.TableInfo(context.Background(), testTable)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := tableInfo.Families, []string{testFamily}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBigTable(t *testing.T) {
	tests := []struct {
		name     string
		bulk     bool
		items    map[string][]Item
		expected []string
	}{
		{
			name: "simple add",
			items: map[string][]Item{
				"foo": {
					{Column: "bar", Data: []byte("foobar")},
				},
			},
			expected: []string{"foobar"},
		},
		{
			name: "bulk add",
			bulk: true,
			items: map[string][]Item{
				"key1": {
					{Column: "col1", Data: []byte("foobar")},
				},
				"key2": {
					{Column: "col2", Data: []byte("foobar")},
				},
				"key3": {
					{Column: "col3", Data: []byte("foobar")},
				},
			},
			expected: []string{"foobar", "foobar", "foobar"},
		},
		{
			name: "with a prefix",
			items: map[string][]Item{
				"foo":       {{}},
				"foofoo":    {{}},
				"foofoofoo": {{}},
				"bar":       {{}},
			},
			expected: []string{"", "", "", ""},
		},
	}
	tables := map[string][]string{"testTable": {""}}
	client, admin := databasetest.NewBigTable(t)
	bt, err := NewBigTableWithClient(context.Background(), client, admin, tables)
	if err != nil {
		t.Fatal(err)
	}
	db := Wrap(bt, "testTable")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				_ = db.Clear()
			}()

			if err := db.BulkAdd(tt.items, WithBatchSize(2)); err != nil {
				t.Error(err)
			}

			t.Run("Read", func(t *testing.T) {
				res, err := db.Read("")
				if err != nil {
					t.Error(err)
				}
				if got, want := len(res), len(tt.expected); got != want {
					t.Errorf("got %v want %v", got, want)
				}
				for _, row := range res {
					for _, v := range row.Values {
						if !slices.Contains(tt.expected, string(v)) {
							t.Errorf("wrong data %s", row)
						}
					}
				}
			})

			t.Run("GetLatestValue", func(t *testing.T) {
				for key, items := range tt.items {
					v, err := db.GetLatestValue(key)
					if err != nil {
						t.Error(err)
					}
					for _, it := range items {
						if got, want := string(v.Values[fmt.Sprintf("%s:%s", it.Family, it.Column)]), string(it.Data); got != want {
							t.Errorf("got %v want %v", got, want)
						}
					}
				}
			})

			t.Run("GetRowKeys", func(t *testing.T) {
				for key := range tt.items {
					keys, err := db.GetRowKeys(key)
					if err != nil {
						t.Error(err)
					}
					count, found := 0, false
					for expectedKey := range tt.items {
						if !strings.HasPrefix(expectedKey, key) {
							continue
						}
						// don't count duplicate inputs since the add prevent duplicate keys
						if expectedKey == key && found {
							continue
						}
						found = expectedKey == key
						count++
						if !slices.Contains(keys, expectedKey) {
							t.Errorf("missing %v in %v", expectedKey, keys)
						}
					}
					if got, want := len(keys), count; got != want {
						t.Errorf("got %v want %v", got, want)
					}
				}
			})

			t.Run("GetRow", func(t *testing.T) {
				for key, items := range tt.items {
					row, err := db.GetRow(key)
					if err != nil {
						t.Error(err)
					}
					if got, want := row.Key, key; got != want {
						t.Errorf("got %v want %v", got, want)
					}
					for _, it := range items {
						if got, want := string(row.Values[fmt.Sprintf("%s:%s", it.Family, it.Column)]), string(it.Data); got != want {
							t.Errorf("got %v want %v", got, want)
						}
					}
				}
				_, err := db.GetRow("key does not exist")
				if !errors.Is(err, ErrNotFound) {
					t.Errorf("expected ErrNotFound got %v", err)
				}
			})

			t.Run("GetRowsWithKeys", func(t *testing.T) {
				rows, err := db.GetRowsWithKeys(maps.Keys(tt.items))
				if err != nil {
					t.Error(err)
				}
				for _, row := range rows {
					expected := make(map[string][]byte)
					for i := 0; i < len(tt.items[row.Key]); i++ {
						it := tt.items[row.Key][i]
						expected[fmt.Sprintf("%s:%s", it.Family, it.Column)] = it.Data
					}
					if got, want := row.Values, expected; !maps.EqualFunc(got, want, func(b1 []byte, b2 []byte) bool { return bytes.Equal(b1, b2) }) {
						t.Errorf("got %v want %v", got, want)
					}
				}
			})
		})
	}

	if err := db.Close(); err != nil {
		t.Errorf("cannot close db: %v", err)
	}
}

func TestGetRowsRange(t *testing.T) {
	tables := map[string][]string{"testTable": {""}}
	client, admin := databasetest.NewBigTable(t)
	bt, err := NewBigTableWithClient(context.Background(), client, admin, tables)
	if err != nil {
		t.Fatal(err)
	}
	db := Wrap(bt, "testTable")

	tests := []struct {
		name     string
		txs      int // must be inferior or equal to 10, otherwise padding is necessary
		expected int
		options  []Option
	}{
		{
			name:     "closed range",
			txs:      3,
			expected: 3,
		},
		{
			name:     "open range",
			txs:      3,
			expected: 1,
			options:  []Option{WithOpenRange(true)},
		},
		{
			name:     "open close range",
			txs:      3,
			expected: 2,
			options:  []Option{WithOpenCloseRange(true)},
		},
		{
			name:     "with limit",
			txs:      9,
			expected: 5,
			options:  []Option{WithLimit(5)},
		},
		{
			name:     "with stats",
			txs:      1,
			expected: 1,
			options:  []Option{WithStats(func(msg string, args ...any) {})},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			high, low := "", ""
			for i := 0; i < tt.txs; i++ {
				key := fmt.Sprintf("%d", 9-i)
				if i == 0 {
					high = key
				}
				if i == tt.txs-1 {
					low = key
				}
				_ = db.Add(key, Item{}, false)
			}
			rows, err := db.GetRowsRange(high, low, tt.options...)
			if err != nil {
				t.Error(err)
			}
			if got, want := len(rows), tt.expected; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
