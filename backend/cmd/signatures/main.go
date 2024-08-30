package signatures

import (
	"encoding/json"
	"os"

	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	_ "github.com/jackc/pgx/v5/stdlib"

	//nolint:gosec
	_ "net/http/pprof"
)

/**
* This function is for indexing smart contract function names from www.4byte.directory
* so that we can label the transction function calls instead of the "id"
**/
func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to the config file, if empty string defaults will be used")

	versionFlag := fs.Bool("version", false, "Show version and exit")
	_ = fs.Parse(os.Args[2:])

	if *versionFlag {
		log.Info(version.Version)
		log.Info(version.GoVersion)
		return
	}

	log.Infof("version: %v, config file path: %v", version.Version, configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg
	log.InfoWithFields(log.Fields{"config": *configPath, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	if utils.Config.Metrics.Enabled {
		go func() {
			log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
			if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}()
	}

	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		SSL:          cfg.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		SSL:          cfg.ReaderDatabase.SSL,
	}, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, "1", utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Error(err, "error initializing bigtable", 0)
		return
	}

	go ImportSignatures(bt, types.MethodSignature)
	time.Sleep(time.Second * 2) // we need a little delay, as the api does not like two requests at the same time
	go ImportSignatures(bt, types.EventSignature)

	utils.WaitForCtrlC()
}

func ImportSignatures(bt *db.Bigtable, st types.SignatureType) {
	// Per default we start with the latest signatures (first page = latest signatures)
	firstPage := "https://www.4byte.directory/api/v1/signatures/"
	if st == types.EventSignature {
		firstPage = "https://www.4byte.directory/api/v1/event-signatures/"
	}
	page := firstPage
	status, err := bt.GetSignatureImportStatus(st)
	if err != nil {
		log.Error(err, "error getting signature import status from bigtable", 0)
		return
	}
	isFirst := true

	// If we never completed syncing all signatures we continue with the next page
	if !status.HasFinished && status.NextPage != nil {
		page = *status.NextPage
		isFirst = false
	}

	var latestTimestamp time.Time
	// Timestamp of the first item from the last run
	if status.LatestTimestamp != nil {
		latestTimestamp, _ = time.Parse(time.RFC3339, *status.LatestTimestamp)
	}
	//nolint:ineffassign
	sleepTime := 4 * time.Second

	for ; ; time.Sleep(sleepTime) { // timout needed due to rate limit
		sleepTime = 4 * time.Second
		log.Infof("Get signatures for: %v", page)
		start := time.Now()
		next, sigs, err := GetNextSignatures(page)

		if err != nil {
			metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_get_signatures_failed", st)).Inc()
			log.Error(err, "error getting signatures", 0)
			sleepTime = time.Minute
			continue
		}

		// If had a complete sync done in the past, we only need to get signatures newer then the onces from our prev. run
		if status.LatestTimestamp != nil && status.HasFinished {
			createdAt, _ := time.Parse(time.RFC3339, *status.LatestTimestamp)
			if createdAt.UnixMilli() <= latestTimestamp.UnixMilli() {
				isFirst = true
				if page != firstPage {
					log.Infof("Our %v signature data of page %v is up to date so we jump to the first page", st, page)
					page = firstPage
				} else {
					log.Infof("Our %v signature data is up to date so we wait for an hour to check again", st)
					sleepTime = time.Hour
				}
				continue
			}
		}

		err = db.BigtableClient.SaveSignatures(sigs, st)
		if err != nil {
			metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_save_to_bt_failed", st)).Inc()
			log.Error(err, "error saving signatures into bigtable", 0)
			sleepTime = time.Minute
			continue
		}

		// Lets save the timestamp from the first (=latest) entry
		if isFirst {
			status.LatestTimestamp = &sigs[0].CreatedAt
			isFirst = false
		}

		if next == nil {
			status.NextPage = nil
			status.HasFinished = true
		} else {
			if !status.HasFinished {
				status.NextPage = next
			}
			page = *next
		}
		if status != nil && (status.HasFinished || status.NextPage != nil) {
			nextPage := "-"
			latestTimestamp := "-"
			if status.NextPage != nil {
				nextPage = *status.NextPage
			}
			if status.LatestTimestamp != nil {
				latestTimestamp = *status.LatestTimestamp
			}
			log.Infof("Save %v Sig ts: %v next: %v", st, latestTimestamp, nextPage)
			err = bt.SaveSignatureImportStatus(*status, st)
			if err != nil {
				metrics.Errors.WithLabelValues(fmt.Sprintf("%v_signatures_save_status_to_bt_failed", st)).Inc()
				log.Error(err, "error saving signature status into bigtable", 0)
				sleepTime = time.Minute
			}
		}
		metrics.TaskDuration.WithLabelValues(fmt.Sprintf("%v_signatures_page_imported", st)).Observe(time.Since(start).Seconds())
		services.ReportStatus(fmt.Sprintf("%v_signatures", st), "Running", nil)
	}
}

func GetNextSignatures(page string) (*string, []types.Signature, error) {
	httpClient := &http.Client{Timeout: time.Second * 10}

	resp, err := httpClient.Get(page)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("error querying signatures api: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	type signatureResponse struct {
		Results []types.Signature `json:"results"`
		Next    *string           `json:"next"`
	}

	respParsed := &signatureResponse{}
	err = json.Unmarshal(body, respParsed)
	if err != nil {
		return nil, nil, err
	}

	return respParsed.Next, respParsed.Results, nil
}
