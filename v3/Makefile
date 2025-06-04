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
	chmod +x bin/beaconchain-api-service
	bin/beaconchain-api-service $(ARGS)
cr-create-db:
    # very basic machine for testing only
	gcloud sql instances create dev \
     --edition=enterprise \
     --database-version=POSTGRES_17 \
     --region=us-central1 \
     --project=${PROJECT_NAME} \
     --tier=db-f1-micro \
     --root-password=postgres
cr-deploy-staging:
	gcloud run deploy --source . --project staging-456307 --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
cr-deploy-personal:
	gcloud run deploy --source . --network default --project ${PROJECT_NAME} --region us-central1 --add-cloudsql-instances=${PROJECT_NAME}:us-central1:dev --args="--environment","personal_cloudrun"