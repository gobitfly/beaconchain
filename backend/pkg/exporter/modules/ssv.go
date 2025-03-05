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
	Type   string            `json:"type"`
	Filter SSVExporterFilter `json:"filter"`
	Data   []SSVExporterData `json:"data"`
}

type SSVExporterFilter struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type SSVExporterData struct {
	Index     int                    `json:"index"`
	Publickey string                 `json:"publicKey"`
	Operators []SSVExporterOperators `json:"operators"`
}

type SSVExporterOperators struct {
	Nodeid    int    `json:"nodeId"`
	Publickey string `json:"publicKey"`
}

type ssvExporter struct {
	db     db.ConsensusDBI
	ctx    context.Context
	dialer Dialer
}

func newSSVExporter(db db.ConsensusDBI) ssvExporter {
	return ssvExporter{
		db:     db,
		ctx:    context.Background(),
		dialer: &realDialer{},
	}
}

func (ssv *ssvExporter) Export() {
	for {
		select {
		case <-ssv.ctx.Done():
			log.Info("ssv export process cancelled")
			return
		default:
			err := ssv.exportSSV()
			if err != nil {
				log.Error(err, "error exporting ssv validators", 0)
			}
			log.Warnf("connection to ssv-exporter closed, reconnecting")

			// ensure it exits early if context is cancelled
			select {
			case <-time.After(time.Second * 10):
				// wait for 10s before reconnecting
			case <-ssv.ctx.Done():
				return
			}
		}
	}
}

func (ssv *ssvExporter) exportSSV() error {
	conn, resp, err := ssv.connectToWebSocket()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer resp.Body.Close()
	done := make(chan struct{})
	go ssv.handleWebSocketMessages(conn, done)

	qryValidatorsTicker := time.NewTicker(time.Minute * 10)
	defer qryValidatorsTicker.Stop()

	for {
		select {
		case <-ssv.ctx.Done():
			log.Info("export loop cancelled", 0)
			return nil
		case <-done:
			return nil
		case <-qryValidatorsTicker.C:
			if err := requestValidators(conn); err != nil {
				return err
			}
		}
	}
}

func (ssv *ssvExporter) connectToWebSocket() (WebSocketConn, *http.Response, error) {
	conn, resp, err := ssv.dialer.Dial(utils.Config.SSVExporter.Address, nil)
	if err != nil {
		return nil, nil, err
	}

	return conn, resp, nil
}

func (ssv *ssvExporter) handleWebSocketMessages(conn WebSocketConn, done chan struct{}) {
	defer close(done)
	for {
		select {
		case <-ssv.ctx.Done():
			log.Info("message handling cancelled", 0)
			return
		default:
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
}

func requestValidators(conn WebSocketConn) error {
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
	var batchSize = 5000
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

func prepareBatchInsert(data []SSVExporterData) ([]string, [][]byte) {
	index := 1
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([][]byte, 0, len(data)*index)

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

type Dialer interface {
	Dial(url string, requestHeader http.Header) (WebSocketConn, *http.Response, error)
}

type WebSocketConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

type realDialer struct{}

func (d *realDialer) Dial(url string, requestHeader http.Header) (WebSocketConn, *http.Response, error) {
	conn, resp, err := websocket.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		return nil, nil, err
	}
	return conn, resp, nil
}
