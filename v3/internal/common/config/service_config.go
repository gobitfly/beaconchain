package config

import (
	"flag"

	"github.com/gobitfly/beaconchain-api/internal/log"
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
	Type         string
	HttpPort     string `yaml:"httpPort"`
	GrpcPort     string `yaml:"grpcPort"`
	ExposeSchema bool   `yaml:"exposeSchema"`

	IsCloudDeployment bool `yaml:"isCloudDeployment"` // temp flag, remove

	ReaderChainDatabase DatabaseConfig `yaml:"readerChainDatabase"`
	WriterChainDatabase DatabaseConfig `yaml:"writerChainDatabase"`
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

	// Optionally read from environment variables (e.g., override with ENV vars)
	viper.AutomaticEnv()

	serviceConfig := &ServiceConfig{}
	err = viper.Unmarshal(serviceConfig)
	if err != nil {
		log.Warnf("unable to decode into config struct, %v", err)
	}
	serviceConfig.Type = *apiType

	logDebugConfigKeys()

	return serviceConfig
}

func logDebugConfigKeys() {
	log.Infof("The following config variables (including Env variables) were loaded: %s",
		viper.AllKeys())
}
