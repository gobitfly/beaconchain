package main

import (
	"os"

	"github.com/gobitfly/beaconchain/cmd/blobindexer"
	"github.com/gobitfly/beaconchain/cmd/eth1indexer"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal(nil, "missing target", 0)
	}
	target := os.Args[1]

	log.Info(target)
	switch target {
	case "blobindexer":
		blobindexer.Run()
	case "eth1indexer":
		eth1indexer.Run()
	}
}
