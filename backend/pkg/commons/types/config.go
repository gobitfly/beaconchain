package types

import (
	"html/template"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

type Bigtable struct {
	Project             string `yaml:"project" env:"PROJECT"`
	Instance            string `yaml:"instance" env:"INSTANCE"`
	Emulator            bool   `yaml:"emulator" env:"EMULATOR"`
	EmulatorPort        int    `yaml:"emulatorPort" env:"EMULATOR_PORT"`
	EmulatorHost        string `yaml:"emulatorHost" env:"EMULATOR_HOST"`
	V2SchemaCutOffEpoch uint64 `yaml:"v2SchemaCutOffEpoch" env:"V2_SCHEMA_CUTT_OFF_EPOCH"`
	Remote              string `yaml:"remote"`
}

// Config is a struct to hold the configuration data
type Config struct {
	JustV2         bool           `yaml:"justV2" env:"JUST_V2"` // temp, remove at some point
	DeploymentType string         `yaml:"deploymentType" env:"DEPLOYMENT_TYPE"`
	ReaderDatabase DatabaseConfig `yaml:"readerDatabase" env:", prefix=READER_"`
	WriterDatabase DatabaseConfig `yaml:"writerDatabase" env:", prefix=WRITER_"`
	AlloyReader    DatabaseConfig `yaml:"alloyReader" env:", prefix=ALLOY_READER_"`
	AlloyWriter    DatabaseConfig `yaml:"alloyWriter" env:", prefix=ALLOY_WRITER_"`
	Bigtable       Bigtable       `yaml:"bigtable" env:", prefix=BIGTABLE_"`
	RawBigtable    Bigtable       `yaml:"rawBigtable" env:", prefix=RAW_BIGTABLE_"`
	BlobIndexer    struct {
		S3 struct {
			Endpoint        string `yaml:"endpoint" env:"ENDPOINT"`                 // s3 endpoint
			Bucket          string `yaml:"bucket" env:"BUCKET"`                     // s3 bucket
			AccessKeyId     string `yaml:"accessKeyId" env:"ACCESS_KEY_ID"`         // s3 access key id
			AccessKeySecret string `yaml:"accessKeySecret" env:"ACCESS_KEY_SECRET"` // s3 access key secret
		} `yaml:"s3" env:", prefix=S3_"`
		PruneMarginEpochs    uint64 `yaml:"pruneMarginEpochs" env:"PRUNE_MARGIN_EPOCHS"`       // PruneMarginEpochs helps blobindexer to decide if connected node has pruned too far to have no holes in the data, set it to same value as lighthouse flag --blob-prune-margin-epochs
		DisableStatusReports bool   `yaml:"disableStatusReports" env:"DISABLE_STATUS_REPORTS"` // disable status reports (no connection to db needed)
	} `yaml:"blobIndexer" env:", prefix=BLOB_INDEXER_"`
	Chain                     `yaml:"chain"`
	Eth1ErigonEndpoint        string `yaml:"eth1ErigonEndpoint" env:"ETH1_ERIGON_ENDPOINT"`
	Eth1GethEndpoint          string `yaml:"eth1GethEndpoint" env:"ETH1_GETH_ENDPOINT"`
	EtherscanAPIKey           string `yaml:"etherscanApiKey" env:"ETHERSCAN_API_KEY"`
	EtherscanAPIBaseURL       string `yaml:"etherscanApiBaseUrl" env:"ETHERSCAN_API_BASEURL"`
	RedisCacheEndpoint        string `yaml:"redisCacheEndpoint" env:"REDIS_CACHE_ENDPOINT"`
	RedisSessionStoreEndpoint string `yaml:"redisSessionStoreEndpoint" env:"REDIS_SESSION_STORE_ENDPOINT"`
	TieredCacheProvider       string `yaml:"tieredCacheProvider" env:"CACHE_PROVIDER"`
	ReportServiceStatus       bool   `yaml:"reportServiceStatus" env:"REPORT_SERVICE_STATUS"`
	ClickHouse                struct {
		ReaderDatabase DatabaseConfig `yaml:"readerDatabase" env:", prefix=READER_"`
		WriterDatabase DatabaseConfig `yaml:"writerDatabase" env:", prefix=WRITER_"`
	} `yaml:"clickhouse" env:", prefix=CLICKHOUSE_"`
	Indexer struct {
		Enabled bool `yaml:"enabled" env:"INDEXER_ENABLED"`
		Node    struct {
			Port     string `yaml:"port" env:"PORT"`
			Host     string `yaml:"host" env:"HOST"`
			Type     string `yaml:"type" env:"TYPE"`
			PageSize int32  `yaml:"pageSize" env:"PAGE_SIZE"`
		} `yaml:"node" env:", prefix=INDEXER_NODE_"`
		ELDepositContractFirstBlock uint64 `yaml:"eth1DepositContractFirstBlock" env:"INDEXER_ETH1_DEPOSIT_CONTRACT_FIRST_BLOCK"`
		DoNotTraceDeposits          bool   `yaml:"doNotTraceDeposits" env:"INDEXER_DO_NOT_TRACE_DEPOSITS"`
		PubKeyTagsExporter          struct {
			Enabled bool `yaml:"enabled" env:"ENABLED"`
		} `yaml:"pubkeyTagsExporter" env:", prefix=PUBKEY_TAGS_EXPORTER_"`
		EnsTransformer struct {
			ValidRegistrarContracts []string `yaml:"validRegistrarContracts" env:"VALID_REGISTRAR_CONTRACTS"`
		} `yaml:"ensTransformer" env:", prefix=ENS"`
	} `yaml:"indexer"`
	Frontend struct {
		Debug                          bool   `yaml:"debug" env:"DEBUG"`
		BeaconchainETHPoolBridgeSecret string `yaml:"beaconchainETHPoolBridgeSecret" env:"BEACONCHAIN_ETHPOOL_BRIDGE_SECRET"`
		Kong                           string `yaml:"kong" env:"KONG"`
		OnlyAPI                        bool   `yaml:"onlyAPI" env:"ONLY_API"`
		CsrfAuthKey                    string `yaml:"csrfAuthKey" env:"CSRF_AUTHKEY"`
		CsrfInsecure                   bool   `yaml:"csrfInsecure" env:"CSRF_INSECURE"`
		DisableCharts                  bool   `yaml:"disableCharts" env:"disableCharts"`
		RecaptchaSiteKey               string `yaml:"recaptchaSiteKey" env:"RECAPTCHA_SITEKEY"`
		RecaptchaSecretKey             string `yaml:"recaptchaSecretKey" env:"RECAPTCHA_SECRETKEY"`
		Enabled                        bool   `yaml:"enabled" env:"ENABLED"`
		BlobProviderUrl                string `yaml:"blobProviderUrl" env:"BLOB_PROVIDER_URL"`
		SiteBrand                      string `yaml:"siteBrand" env:"SITE_BRAND"`
		Keywords                       string `yaml:"keywords" env:"KEYWORDS"`
		// Imprint is deprdecated place imprint file into the legal directory
		Imprint string `yaml:"imprint" env:"IMPRINT"`
		Legal   struct {
			TermsOfServiceUrl string `yaml:"termsOfServiceUrl" env:"TERMS_OF_SERVICE_URL"`
			PrivacyPolicyUrl  string `yaml:"privacyPolicyUrl" env:"PRIVACY_POLICY_URL"`
			ImprintTemplate   string `yaml:"imprintTemplate" env:"IMPRINT_TEMPLATE"`
		} `yaml:"legal" env:", prefix=LEGAL_"`
		SiteDomain   string `yaml:"siteDomain" env:"SITE_DOMAIN"`
		SiteName     string `yaml:"siteName" env:"SITE_NAME"`
		SiteTitle    string `yaml:"siteTitle" env:"SITE_TITLE"`
		SiteSubtitle string `yaml:"siteSubtitle" env:"SITE_SUBTITLE"`
		Server       struct {
			Port string `yaml:"port" env:"PORT"`
			Host string `yaml:"host" env:"HOST"`
		} `yaml:"server" env:", prefix=SERVER_"`
		ReaderDatabase DatabaseConfig `yaml:"readerDatabase" env:", prefix=READER_"`
		WriterDatabase DatabaseConfig `yaml:"writerDatabase" env:", prefix=WRITER_"`
		Stripe         struct {
			Webhook   string `yaml:"webhook" env:"WEBHOOK"`
			SecretKey string `yaml:"secretKey" env:"SECRET_KEY"`
			PublicKey string `yaml:"publicKey" env:"PUBLIC_KEY"`

			Sapphire string `yaml:"sapphire" env:"SAPPHIRE"`
			Emerald  string `yaml:"emerald" env:"EMERALD"`
			Diamond  string `yaml:"diamond" env:"DIAMOND"`
			Whale    string `yaml:"whale" env:"WHALE"`
			Goldfish string `yaml:"goldfish" env:"GOLDFISH"`
			Plankton string `yaml:"plankton" env:"PLANKTON"`

			Iron         string `yaml:"iron" env:"IRON"`
			IronYearly   string `yaml:"ironYearly" env:"IRON_YEARLY"`
			Silver       string `yaml:"silver" env:"SILVER"`
			SilverYearly string `yaml:"silverYearly" env:"SILVER_YEARLY"`
			Gold         string `yaml:"gold" env:"GOLD"`
			GoldYearly   string `yaml:"goldYearly" env:"GOLD_YEARLY"`

			Guppy         string `yaml:"guppy" env:"GUPPY"`
			GuppyYearly   string `yaml:"guppyYearly" env:"GUPPY_YEARLY"`
			Dolphin       string `yaml:"dolphin" env:"DOLPHIN"`
			DolphinYearly string `yaml:"dolphinYearly" env:"DOLPHIN_YEARLY"`
			Orca          string `yaml:"orca" env:"ORCA"`
			OrcaYearly    string `yaml:"orcaYearly" env:"ORCA_YEARLY"`

			VdbAddon1k        string `yaml:"vdbAddon1k" env:"VDB_ADDON_1K"`
			VdbAddon1kYearly  string `yaml:"vdbAddon1kYearly" env:"VDB_ADDON_1K_YEARLY"`
			VdbAddon10k       string `yaml:"vdbAddon10k" env:"VDB_ADDON_10K"`
			VdbAddon10kYearly string `yaml:"vdbAddon10kYearly" env:"VDB_ADDON_10K_YEARLY"`
		} `env:", prefix=STRIPE_"` // had no yaml tag, not touching it for now
		Ratelimits struct {
			FreeDay       int `yaml:"freeDay" env:"FREE_DAY"`
			FreeMonth     int `yaml:"freeMonth" env:"FREE_MONTH"`
			SapphierDay   int `yaml:"sapphireDay" env:"SAPPHIRE_DAY"`
			SapphierMonth int `yaml:"sapphireDay" env:"SAPPHIRE_MONTH"`
			EmeraldDay    int `yaml:"emeraldDay" env:"EMERALD_DAY"`
			EmeraldMonth  int `yaml:"emeraldMonth" env:"EMERALD_MONTH"`
			DiamondDay    int `yaml:"diamondDay" env:"DIAMOND_DAY"`
			DiamondMonth  int `yaml:"diamondMonth" env:"DIAMOND_MONTH"`
		} `yaml:"ratelimits" env:", prefix=RATELIMITS_"`
		RatelimitUpdateInterval time.Duration `yaml:"ratelimitUpdateInterval" env:"RATELIMIT_UPDATE_INTERVAL"`
		RatelimitEnabled        bool          `yaml:"ratelimitEnabled" env:"RATELIMIT_ENABLED"`
		RatelimitRedisTimeout   time.Duration `yaml:"ratelimitRedisTimeout" env:"RATELIMIT_REDIS_TIMEOUT"`
		SessionSecret           string        `yaml:"sessionSecret" env:"SESSION_SECRET"`
		SessionSameSiteNone     bool          `yaml:"sessionSameSiteNone" env:"SESSION_SAMESITE_NONE"`
		SessionCookieDomain     string        `yaml:"sessionCookieDomain" env:"SESSION_COOKIE_DOMAIN"`
		JwtSigningSecret        string        `yaml:"jwtSigningSecret" env:"JWT_SECRET"`
		JwtIssuer               string        `yaml:"jwtIssuer" env:"JWT_ISSUER"`
		JwtValidityInMinutes    int           `yaml:"jwtValidityInMinutes" env:"JWT_VALIDITY_INMINUTES"`
		MaxMailsPerEmailPerDay  int           `yaml:"maxMailsPerEmailPerDay" env:"MAX_MAIL_PER_EMAIL_PER_DAY"`
		Mail                    struct {
			SMTP struct {
				Server   string `yaml:"server" env:"SERVER"`
				Host     string `yaml:"host" env:"HOST"`
				User     string `yaml:"user" env:"USER"`
				Password string `yaml:"password" env:"PASSWORD"`
			} `yaml:"smtp" env:", prefix=SMTP_"`
			Mailgun struct {
				Domain     string `yaml:"domain" env:"DOMAIN"`
				PrivateKey string `yaml:"privateKey" env:"PRIVATE_KEY"`
				Sender     string `yaml:"sender" env:"SENDER"`
			} `yaml:"mailgun" env:", prefix=MAILGUN_"`
			Contact struct {
				SupportEmail string `yaml:"supportEmail" env:"SUPPORT_EMAIL"`
				InquiryEmail string `yaml:"inquiryEmail" env:"INQUIRY_EMAIL"`
			} `yaml:"contact" env:", prefix=CONTACT_"`
		} `yaml:"mail" env:", prefix=MAIL_"`
		GATag         string `yaml:"gatag" env:"GATAG"`
		VerifyAppSubs bool   `yaml:"verifyAppSubscriptions" env:"VERIFY_APP_SUBSCRIPTIONS"`
		Apple         struct {
			LegacyAppSubsAppleSecret string `yaml:"appSubsAppleSecret" env:"APP_SUBS_APPLE_SECRET"`
			KeyID                    string `yaml:"keyID" env:"APPLE_APP_KEY_ID"`
			IssueID                  string `yaml:"issueID" env:"APPLE_ISSUE_ID"`
			Certificate              string `yaml:"certificate" env:"APPLE_CERTIFICATE"`
		} `yaml:"apple"`
		AppSubsGoogleJSONPath string `yaml:"appSubsGoogleJsonPath" env:"APP_SUBS_GOOGLE_JSON_PATH"`
		DisableStatsInserts   bool   `yaml:"disableStatsInserts" env:"DISABLE_STATS_INSERTS"`
		ShowDonors            struct {
			Enabled bool   `yaml:"enabled" env:"SHOW_DONORS_ENABLED"`
			URL     string `yaml:"gitcoinURL" env:"GITCOIN_URL"`
		} `yaml:"showDonors"`
		Countdown struct {
			Enabled   bool          `yaml:"enabled" env:"ENABLED"`
			Title     template.HTML `yaml:"title" env:"TITLE"`
			Timestamp uint64        `yaml:"timestamp" env:"TIMESTAMP"`
			Info      string        `yaml:"info" env:"INFO"`
		} `yaml:"countdown" env:", prefix=COUNTDOWN_"`
		HttpReadTimeout    time.Duration `yaml:"httpReadTimeout" env:"HTTP_READ_TIMEOUT"`
		HttpWriteTimeout   time.Duration `yaml:"httpWriteTimeout" env:"HTTP_WRITE_TIMEOUT"`
		HttpIdleTimeout    time.Duration `yaml:"httpIdleTimeout" env:"HTTP_IDLE_TIMEOUT"`
		ClCurrency         string        `yaml:"clCurrency" env:"CL_CURRENCY"`
		ClCurrencyDivisor  int64         `yaml:"clCurrencyDivisor" env:"CL_CURRENCY_DIVISOR"`
		ClCurrencyDecimals int64         `yaml:"clCurrencyDecimals" env:"CL_CURRENCY_DECIMALS"`
		ElCurrency         string        `yaml:"elCurrency" env:"EL_CURRENCY"`
		ElCurrencyDivisor  int64         `yaml:"elCurrencyDivisor" env:"EL_CURRENCY_DIVISOR"`
		ElCurrencyDecimals int64         `yaml:"elCurrencyDecimals" env:"EL_CURRENCY_DECIMALS"`
		MainCurrency       string        `yaml:"mainCurrency" env:"MAIN_CURRENCY"`
	} `yaml:"frontend" env:", prefix=FRONTEND_"`
	Metrics struct {
		Enabled    bool   `yaml:"enabled" env:"ENABLED"`
		Address    string `yaml:"address" env:"ADDRESS"`
		Pprof      bool   `yaml:"pprof" env:"PPROF"`
		PprofExtra bool   `yaml:"pprofExtra" env:"PPROF_EXTRA"`
	} `yaml:"metrics" env:", prefix=METRICS_"`
	Notifications struct {
		UserDBNotifications                           bool    `yaml:"userDbNotifications" env:"USERDB_NOTIFICATIONS_ENABLED"`
		FirebaseCredentialsPath                       string  `yaml:"firebaseCredentialsPath" env:"NOTIFICATIONS_FIREBASE_CRED_PATH"`
		ValidatorBalanceDecreasedNotificationsEnabled bool    `yaml:"validatorBalanceDecreasedNotificationsEnabled" env:"VALIDATOR_BALANCE_DECREASED_NOTIFICATIONS_ENABLED"`
		PubkeyCachePath                               string  `yaml:"pubkeyCachePath" env:"NOTIFICATIONS_PUBKEY_CACHE_PATH"`
		OnlineDetectionLimit                          int     `yaml:"onlineDetectionLimit" env:"ONLINE_DETECTION_LIMIT"`
		OfflineDetectionLimit                         int     `yaml:"offlineDetectionLimit" env:"OFFLINE_DETECTION_LIMIT"`
		MachineEventThreshold                         uint64  `yaml:"machineEventThreshold" env:"MACHINE_EVENT_THRESHOLD"`
		MachineEventFirstRatioThreshold               float64 `yaml:"machineEventFirstRatioThreshold" env:"MACHINE_EVENT_FIRST_RATIO_THRESHOLD"`
		MachineEventSecondRatioThreshold              float64 `yaml:"machineEventSecondRatioThreshold" env:"MACHINE_EVENT_SECOND_RATIO_THRESHOLD"`
	} `yaml:"notifications"`
	SSVExporter struct {
		Enabled bool   `yaml:"enabled" env:"ENABLED"`
		Address string `yaml:"address" env:"ADDRESS"`
	} `yaml:"SSVExporter" env:", prefix=SSV_EXPORTER_"`
	RocketpoolExporter struct {
		Enabled bool `yaml:"enabled" env:"ENABLED"`
	} `yaml:"rocketpoolExporter" env:", prefix=ROCKETPOOL_EXPORTER_"`
	MevBoostRelayExporter struct {
		Enabled bool `yaml:"enabled" env:"ENABLED"`
	} `yaml:"mevBoostRelayExporter" env:", prefix=MEVBOOSTRELAY_EXPORTER_"`
	Pprof struct {
		Enabled bool   `yaml:"enabled" env:"ENABLED"`
		Port    string `yaml:"port" env:"PORT"`
	} `yaml:"pprof" env:", prefix=PPROF_"`
	NodeJobsProcessor struct {
		ElEndpoint string `yaml:"elEndpoint" env:"EL_ENDPOINT"`
		ClEndpoint string `yaml:"clEndpoint" env:"CL_ENDPOINT"`
	} `yaml:"nodeJobsProcessor" env:", prefix=NODE_JOBS_PROCESSOR_"`
	Monitoring struct {
		ApiKey                          string                           `yaml:"apiKey" env:"MONITORING_API_KEY"`
		ServiceMonitoringConfigurations []ServiceMonitoringConfiguration `yaml:"serviceMonitoringConfigurations" env:"SERVICE_MONITORING_CONFIGURATIONS"`
	} `yaml:"monitoring"`
	InternalAlerts InternalAlertDiscord `yaml:"internalAlerts"`

	ApiKeySecret     string   `yaml:"apiKeySecret" env:"API_KEY_SECRET"`
	CorsAllowedHosts []string `yaml:"corsAllowedHosts" env:"CORS_ALLOWED_HOSTS"`

	SkipDataAccessServiceInitWait bool `yaml:"skipDataAccessServiceInitWait" env:"SKIP_DATA_ACCESS_SERVICE_INIT_WAIT"`
}

type Chain struct {
	Name                       string `yaml:"name" env:"CHAIN_NAME"`
	Id                         uint64 `yaml:"id" env:"CHAIN_ID"`
	GenesisTimestamp           uint64 `yaml:"genesisTimestamp" env:"CHAIN_GENESIS_TIMESTAMP"`
	GenesisValidatorsRoot      string `yaml:"genesisValidatorsRoot" env:"CHAIN_GENESIS_VALIDATORS_ROOT"`
	DomainBLSToExecutionChange string `yaml:"domainBLSToExecutionChange" env:"CHAIN_DOMAIN_BLS_TO_EXECUTION_CHANGE"`
	DomainVoluntaryExit        string `yaml:"domainVoluntaryExit" env:"CHAIN_DOMAIN_VOLUNTARY_EXIT"`
	ClConfigPath               string `yaml:"clConfigPath" env:"CHAIN_CL_CONFIG_PATH"`
	ElConfigPath               string `yaml:"elConfigPath" env:"CHAIN_EL_CONFIG_PATH"`
	ClConfig                   ClChainConfig
	ElConfig                   *params.ChainConfig
}

type InternalAlertDiscord struct {
	DiscordWebhookUrl string `yaml:"discordWebhookUrl" env:"INTERNAL_ALERTS_DISCORD_WEBHOOK_URL"`
	DiscordUserName   string `yaml:"discordUserName" env:"INTERNAL_ALERTS_DISCORD_USER_NAME"`
	AvatarURL         string `yaml:"avatarURL" env:"INTERNAL_ALERTS_AVATAR_URL"`
}

type DatabaseConfig struct {
	Username     string `yaml:"user" env:"DB_USERNAME"`
	Password     string `yaml:"password" env:"DB_PASSWORD"`
	Name         string `yaml:"name" env:"DB_NAME"`
	Host         string `yaml:"host" env:"DB_HOST"`
	Port         string `yaml:"port" env:"DB_PORT"`
	MaxOpenConns int    `yaml:"maxOpenConns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `yaml:"maxIdleConns" env:"DB_MAX_IDLE_CONNS"`
	SSL          bool   `yaml:"ssl" env:"DB_SSL"`
	Failovers    []struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"failovers"`
}

type ServiceMonitoringConfiguration struct {
	Name     string        `yaml:"name" env:"NAME"`
	Duration time.Duration `yaml:"duration" env:"DURATION"`
}
