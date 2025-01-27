package database

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

const schemaTest = `
		CREATE TABLE test (
			  Foo String
			, Bar String
		) Engine = Memory
`
const tableTest = "test"
const querySelectAll = "SELECT * FROM test"
const querySelectFirst = "SELECT TOP 1 * FROM test"

type structTest struct {
	Foo string
	Bar string
}

func TestClickHouse(t *testing.T) {
	var tests = []struct {
		name       string
		query      string
		records    []*structTest
		wantRecord *int
	}{
		{
			name:  "on record",
			query: querySelectAll,
			records: []*structTest{
				{Foo: "foo"},
			},
		},
		{
			name:  "two records",
			query: querySelectAll,
			records: []*structTest{
				{Foo: "foo"},
				{Bar: "bar"},
			},
		},
		{
			name:  "two records select first",
			query: querySelectFirst,
			records: []*structTest{
				{Foo: "foo"},
				{Bar: "bar"},
			},
			wantRecord: toPointer(0),
		},
		{
			name:  "two records with where",
			query: `SELECT * FROM test WHERE Bar = 'bar'`,
			records: []*structTest{
				{Foo: "foo"},
				{Bar: "bar"},
			},
			wantRecord: toPointer(1),
		},
	}
	client, _ := databasetest.NewClickHouse(t)
	clickHouse, err := NewClickHouseWithClient(client, schemaTest)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := clickHouse.clearTable(tableTest); err != nil {
					t.Fatal(err)
				}
			}()
			if err := clickHouse.Add(tableTest, toAnyArray(tt.records)); err != nil {
				t.Fatal(err)
			}

			var records []structTest
			if err := clickHouse.Read(tt.query, ScanArray(&records)); err != nil {
				t.Fatal(err)
			}
			if tt.wantRecord != nil {
				if len(records) != 1 {
					t.Fatalf("got %d records, want 1", len(records))
				}
				tt.records = []*structTest{tt.records[*tt.wantRecord]}
			}
			for i, record := range records {
				if got, want := record.Foo, tt.records[i].Foo; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				if got, want := record.Bar, tt.records[i].Bar; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			}
		})
	}
}

func toAnyArray[T any](array []T) []any {
	ret := make([]any, len(array))
	for i, v := range array {
		ret[i] = v
	}
	return ret
}

func toPointer[T any](i T) *T {
	return &i
}
