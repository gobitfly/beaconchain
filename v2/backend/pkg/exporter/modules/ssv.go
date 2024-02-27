package modules

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	Data []struct {
		Index     int    `json:"index"`
		Publickey string `json:"publicKey"`
		Operators []struct {
			Nodeid    int    `json:"nodeId"`
			Publickey string `json:"publicKey"`
		} `json:"operators"`
	} `json:"data"`
}

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
	c, r, err := websocket.DefaultDialer.Dial(utils.Config.SSVExporter.Address, nil)
	if err != nil {
		return err
	}
	defer c.Close()
	defer r.Body.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Error(err, "error reading message from ssv-exporter", 0)
				return
			}

			t0 := time.Now()
			res := SSVExporterResponse{}
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Error(err, "error unmarshaling json from ssv-exporter", 0)
				continue
			}
			log.InfoWithFields(log.Fields{"number": len(res.Data)}, "exporting ssv validators")
			err = saveSSV(&res)
			if err != nil {
				log.Error(err, "error tagging ssv validators", 0)
				continue
			}
			log.InfoWithFields(log.Fields{"number": len(res.Data), "duration": time.Since(t0)}, "tagged ssv validators")
		}
	}()

	qryValidatorsTicker := time.NewTicker(time.Minute * 10)
	defer qryValidatorsTicker.Stop()

	for {
		err := c.WriteMessage(websocket.TextMessage, []byte(`{"type":"validator","filter":{"from":0}}`))
		if err != nil {
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

func saveSSV(res *SSVExporterResponse) error {
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Error(err, "error rolling back transaction", 0)
		}
	}()

	// for now make sure to correct wrongly marked validators
	for {
		res, err := tx.Exec(`delete from validator_tags where publickey in (select publickey from validator_tags where tag = 'ssv' limit 1000)`)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	batchSize := 5000
	for b := 0; b < len(res.Data); b += batchSize {
		start := b
		end := b + batchSize
		if len(res.Data) < end {
			end = len(res.Data)
		}
		n := 1
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*n)
		for i, d := range res.Data[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'ssv')", i*n+1))
			pubkey, err := hex.DecodeString(strings.Replace(d.Publickey, "0x", "", -1))
			if err != nil {
				return err
			}
			valueArgs = append(valueArgs, pubkey)
		}
		_, err := tx.Exec(fmt.Sprintf(`insert into validator_tags (publickey, tag) values %s on conflict (publickey, tag) do nothing`, strings.Join(valueStrings, ",")), valueArgs...)
		if err != nil {
			return err
		}
	}

	// currently the ssv-exporter also exports publickeys that are not actually part of the network
	for {
		res, err := tx.Exec(`delete from validator_tags where publickey in (select publickey from validator_tags where publickey not in (select pubkey from validators) limit 1000)`)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
