package config

import (
	"github.com/gobitfly/beaconchain-api/internal/log"
	"github.com/spf13/viper"
)

const DEFAULT_DEV_IDENTIFIER = "default"

type Config struct {
	environment Environment // The environment (Dev,Staging,Prod) the service is hosted in
	httpPort    string
	grpcPort    string
}

type Bigtable struct {
	Project             string `yaml:"project"`
	Instance            string `yaml:"instance"`
	Emulator            bool   `yaml:"emulator"`
	EmulatorPort        int    `yaml:"emulatorPort"`
	EmulatorHost        string `yaml:"emulatorHost"`
	V2SchemaCutOffEpoch uint64 `yaml:"v2SchemaCutOffEpoch"`
	Remote              string `yaml:"remote"`
}

type DatabaseConfig struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DbName       string `yaml:"dbName"`
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	SSL          bool   `yaml:"ssl"`
	Failovers    []struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"failovers"`
}

type ServiceConfig struct {
	HttpPort         string         `yaml:"httpPort"`
	GrpcPort         string         `yaml:"grpcPort"`
	ReaderDatabase   DatabaseConfig `yaml:"readerDatabase"`
	WriterDatabase   DatabaseConfig `yaml:"writerDatabase"`
	ReaderClickhouse DatabaseConfig `yaml:"readerClickhouse"`
	WriterClickhouse DatabaseConfig `yaml:"writerClickhouse"`
	Bigtable         Bigtable       `yaml:"bigtable"`
	RawBigtable      Bigtable       `yaml:"rawBigtable"`
}

// Two kinds of configs:
// 1. Chain Config, used to define the "Specification" of each chain
// 3. Service Config, which is used to define the parameters that the service itself runs with.

func LoadChainConfig(chain ChainName) {
	// "configs/chain/default.yaml"
	viper.AddConfigPath("configs/chain")
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	// Read the default config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file, %s", err)
	}

	// Now load in the override config file. It replaces anything which exists in both
	viper.SetConfigName(string(chain))
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatal("Error reading config file, %s", err)
	}

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	logDebugConfigKeys()
}

func logDebugConfigKeys() {
	log.Debugf("The following config variables (including Env variables) were loaded: %s",
		viper.AllKeys())
}
