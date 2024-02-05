GITCOMMIT=`git describe --always`
VERSION=`git describe --always --tags`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=github.com/gobitfly/beaconchain/commons
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"

all: exporter

clean:
	rm -rf bin

test:
	go test ./...

binaries:
	mkdir -p bin

exporter: binaries
	cd exporter; go build --ldflags=${LDFLAGS} -o ../bin/exporter ./cmd/main.go

api: binaries
	cd api; go build --ldflags=${LDFLAGS} -o ../bin/api ./cmd/main.go
	
