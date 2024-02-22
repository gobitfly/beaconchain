GITCOMMIT=`git describe --always`
VERSION=`git describe --always --tags`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=github.com/gobitfly/beaconchain/commons
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"

all: exporter blobindexer misc

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
	
