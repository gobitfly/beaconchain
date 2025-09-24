package externalspec

import (
	_ "embed"
)

//go:embed openapi3.yaml
var RawBytes []byte
