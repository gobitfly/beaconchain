package modules

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/exporter/types"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"

	"github.com/gorilla/websocket"
)

var batchSize = 5000

func ssvExporter() {
	for {
		err := exportSSV()
		if err != nil {
			log.Error(err, "error exporting ssv validators", 0)
		}
		log.Warnf("connection to ssv-exporter closed, reconnecting")
		time.Sleep(time.Second * 10)
	}
}

func exportSSV() error {
	conn, r, err := connectToWebSocket()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer r.Body.Close()

	done := make(chan struct{})
	go handleWebSocketMessages(conn, done)

	qryValidatorsTicker := time.NewTicker(constants.Duration10Mins)
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

func handleWebSocketMessages(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		message, err := readWebSocketMessage(conn)
		if err != nil {
			log.Error(err, "error reading message from ssv-exporter", 0)
			return
		}

		res, err := unmarshalSSVResponse(message)
		if err != nil {
			log.Error(err, "error unmarshaling json from ssv-exporter", 0)
			continue
		}

		if err := processSSVResponse(res); err != nil {
			log.Error(err, "error processing ssv validators", 0)
			continue
		}
	}
}

func readWebSocketMessage(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	return message, err
}

func unmarshalSSVResponse(message []byte) (*types.SSVExporterResponse, error) {
	var res types.SSVExporterResponse
	err := json.Unmarshal(message, &res)
	return &res, err
}

func processSSVResponse(res *types.SSVExporterResponse) error {
	timeStart := time.Now()
	log.InfoWithFields(log.Fields{"number": len(res.Data)}, "exporting ssv validators")

	if err := saveSSV(res); err != nil {
		return err
	}

	log.InfoWithFields(log.Fields{"number": len(res.Data), "duration": time.Since(timeStart)}, "tagged ssv validators")

	return nil
}

func requestValidators(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"validator","filter":{"from":0}}`))
}

func saveSSV(res *types.SSVExporterResponse) error {
	// make sure to correct wrongly marked validators
	if err := db.DeleteInvalidTags(); err != nil {
		return err
	}

	if err := insertSSVTags(res); err != nil {
		return err
	}

	// currently the ssv-exporter also exports publickeys that are not actually part of the network
	if err := db.DeleteValidatorTags(); err != nil {
		return err
	}

	return nil
}

func insertSSVTags(response *types.SSVExporterResponse) error {
	for b := 0; b < len(response.Data); b += batchSize {
		start := b
		end := b + batchSize
		if len(response.Data) < end {
			end = len(response.Data)
		}
		valueStrings, valueArgs := prepareBatchInsert(response.Data[start:end])

		err := db.SaveValidatorTags(valueStrings, valueArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareBatchInsert(data []types.SSVExporterData) ([]string, []interface{}) {
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
