package notification

import (
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

// Used for isolated testing
func GetNotificationsForEpoch(pubkeyCachePath string, epoch uint64) (types.NotificationsPerUserId, error) {
	err := initPubkeyCache(pubkeyCachePath)
	if err != nil {
		log.Fatal(err, "error initializing pubkey cache path for notifications", 0)
	}
	return collectNotifications(epoch)
}
