GITCOMMIT=`git rev-parse --short=8 HEAD`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
GITDATESHORT=$$(TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD | sed 's/[-T:]//g' | sed 's/\(+.*\)$$//g')
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
VERSION=${GITDATESHORT}-${GITCOMMIT}
PACKAGE=github.com/gobitfly/beaconchain/pkg/commons
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE} -s -w"
CGO_CFLAGS="-O -D__BLST_PORTABLE__"
CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

all: 
	mkdir -p bin
	CGO_CFLAGS=${CGO_CFLAGS} CGO_CFLAGS_ALLOW=${CGO_CFLAGS_ALLOW} go build --ldflags=${LDFLAGS} -o ./bin/bc ./cmd/main.go

clean:
	rm -rf bin

test:
	go test ./...

frontend-types:
	go run cmd/main.go typescript-converter --out ../frontend/types/api

addhooks:
	git config core.hooksPath hooks

api-docs:
	go run cmd/main.go api_docs