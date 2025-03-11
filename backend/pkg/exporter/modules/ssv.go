package modules

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"

	"github.com/gorilla/websocket"
)

type ssvExporter struct {
	db db2.ConsensusRepository
}

func newSSVExporter(db db2.ConsensusRepository) *ssvExporter {
	return &ssvExporter{
		db: db,
	}
}

func (ssv *ssvExporter) Export(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("ssv export process cancelled")
			return
		default:
			err := ssv.exportSSV(ctx)
			if err != nil {
				log.Error(err, "error exporting ssv validators", 0)
			}
			log.Warnf("connection to ssv-exporter closed, reconnecting")
			time.Sleep(time.Second * 10)
		}
	}
}

func (ssv *ssvExporter) exportSSV(ctx context.Context) error {
	conn, resp, err := websocket.DefaultDialer.Dial(utils.Config.SSVExporter.Address, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer resp.Body.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Error(err, "error reading message from ssv-exporter", 0)
				return
			}

			timeStart := time.Now()
			var res types.SSVExporterResponse
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Error(err, "error unmarshaling json from ssv-exporter", 0)
				continue
			}

			log.InfoWithFields(log.Fields{"number": len(res.Data)}, "exporting ssv validators")
			err = ssv.saveSSV(&res)
			if err != nil {
				log.Error(err, "error tagging ssv validators", 0)
				continue
			}
			log.InfoWithFields(log.Fields{"number": len(res.Data), "duration": time.Since(timeStart)}, "tagged ssv validators")
		}
	}()

	// query validators periodically
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("export loop cancelled")
			return nil
		case <-done:
			log.Info("websocket connection closed")
			return nil
		case <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"validator","filter":{"from":0}}`))
			if err != nil {
				return err
			}
		}
	}
}

func (ssv *ssvExporter) saveSSV(res *types.SSVExporterResponse) error {
	// make sure to correct wrongly marked validators
	err := ssv.db.DeleteInvalidTags()
	if err != nil {
		return err
	}

	batchSize := 5000
	for start := 0; start < len(res.Data); start += batchSize {
		end := start + batchSize
		if end > len(res.Data) {
			end = len(res.Data)
		}

		err := ssv.db.SaveValidatorTags(res.Data[start:end])
		if err != nil {
			return err
		}
	}

	// currently the ssv-exporter also exports publickeys that are not actually part of the network
	err = ssv.db.DeleteValidatorTags()
	if err != nil {
		return err
	}

	return nil
}
