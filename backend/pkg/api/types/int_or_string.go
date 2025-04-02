package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
)

// IntOrString is a custom type that can be unmarshalled from either an int or a string (strings will also be parsed to int if possible).
// if unmarshaling throws no errors one of the two fields will be set, the other will be nil.
type IntOrString struct {
	IntValue *uint64
	StrValue *string
}

func (v *IntOrString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return fmt.Errorf("null value not allowed")
	}
	// Attempt to unmarshal as uint64 first
	var intValue uint64
	if err := json.Unmarshal(data, &intValue); err == nil {
		v.IntValue = &intValue
		return nil
	}

	// If unmarshalling as uint64 fails, try to unmarshal as string
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		strValue = strings.TrimSpace(strValue)
		if parsedInt, err := strconv.ParseUint(strValue, 10, 64); err == nil {
			v.IntValue = &parsedInt
		} else {
			v.StrValue = &strValue
		}
		return nil
	}

	// If both unmarshalling attempts fail, return an error
	return fmt.Errorf("failed to unmarshal IntOrString from json: %s", string(data))
}

func (v IntOrString) String() string {
	if v.IntValue != nil {
		return strconv.FormatUint(*v.IntValue, 10)
	}
	if v.StrValue != nil {
		return *v.StrValue
	}
	return ""
}

func (IntOrString) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string"}, {Type: "integer"},
		},
	}
}
