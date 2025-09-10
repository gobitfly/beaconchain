all: generate-static-files
	buf build
	buf generate --template buf.gen.api.yaml
	buf generate --template buf.gen.domain.yaml
	mkdir -p bin
	go build -o ./bin/beaconchain-api-service ./cmd/main.go
generate-static-files: generate-proto build-http-client generate-mocks
	# No-op
generate-proto:
	rm -rf api/gen
	rm -rf domain/gen
	buf dep update
	buf build
	# Generate api protos
	buf generate --template buf.gen.api.yaml
	# Generate domain protos
	buf generate --template buf.gen.domain.yaml
build-http-client:
	go install github.com/go-swagger/go-swagger/cmd/swagger@latest
	rm -rf api/gen/client
	mkdir api/gen/client
	swagger generate client \
		-f api/gen/api_service/v1/external.swagger.json \
		-t api/gen/client \
		--name=ApiServiceClient
generate-mocks:
	go install github.com/vektra/mockery/v3@latest
	mockery
build-service:
	mkdir -p bin
	go build -o bin/beaconchain-api-service ./cmd/main.go
clean:
	rm -rf bin
lint: build-http-client
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
cr-deploy-staging: generate-proto build-http-client
	gcloud run deploy --source . --project staging-456307 --network mono-vpc --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
cr-deploy-personal:
	gcloud run deploy --source . --network default --project ${PROJECT_NAME} --region us-central1 --add-cloudsql-instances=${PROJECT_NAME}:us-central1:dev --args="--environment","personal_cloudrun"
cr-deploy-personal: generate-proto build-http-client
	gcloud builds submit --project ${PROJECT_NAME} --substitutions=_PROJECT_NAME="${PROJECT_NAME}" --config=build/package/cloudbuild.yaml
integration-test:
	TAGS=$(TAGS) ENV_FILE=test/testEnv/.env.$(NETWORK) go test -v $(TEST)

