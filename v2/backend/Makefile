PACKAGE=beaconchain
all: exporter

clean:
	rm -rf bin

test:
	go test ./...

binaries:
	mkdir -p bin

exporter: binaries
	go build -o bin/exporter ./exporter/cmd/main.go

api: binaries
	go build -o bin/api ./api/cmd/main.go

