PACKAGE=eth2-staking-pool
all: seeder

clean:
	rm -rf bin

test:
	go test ./...

binaries:
	mkdir -p bin

seeder: binaries
	go build -o bin/seeder ./main.go

