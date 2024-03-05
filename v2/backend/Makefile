GITCOMMIT=`git describe --always`
VERSION=`git describe --always --tags`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=github.com/gobitfly/beaconchain/pkg/commons
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"

all: exporter blobindexer misc ethstore-exporter api rewards-exporter eth1indexer stats user-service notification-sender notification-collector node-jobs-processor signatures

clean:
	rm -rf bin

test:
	go test ./...

binaries:
	mkdir -p bin

exporter: binaries
	go build --ldflags=${LDFLAGS} -o ./bin/exporter ./cmd/exporter/main.go

blobindexer: binaries
	go build --ldflags=${LDFLAGS} -o ./bin/blobindexer ./cmd/blobindexer/main.go

misc: binaries
	go build --ldflags=${LDFLAGS} -o ./bin/misc ./cmd/misc/main.go

api: binaries
	go build --ldflags=${LDFLAGS} -o ./bin/api ./cmd/api/main.go

ethstore-exporter:
	go build --ldflags=${LDFLAGS} -o bin/ethstore-exporter ./cmd/ethstore_exporter/main.go

rewards-exporter:
	go build --ldflags=${LDFLAGS} -o bin/rewards-exporter ./cmd/rewards_exporter/main.go

eth1indexer:
	go build --ldflags=${LDFLAGS} -o bin/eth1indexer ./cmd/eth1indexer/main.go

stats:
	go build --ldflags=${LDFLAGS} -o bin/statistics ./cmd/statistics/main.go

user-service:
	go build --ldflags=${LDFLAGS} -o bin/user-service ./cmd/user_service/main.go

notification-sender:
	go build --ldflags=${LDFLAGS} -o bin/notification-sender ./cmd/notification_sender/main.go

notification-collector:
	go build --ldflags=${LDFLAGS} -o bin/notification-collector ./cmd/notification_collector/main.go

node-jobs-processor:
	go build --ldflags=${LDFLAGS} -o bin/node-jobs-processor ./cmd/node_jobs_processor/main.go

signatures:
	go build --ldflags=${LDFLAGS} -o bin/signatures cmd/signatures/main.go

addhooks:
	git config core.hooksPath hooks