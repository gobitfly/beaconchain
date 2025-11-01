GITCOMMIT=`git describe --always`
VERSION=`git describe --always --tags`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
LDFLAGS="-X version.Version=${VERSION} -X version.BuildDate=${BUILDDATE} -X version.GitCommit=${GITCOMMIT} -X version.GitDate=${GITDATE} -s -w"
CGO_CFLAGS="-O -D__BLST_PORTABLE__"
CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

.PHONY: v1 v2

all: v1 v2 v3

v3: v3-generate-static-files
	cd v3 && buf build
	cd v3 && buf generate --template buf.gen.domain.yaml
	mkdir -p bin
	go build -o ./bin/v3/beaconchain-api-service ./v3/cmd/main.go
v3-generate-static-files: v3-generate-proto v3-generate-api-stubs v3-generate-mocks
	# No-op
v3-generate-proto:
	cd v3 && buf dep update
	cd v3 && buf build
	# Generate domain protos
	cd v3 && buf generate --template buf.gen.domain.yaml
v3-generate-api-stubs:
	go get -C v3/ -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go get -C v3/ github.com/oapi-codegen/nullable
	go generate -C v3/ ./...
v3-generate-mocks:
	go install github.com/vektra/mockery/v3@latest
	cd v3 && mockery
v3-build-service:
	mkdir -p bin/v3
	go build -o bin/v3/beaconchain-api-service ./v3/cmd/main.go
v3-lint: v3-generate-api-stubs
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	golangci-lint run --timeout 5m --config ./v3/.golangci.yml ./v3/...
	cd v3 && buf lint
v3-run:
	chmod +x bin/v3/beaconchain-api-service
	bin/v3/beaconchain-api-service $(ARGS)
v3-cr-create-registry:
	gcloud artifacts repositories create service \
		--repository-format docker \
		--location us-central1 \
		--project ${PROJECT_NAME}
v3-cr-deploy-staging: v3-generate-proto v3-generate-api-stubs
	gcloud run deploy --source . --project staging-456307 --network mono-vpc --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
v3-cr-deploy-personal:
	gcloud run deploy --source . --network default --project ${PROJECT_NAME} --region us-central1 --add-cloudsql-instances=${PROJECT_NAME}:us-central1:dev --args="--environment","personal_cloudrun"
v3-cr-deploy-personal: v3-generate-proto v3-generate-api-stubs
	gcloud builds submit --project ${PROJECT_NAME} --substitutions=_PROJECT_NAME="${PROJECT_NAME}" --config=build/package/cloudbuild.yaml
v3-integration-test:
	TAGS=$(TAGS) ENV_FILE=v3/test/testEnv/.env.$(NETWORK) go test -C v3/ -v $(TEST)

v2: v2-backend

v2-backend:
	mkdir -p bin/v2
	go install github.com/swaggo/swag/cmd/swag@latest && swag init --ot json -o ./v2/backend/pkg/api/docs -d ./v2/backend/pkg/api/ -g ./handlers/public.go
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/v2/bc ./v2/backend/cmd/main.go

v2-test:
	go test -C v2/backend ./...

v1: v1-explorer v1-stats v1-frontend-data-updater v1-rewards-exporter v1-eth1indexer v1-blobindexer v1-node-jobs-processor v1-signatures v1-misc v1-notification-sender v1-notification-collector v1-user-service v1-validator-tagger

v1-explorer:
	rm -rf bin/v1
	mkdir -p bin/v1
	echo "Creating frontend asset bundle..."
	go run v1/cmd/bundle/main.go
	echo "Bundling API docs..."
	npm run bundle-api-docs
	echo "Building explorer..."
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/explorer v1/cmd/explorer/main.go

v1-stats:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/statistics v1/cmd/statistics/main.go

v1-frontend-data-updater:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/frontend-data-updater v1/cmd/frontend-data-updater/main.go

v1-rewards-exporter:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/rewards-exporter v1/cmd/rewards-exporter/main.go

v1-eth1indexer:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/eth1indexer v1/cmd/eth1indexer/main.go

v1-blobindexer:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/blobindexer v1/cmd/blobindexer/blobindexer.go

v1-node-jobs-processor:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/node-jobs-processor v1/cmd/node-jobs-processor/main.go

v1-signatures:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/signatures v1/cmd/signatures/main.go

v1-misc:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/misc v1/cmd/misc/main.go

v1-notification-sender:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/notification-sender v1/cmd/notification-sender/main.go

v1-notification-collector:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/notification-collector v1/cmd/notification-collector/main.go

v1-user-service:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/user-service v1/cmd/user-service/main.go

v1-validator-tagger:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/v1/validator-tagger v1/cmd/validator-tagger/main.go