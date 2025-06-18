all: 
	buf build
	buf generate
	mkdir -p bin
	go build -o ./beaconchain-api-service ./cmd/main.go
generate-proto:
	rm -rf api/gen
	buf dep update
	buf build
	# Uses buf.gen.yaml and /api/protos/* to generate the api/gen proto files.
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
cr-create-registry:
	gcloud artifacts repositories create service \
		--repository-format docker \
		--location us-central1 \
		--project ${PROJECT_NAME}
cr-deploy-staging: generate-proto
	gcloud run deploy --source . --project staging-456307 --network mono-vpc --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
cr-deploy-personal: generate-proto
	gcloud builds submit --project ${PROJECT_NAME} --tag "us-central1-docker.pkg.dev/${PROJECT_NAME}/service/beaconchain-api:latest"

