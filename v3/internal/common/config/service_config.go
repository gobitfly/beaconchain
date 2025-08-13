package config

import (
	"context"
	"flag"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"cloud.google.com/go/storage"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
}

type RedisConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type BigtableConfig struct {
	Project             string `yaml:"project"`
	Instance            string `yaml:"instance"`
	Emulator            bool   `yaml:"emulator"`
	EmulatorPort        int    `yaml:"emulatorPort"`
	EmulatorHost        string `yaml:"emulatorHost"`
	V2SchemaCutOffEpoch uint64 `yaml:"v2SchemaCutOffEpoch"`
	Remote              string `yaml:"remote"`
}

type DatabaseConfig struct {
	IsCloudConnection bool   `yaml:"isCloudConnection"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	DbName            string `yaml:"dbName"`
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	MaxOpenConns      int    `yaml:"maxOpenConns"`
	MaxIdleConns      int    `yaml:"maxIdleConns"`
	SSL               bool   `yaml:"ssl"`
	Failovers         []struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"failovers"`
}

type ServiceConfig struct {
	Type               string
	HttpPort           string `yaml:"httpPort"`
	GrpcPort           string `yaml:"grpcPort"`
	ExposeSchema       bool   `yaml:"exposeSchema"`
	InternalServiceUri string `yaml:"internalServiceUri"` // output only
	ExternalServiceUri string `yaml:"externalServiceUri"` // output only

	IsCloudDeployment bool `yaml:"isCloudDeployment"` // temp flag, remove

	ReaderChainDatabaseMainnet DatabaseConfig `yaml:"readerChainDatabaseMainnet"`
	WriterChainDatabaseMainnet DatabaseConfig `yaml:"writerChainDatabaseMainnet"`
	ReaderChainDatabaseGnosis  DatabaseConfig `yaml:"readerChainDatabaseGnosis"`
	WriterChainDatabaseGnosis  DatabaseConfig `yaml:"writerChainDatabaseGnosis"`
	ReaderChainDatabaseHoodi   DatabaseConfig `yaml:"readerChainDatabaseHoodi"`
	WriterChainDatabaseHoodi   DatabaseConfig `yaml:"writerChainDatabaseHoodi"`

	ReaderAdminDatabase DatabaseConfig `yaml:"readerAdminDatabase"`
	WriterAdminDatabase DatabaseConfig `yaml:"writerAdminDatabase"`
	ReaderClickhouse    DatabaseConfig `yaml:"readerClickhouse"`
	WriterClickhouse    DatabaseConfig `yaml:"writerClickhouse"`
	Bigtable            BigtableConfig `yaml:"bigtable"`
	RawBigtable         BigtableConfig `yaml:"rawBigtable"`
	Redis               RedisConfig    `yaml:"redis"`
}

// Two kinds of configs:
// 1. Chain Config, used to define the "Specification" of each chain
// 3. Service Config, which is used to define the parameters that the service itself runs with.

func LoadServiceConfig() *ServiceConfig {

	// using standard library "flag" package
	env := flag.String("environment", "Development", "Name of the environment")
	apiType := flag.String("type", "external", "api to launch (internal or external)")
	gcsBucket := flag.String("gcs-bucket", "", "GCS bucket name")
	gcsObject := flag.String("gcs-object", "beaconchain-config.yaml", "GCS object name")
	flag.Parse()

	log.Infof("Found flag environment: %s", *env)

	log.Info("Got this far: " + *env)
	// "configs/service/default.yaml"
	viper.AddConfigPath("configs/service")    // Typical "Run from cmd-line path"
	viper.AddConfigPath("../configs/service") // Typical "Run debug from vs-code path"
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	// Read the default config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Now load in the override config file. It replaces anything which exists in both
	viper.SetConfigName(*env)
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatalf("Error reading %s config: %v", *env, err)
	}

	// load config from gcs
	if gcsBucket != nil && *gcsBucket != "" {
		log.Infof("Reading config from gcs bucket: %s", *gcsBucket)
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to create GCS client: %v", err)
		}
		defer func() {
			if cerr := client.Close(); cerr != nil {
				log.Fatalf("failed to close GCS client: %v", cerr)
			}
		}()

		rc, err := client.Bucket(*gcsBucket).Object(*gcsObject).NewReader(ctx)
		if err != nil {
			log.Error(fmt.Errorf("failed to read object: %v", err))
		}
		defer func() {
			if cerr := rc.Close(); cerr != nil {
				log.Fatalf("failed to close GCS object reader: %v", cerr)
			}
		}()
		// override
		if err := viper.ReadConfig(rc); err != nil {
			log.Error(fmt.Errorf("failed to read config: %v", err))
		}
	}

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	// resolve secrets
	if err := replaceNestedSecrets(context.Background(), "", viper.AllSettings()); err != nil {
		log.Error("failed to resolve secrets", err)
	}

	serviceConfig := &ServiceConfig{}
	err = viper.Unmarshal(serviceConfig)
	if err != nil {
		log.Error("unable to decode into config struct", err)
	}
	serviceConfig.Type = *apiType

	logDebugConfigKeys()

	return serviceConfig
}

func fetchSecret(ctx context.Context, secretName string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Fatalf("failed to close secretmanager client: %v", cerr)
		}
	}()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}
	return string(result.Payload.Data), nil
}

func replaceNestedSecrets(ctx context.Context, parentKey string, data map[string]interface{}) error {
	for k, v := range data {
		fullKey := k
		if parentKey != "" {
			fullKey = fmt.Sprintf("%s.%s", parentKey, k)
		}
		switch val := v.(type) {
		case string:
			if strings.HasPrefix(val, "gsm:") {
				secretVal, err := fetchSecret(ctx, strings.TrimPrefix(val, "gsm:"))
				if err != nil {
					return err
				}
				viper.Set(fullKey, secretVal)
			}
		case map[string]interface{}:
			if err := replaceNestedSecrets(ctx, fullKey, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func logDebugConfigKeys() {
	log.Infof("The following config variables (including Env variables) were loaded: %s",
		viper.AllKeys())
}
