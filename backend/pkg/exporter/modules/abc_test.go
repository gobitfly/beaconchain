package modules

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/consapi"
)

func TestABC(t *testing.T) {

	jsonStr := `{"Name":"John", "Age":"20"}`

	var target testStruct
	err := json.Unmarshal([]byte(jsonStr), &target)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("age: %v\n", *target.Age)
}

type testStruct struct {
	Name string `json:"Name"`
	Age  *int   `json:"Age,string"`
}

func TestSpec(t *testing.T) {
	consApi := consapi.NewClient("http://localhost:14000")
	spec, err := consApi.GetSpec()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("spec: %v\n", *spec.Data.AltairForkEpoch)
}

func TestMail(t *testing.T) {

}
