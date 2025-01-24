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

type structTest struct {
	Foo string
	Bar string
}

func TestClickHouse(t *testing.T) {
	client, _ := databasetest.NewClickHouse(t)
	clickHouse, err := NewClickHouseWithClient(client, schemaTest)
	if err != nil {
		t.Fatal(err)
	}
	if err := clickHouse.Add(tableTest, []any{
		&structTest{
			Foo: "foo",
		},
		&structTest{
			Bar: "bar",
		},
	}); err != nil {
		t.Fatal(err)
	}

	var res []structTest
	if err := clickHouse.Read("SELECT * FROM test", ScanArray(&res)); err != nil {
		t.Fatal(err)
	}
	if got, want := res[0].Foo, "foo"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := res[1].Bar, "bar"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
