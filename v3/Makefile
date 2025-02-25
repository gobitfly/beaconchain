all: 
	buf build
	buf generate
	mkdir -p bin
	go build -o ./beaconchain-api-service ./cmd/main.go
generate-proto:
	rm -rf api/gen
	buf build
	# Uses bug.gen.yaml and /api/protos/* to generate the api/gen proto files.
	# Can be rate-limited, so don't spam it
	buf generate
build-service:
	mkdir -p bin
	go build -o bin/beaconchain-api-service ./cmd/main.go
clean:
	rm -rf bin
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
	golangci-lint run --timeout 5m
	buf lint
run:
	chmod +x beaconchain-api-service
	./beaconchain-api-service

