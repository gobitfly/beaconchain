all: 
	buf build
	buf generate
	mkdir -p bin
	go build -o ./bin/beaconchain-api-service ./cmd/main.go
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
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
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
cr-deploy-personal:
	gcloud run deploy --source . --network default --project ${PROJECT_NAME} --region us-central1 --add-cloudsql-instances=${PROJECT_NAME}:us-central1:dev --args="--environment","personal_cloudrun"
cr-deploy-personal: generate-proto
	gcloud builds submit --project ${PROJECT_NAME} --substitutions=_PROJECT_NAME="${PROJECT_NAME}" --config=build/package/cloudbuild.yaml
integration-test:
	TAGS=$(TAGS) ENV_FILE=test/testEnv/.env.$(NETWORK) go test -v $(TEST)

