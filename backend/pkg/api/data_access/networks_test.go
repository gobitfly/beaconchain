package dataaccess

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/stretchr/testify/assert"
)

func TestGetAllNetworks(t *testing.T) {
	d := &DataAccessService{}
	t.Run("GetAllNetworks", func(t *testing.T) {
		networks, err := d.GetAllNetworks()
		assert.Nil(t, err)
		assert.Contains(t, networks, types.NetworkInfo{
			ChainId:           1,
			Name:              "ethereum",
			NotificationsName: "mainnet",
		})
		assert.Contains(t, networks, types.NetworkInfo{
			ChainId:           100,
			Name:              "gnosis",
			NotificationsName: "gnosis",
		})
		assert.Contains(t, networks, types.NetworkInfo{
			ChainId:           17000,
			Name:              "holesky",
			NotificationsName: "holesky",
		})
		assert.Contains(t, networks, types.NetworkInfo{
			ChainId:           560048,
			Name:              "hoodi",
			NotificationsName: "hoodi",
		})
		assert.Contains(t, networks, types.NetworkInfo{
			ChainId:           11155111,
			Name:              "sepolia",
			NotificationsName: "sepolia",
		})
	})
}
