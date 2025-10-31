package notification

import (
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/exporter/modules"
)

// Used for isolated testing
func GetNotificationsForEpoch(pubkeyCachePath string, epoch uint64) (types.NotificationsPerUserId, error) {
	mc, err := modules.GetModuleContext()
	if err != nil {
		log.Fatal(err, "error getting module context", 0)
	}

	err = initPubkeyCache(pubkeyCachePath)
	if err != nil {
		log.Fatal(err, "error initializing pubkey cache path for notifications", 0)
	}
	return collectNotifications(mc, epoch)
}

func GetHeadNotificationsForEpoch(pubkeyCachePath string, epoch uint64) (types.NotificationsPerUserId, error) {
	mc, err := modules.GetModuleContext()
	if err != nil {
		log.Fatal(err, "error getting module context", 0)
	}

	err = initPubkeyCache(pubkeyCachePath)
	if err != nil {
		log.Fatal(err, "error initializing pubkey cache path for notifications", 0)
	}
	return collectHeadNotifications(mc, epoch)
}

// Used for isolated testing
func GetUserNotificationsForEpoch(pubkeyCachePath string, epoch uint64) (types.NotificationsPerUserId, error) {
	err := initPubkeyCache(pubkeyCachePath)
	if err != nil {
		log.Fatal(err, "error initializing pubkey cache path for notifications", 0)
	}
	return collectUserDbNotifications(epoch)
}
