package modules

import (
	"fmt"
	"sync"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/pkg/errors"
)

type epochWriter struct {
	*dashboardData
	mutex *sync.Mutex
}

func newEpochWriter(d *dashboardData) *epochWriter {
	return &epochWriter{
		dashboardData: d,
		mutex:         &sync.Mutex{},
	}
}

func (d *epochWriter) WriteEpochsData(epochs []uint64, data *db.VDBDataEpochColumns) error {
	uuid := "epochs"
	for _, epoch := range epochs {
		uuid += fmt.Sprintf("_%d", epoch)
	}
	d.log.Infof("Writing epoch data for %d epochs", epochs)
	err := db.UltraFastDumpToClickhouse(data, edb.EpochWriterSink, uuid)
	if err != nil {
		return errors.Wrap(err, "failed to dump to clickhouse")
	}
	return nil
}
