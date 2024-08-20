GITCOMMIT=`git rev-parse --short=8 HEAD`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
GITDATESHORT=$$(TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD | sed 's/[-T:]//g' | sed 's/\(+.*\)$$//g')
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
VERSION=${GITDATESHORT}-${GITCOMMIT}
PACKAGE=github.com/gobitfly/beaconchain/pkg/commons
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"
CGO_CFLAGS="-O -D__BLST_PORTABLE__"
CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

all: exporter blobindexer misc ethstore-exporter api api-docs rewards-exporter eth1indexer stats user-service notification-sender notification-collector node-jobs-processor signatures

clean:
	rm -rf bin

test:
	go test ./...

binaries:
	mkdir -p bin

exporter: binaries
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/exporter ./cmd/exporter/main.go

blobindexer: binaries
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/blobindexer ./cmd/blobindexer/main.go

misc: binaries
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/misc ./cmd/misc/main.go

api: binaries
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/api ./cmd/api/main.go

ethstore-exporter:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/ethstore-exporter ./cmd/ethstore_exporter/main.go

rewards-exporter:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/rewards-exporter ./cmd/rewards_exporter/main.go

eth1indexer:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/eth1indexer ./cmd/eth1indexer/main.go

stats:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/statistics ./cmd/statistics/main.go

user-service:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/user-service ./cmd/user_service/main.go

notification-sender:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/notification-sender ./cmd/notification_sender/main.go

notification-collector:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/notification-collector ./cmd/notification_collector/main.go

node-jobs-processor:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/node-jobs-processor ./cmd/node_jobs_processor/main.go

signatures:
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o bin/signatures cmd/signatures/main.go

frontend-types:
	go run cmd/typescript_converter/main.go --out ../frontend/types/api

addhooks:
	git config core.hooksPath hooks

api-docs:
	go run cmd/api_docs/main.go