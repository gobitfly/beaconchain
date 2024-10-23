package dataaccess

import (
	"context"
	"strings"

	apiTypes "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type MachineRepository interface {
	GetUserMachineMetrics(context context.Context, userID uint64, limit int, offset int) (*apiTypes.MachineMetricsData, error)
	PostUserMachineMetrics(context context.Context, userID uint64, machine, process string, data []byte) error
}

func (d *DataAccessService) GetUserMachineMetrics(ctx context.Context, userID uint64, limit int, offset int) (*apiTypes.MachineMetricsData, error) {
	data := &apiTypes.MachineMetricsData{}

	g := errgroup.Group{}

	g.Go(func() error {
		var err error
		data.SystemMetrics, err = d.bigtable.GetMachineMetricsSystem(types.UserId(userID), limit, offset)
		return err
	})

	g.Go(func() error {
		var err error
		data.ValidatorMetrics, err = d.bigtable.GetMachineMetricsValidator(types.UserId(userID), limit, offset)
		return err
	})

	g.Go(func() error {
		var err error
		data.NodeMetrics, err = d.bigtable.GetMachineMetricsNode(types.UserId(userID), limit, offset)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, errors.Wrap(err, "could not get stats")
	}

	return data, nil
}

func (d *DataAccessService) PostUserMachineMetrics(ctx context.Context, userID uint64, machine, process string, data []byte) error {
	err := db.BigtableClient.SaveMachineMetric(process, types.UserId(userID), machine, data)
	if err != nil {
		if strings.HasPrefix(err.Error(), "rate limit") {
			return err
		}
		return errors.Wrap(err, "could not save stats")
	}
	return nil
}
