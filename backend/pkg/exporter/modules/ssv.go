package modules

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"

	"github.com/gorilla/websocket"
)

type SSVExporterResponse struct {
	Type   string `json:"type"`
	Filter struct {
		From int `json:"from"`
		To   int `json:"to"`
	} `json:"filter"`
	Data []SSVExporterData `json:"data"`
}

type SSVExporterData struct {
	Index     int    `json:"index"`
	Publickey string `json:"publicKey"`
	Operators []struct {
		Nodeid    int    `json:"nodeId"`
		Publickey string `json:"publicKey"`
	} `json:"operators"`
}

var batchSize = 5000

type ssvExporter struct {
	db  db.ConsensusDBI
	ctx context.Context
}

func newSSVExporter(db db.ConsensusDBI) ssvExporter {
	return ssvExporter{
		db:  db,
		ctx: context.Background(),
	}
}

func (ssv *ssvExporter) Export() {
	for {
		err := ssv.exportSSV()
		if err != nil {
			log.Error(err, "error exporting ssv validators", 0)
		}
		log.Warnf("connection to ssv-exporter closed, reconnecting")
		time.Sleep(time.Second * 10)
	}
}

func (ssv *ssvExporter) exportSSV() error {
	conn, r, err := connectToWebSocket()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer r.Body.Close()

	done := make(chan struct{})
	go ssv.handleWebSocketMessages(conn, done)

	qryValidatorsTicker := time.NewTicker(time.Minute * 10)
	defer qryValidatorsTicker.Stop()

	for {
		if err := requestValidators(conn); err != nil {
			return err
		}

		select {
		case <-qryValidatorsTicker.C:
			continue
		case <-done:
			return nil
		}
	}
}

func connectToWebSocket() (*websocket.Conn, *http.Response, error) {
	conn, r, err := websocket.DefaultDialer.Dial(utils.Config.SSVExporter.Address, nil)
	if err != nil {
		return nil, nil, err
	}

	return conn, r, nil
}

func (ssv *ssvExporter) handleWebSocketMessages(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error(err, "error reading message from ssv-exporter", 0)
			return
		}

		timeStart := time.Now()
		res := SSVExporterResponse{}
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
}

func requestValidators(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"validator","filter":{"from":0}}`))
}

func (ssv *ssvExporter) saveSSV(res *SSVExporterResponse) error {
	// make sure to correct wrongly marked validators
	if err := ssv.db.DeleteInvalidTags(); err != nil {
		return err
	}

	if err := ssv.insertSSVTags(res); err != nil {
		return err
	}

	// currently the ssv-exporter also exports publickeys that are not actually part of the network
	if err := ssv.db.DeleteValidatorTags(); err != nil {
		return err
	}

	return nil
}

func (ssv *ssvExporter) insertSSVTags(response *SSVExporterResponse) error {
	for b := 0; b < len(response.Data); b += batchSize {
		start := b
		end := b + batchSize
		if len(response.Data) < end {
			end = len(response.Data)
		}
		valueStrings, valueArgs := prepareBatchInsert(response.Data[start:end])

		err := ssv.db.SaveValidatorTags(valueStrings, valueArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareBatchInsert(data []SSVExporterData) ([]string, []interface{}) {
	index := 1
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*index)

	for i, d := range data {
		pubkey, err := hex.DecodeString(strings.Replace(d.Publickey, "0x", "", -1))
		if err != nil {
			log.Error(err, "error decoding public key", 0)
			continue
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'ssv')", i*index+1))
		valueArgs = append(valueArgs, pubkey)
	}

	return valueStrings, valueArgs
}
