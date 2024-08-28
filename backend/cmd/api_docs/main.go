package api_docs

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/go-openapi/spec"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/ratelimit"
	commonTypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/swaggo/swag/gen"
)

const (
	searchDir  = "./pkg/api"
	mainAPI    = "handlers/public.go"
	outputDir  = "./docs"
	outputType = "json" // can also be "yaml"

	apiPrefix = "/api/v2/"
)

// Expects the following flags:
// --config: (optional) Path to the config file to add endpoint weights to the swagger docs

// Standard usage (execute in backend folder): go run cmd/main.go api-docs --config <path-to-config-file>

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	apiWeights, err := da.GetApiWeights(ctx)
	if err != nil {
		log.Fatal(err, "error loading endpoint weights from db", 0)
	}

	// insert endpoint weights into swagger doc
	data, err := os.ReadFile(outputDir + "/swagger." + outputType)
	if err != nil {
		log.Fatal(err, "error reading swagger docs", 0)
	}
	newData, err := getDataWithWeights(data, apiWeights)
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

func getDataWithWeights(data []byte, apiWeightItems []types.ApiWeightItem) ([]byte, error) {
	// unmarshal swagger file
	var swagger *spec.Swagger
	if err := json.Unmarshal(data, &swagger); err != nil {
		return nil, fmt.Errorf("error unmarshalling swagger docs: %w", err)
	}

	// iterate endpoints from swagger file
	for pathString, pathItem := range swagger.Paths.Paths {
		pathString = apiPrefix + pathString
		for methodString, operation := range getPathItemOperationMap(pathItem) {
			if operation == nil {
				continue
			}
			// get weight and bucket for each endpoint
			weight := ratelimit.DefaultWeight
			bucket := ratelimit.DefaultBucket
			index := slices.IndexFunc(apiWeightItems, func(item types.ApiWeightItem) bool {
				return pathString == item.Endpoint && methodString == item.Method
			})
			if index != -1 {
				weight = apiWeightItems[index].Weight
				bucket = apiWeightItems[index].Bucket
			}
			// insert weight and bucket into endpoint summary
			plural := ""
			if weight > 1 {
				plural = "s"
			}
			operation.Summary = fmt.Sprintf("(%d %s credit%s) %s", weight, bucket, plural, operation.Summary)
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
