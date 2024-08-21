package main

import (
	"os"

	"github.com/gobitfly/beaconchain/cmd/api"
	"github.com/gobitfly/beaconchain/cmd/blobindexer"
	"github.com/gobitfly/beaconchain/cmd/eth1indexer"
	"github.com/gobitfly/beaconchain/cmd/ethstore_exporter"
	"github.com/gobitfly/beaconchain/cmd/exporter"
	"github.com/gobitfly/beaconchain/cmd/misc"
	"github.com/gobitfly/beaconchain/cmd/node_jobs_processor"
	"github.com/gobitfly/beaconchain/cmd/notification_collector"
	"github.com/gobitfly/beaconchain/cmd/notification_sender"
	"github.com/gobitfly/beaconchain/cmd/rewards_exporter"
	"github.com/gobitfly/beaconchain/cmd/signatures"
	"github.com/gobitfly/beaconchain/cmd/statistics"
	"github.com/gobitfly/beaconchain/cmd/typescript_converter"
	"github.com/gobitfly/beaconchain/cmd/user_service"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal(nil, "missing target", 0)
	}
	target := os.Args[1]

	log.Info(target)
	switch target {
	case "api":
		api.Run()
	case "blobindexer":
		blobindexer.Run()
	case "eth1indexer":
		eth1indexer.Run()
	case "ethstore-exporter":
		ethstore_exporter.Run()
	case "exporter":
		exporter.Run()
	case "misc":
		misc.Run()
	case "node-jobs-processor":
		node_jobs_processor.Run()
	case "notification-collector":
		notification_collector.Run()
	case "notification-sender":
		notification_sender.Run()
	case "rewards-exporter":
		rewards_exporter.Run()
	case "signatures":
		signatures.Run()
	case "statistics":
		statistics.Run()
	case "typescript-converter":
		typescript_converter.Run()
	case "user-service":
		user_service.Run()
	default:
		log.Fatal(nil, "unknown target", 0)
	}
}
