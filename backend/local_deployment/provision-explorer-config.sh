#! /bin/bash
CL_PORT=$(kurtosis port print my-testnet cl-1-lighthouse-geth http --format number)
echo "CL Node port is $CL_PORT"

EL_PORT=$(kurtosis port print my-testnet el-1-geth-lighthouse rpc --format number)
echo "EL Node port is $EL_PORT"

REDIS_PORT=$(kurtosis port print my-testnet redis redis --format number)
echo "Redis port is $REDIS_PORT"

POSTGRES_PORT=$(kurtosis port print my-testnet postgres postgres --format number)
echo "Postgres port is $POSTGRES_PORT"

ALLOY_PORT=$(kurtosis port print my-testnet alloy alloy --format number)
echo "Alloy port is $ALLOY_PORT"

CLICKHOUSE_PORT=$(kurtosis port print my-testnet clickhouse clickhouse --format number)
echo "Clickhouse port is $CLICKHOUSE_PORT"

LBT_PORT=$(kurtosis port print my-testnet littlebigtable littlebigtable --format number)
echo "Little bigtable port is $LBT_PORT"

cat <<EOF > .env
CL_PORT=$CL_PORT
EL_PORT=$EL_PORT
REDIS_PORT=$REDIS_PORT
POSTGRES_PORT=$POSTGRES_PORT
ALLOY_PORT=$ALLOY_PORT
CLICKHOUSE_PORT=$CLICKHOUSE_PORT
LBT_PORT=$LBT_PORT
EOF

touch elconfig.json
cat >elconfig.json <<EOL
{
    "byzantiumBlock": 0,
    "constantinopleBlock": 0
}
EOL

touch config.yml

cat >config.yml <<EOL
justV2: false
chain:
  clConfigPath: 'node'
  elConfigPath: 'local_deployment/elconfig.json'
readerDatabase:
  name: db
  host: 127.0.0.1
  port: "$POSTGRES_PORT"
  user: postgres
  password: "pass"
writerDatabase:
  name: db
  host: 127.0.0.1
  port: "$POSTGRES_PORT"
  user: postgres
  password: "pass"
alloyReader:
  name: alloy
  host: 127.0.0.1
  port: "$ALLOY_PORT"
  user: postgres
  password: "pass"
alloyWriter:
  name: alloy
  host: 127.0.0.1
  port: "$ALLOY_PORT"
  user: postgres
  password: "pass"
clickhouse:
  readerDatabase:
    name: clickhouse
    host: 127.0.0.1
    port: "$CLICKHOUSE_PORT"
    user: postgres
    password: "pass"
  writerDatabase:
    name: clickhouse
    host: 127.0.0.1
    port: "$CLICKHOUSE_PORT"
    user: postgres
    password: "pass"
bigtable:
  project: explorer
  instance: explorer
  emulator: true
  emulatorPort: $LBT_PORT
eth1ErigonEndpoint: 'http://127.0.0.1:$EL_PORT'
eth1GethEndpoint: 'http://127.0.0.1:$EL_PORT'
redisCacheEndpoint: '127.0.0.1:$REDIS_PORT'
tieredCacheProvider: 'redis'
frontend:
  debug: true
  sessionSameSiteNone: false
  siteDomain: "localhost:8080"
  siteName: 'Open Source Ethereum (ETH) Testnet Explorer' # Name of the site, displayed in the title tag
  siteSubtitle: "Showing a local testnet."
  server:
    host: '0.0.0.0' # Address to listen on
    port: '8080' # Port to listen on
  readerDatabase:
    name: db
    host: 127.0.0.1
    port: "$POSTGRES_PORT"
    user: postgres
    password: "pass"
  writerDatabase:
    name: db
    host: 127.0.0.1
    port: "$POSTGRES_PORT"
    user: postgres
    password: "pass"
  sessionSecret: "11111111111111111111111111111111"
  jwtSigningSecret: "1111111111111111111111111111111111111111111111111111111111111111"
  jwtIssuer: "localhost"
  jwtValidityInMinutes: 30
  maxMailsPerEmailPerDay: 10
  mail:
    mailgun:
      sender: no-reply@localhost
      domain: mg.localhost
      privateKey: "key-11111111111111111111111111111111"
  csrfAuthKey: '1111111111111111111111111111111111111111111111111111111111111111'
  legal:
    termsOfServiceUrl: "tos.pdf"
    privacyPolicyUrl: "privacy.pdf"
    imprintTemplate: '{{ define "js" }}{{ end }}{{ define "css" }}{{ end }}{{ define "content" }}Imprint{{ end }}'

indexer:
  # fullIndexOnStartup: false # Perform a one time full db index on startup
  # indexMissingEpochsOnStartup: true # Check for missing epochs and export them after startup
  node:
    host: 127.0.0.1
    port: '$CL_PORT'
    type: lighthouse
  eth1DepositContractFirstBlock: 0

corsAllowedHosts: ["http://local.beaconcha.in:3000"]
EOL

echo "generated config written to config.yml"

echo "initializing bigtable schema"
PROJECT="explorer"
INSTANCE="explorer"
HOST="127.0.0.1:$LBT_PORT"
cd ..
go run ./cmd/misc/main.go -config local_deployment/config.yml -command initBigtableSchema

echo "bigtable schema initialization completed"

echo "provisioning postgres/clickhouse db schema"
go run ./cmd/misc/main.go -config local_deployment/config.yml -command applyDbSchema -target-version -2 -target-database postgres
go run ./cmd/misc/main.go -config local_deployment/config.yml -command applyDbSchema -target-version -2 -target-database clickhouse
echo "postgres/clickhouse db schema initialization completed"

echo "provisioning alloy db schema"
cd ../perfTesting
go run main.go -cmd seed -db.dsn postgres://postgres:pass@localhost:$ALLOY_PORT/alloy?sslmode=disable --seeder.validators 128 --seeder.users 5
cd ../backend/db_migrations
echo "migrating dp schemas"
goose postgres "postgres://postgres:pass@localhost:$ALLOY_PORT/alloy?sslmode=disable" reset
goose postgres "postgres://postgres:pass@localhost:$ALLOY_PORT/alloy?sslmode=disable" up
echo "alloy db schema initialization completed"

echo "adding test user"
HASHED_PW=$(htpasswd -nbBC 10 user password | cut -d ":" -sf 2)
psql postgres://postgres:pass@localhost:$POSTGRES_PORT/db?sslmode=disable -c "INSERT INTO users(password, email, email_confirmed) \
VALUES ('$HASHED_PW', 'test@beaconcha.in', true);"
echo "created test user with email 'test@beaconcha.in' and password 'password' "
