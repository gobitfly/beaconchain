package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/spec"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	commonTypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/swaggo/swag/gen"
)

const (
	searchDir  = "./pkg/api"
	mainAPI    = "handlers/public.go"
	outputDir  = "./docs"
	outputType = "json" // can also be yaml

	apiPrefix = "/api/v2"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		log.Infof(version.Version)
		log.Infof(version.GoVersion)
		return
	}
	// generate swagger doc
	config := &gen.Config{
		SearchDir:   searchDir,
		MainAPIFile: mainAPI,
		OutputDir:   outputDir,
		OutputTypes: []string{outputType},
	}
	err := gen.New().Build(config)
	if err != nil {
		log.Fatal(err, "error generating swagger docs", 0)
	}

	log.Info("\n-------------\nswagger docs generated successfully, now loading endpoint weights from db\n-------------")

	if *configPath == "" {
		log.Warn("no config file provided, weights will not be inserted into swagger docs", 0)
		os.Exit(0)
	}

	// load endpoint weights from db
	cfg := &commonTypes.Config{}
	err = utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}

	da := dataaccess.NewDataAccessService(cfg)
	defer da.Close()
	apiWeights, err := da.GetApiWeights(context.Background())
	if err != nil {
		log.Fatal(err, "error loading endpoint weights from db", 0)
	}

	// insert endpoint weights into swagger doc
	data, err := os.ReadFile(outputDir + "/swagger." + outputType)
	if err != nil {
		log.Fatal(err, "error reading swagger docs", 0)
	}
	newData, err := insertApiWeights(data, apiWeights)
	if err != nil {
		log.Fatal(err, "error inserting api weights", 0)
	}

	// write updated swagger doc
	err = os.WriteFile(outputDir+"/swagger."+outputType, newData, os.ModePerm)
	if err != nil {
		log.Fatal(err, "error writing new swagger docs", 0)
	}

	log.Info("\n-------------\napi weights inserted successfully\n-------------")
}

func insertApiWeights(data []byte, apiWeightItems []types.ApiWeightItem) ([]byte, error) {
	// unmarshal swagger file
	var swagger *spec.Swagger
	if err := json.Unmarshal(data, &swagger); err != nil {
		return nil, fmt.Errorf("error unmarshalling swagger docs: %w", err)
	}

	// iterate endpoints from swagger file
	for pathString, pathItem := range swagger.Paths.Paths {
		for methodString, operation := range getPathItemOperationMap(pathItem) {
			if operation == nil {
				continue
			}
			// get weight and bucket for each endpoint
			weight := 1
			bucket := ""
			for _, apiWeightItem := range apiWeightItems {
				// ignore endpoints that don't belong to v2
				if !strings.HasPrefix(apiWeightItem.Endpoint, apiPrefix) {
					continue
				}
				// compare endpoint and method
				if pathString == strings.TrimPrefix(apiWeightItem.Endpoint, apiPrefix) && methodString == apiWeightItem.Method {
					weight = apiWeightItem.Weight
					bucket = apiWeightItem.Bucket + " "
					break
				}
			}
			// insert weight and bucket into endpoint summary
			plural := ""
			if weight > 1 {
				plural = "s"
			}
			operation.Summary = fmt.Sprintf("(%d %scredit%s) %s", weight, bucket, plural, operation.Summary)
		}
	}

	return json.MarshalIndent(swagger, "", "  ")
}

func getPathItemOperationMap(item spec.PathItem) map[string]*spec.Operation {
	return map[string]*spec.Operation{
		"get":     item.Get,
		"put":     item.Put,
		"post":    item.Post,
		"delete":  item.Delete,
		"options": item.Options,
		"head":    item.Head,
		"patch":   item.Patch,
	}
}
