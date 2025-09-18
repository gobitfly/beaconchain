all: generate-static-files
	buf build
	buf generate --template buf.gen.api.yaml
	buf generate --template buf.gen.domain.yaml
	mkdir -p bin
	go build -o ./bin/beaconchain-api-service ./cmd/main.go
generate-static-files: generate-proto generate-external-api generate-mocks
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
generate-external-api:
	rm -f api/external-client/client.gen.go
	rm -f api/external-server/external-server.gen.go
	rm -f api/external-server/external-types.gen.go
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go get github.com/oapi-codegen/nullable
	go generate ./...
generate-mocks:
	go install github.com/vektra/mockery/v3@latest
	mockery
build-service:
	mkdir -p bin
	go build -o bin/beaconchain-api-service ./cmd/main.go
clean:
	rm -rf bin
lint: generate-external-api
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
cr-deploy-staging: generate-proto generate-external-api
	gcloud run deploy --source . --project staging-456307 --network mono-vpc --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
cr-deploy-personal:
	gcloud run deploy --source . --network default --project ${PROJECT_NAME} --region us-central1 --add-cloudsql-instances=${PROJECT_NAME}:us-central1:dev --args="--environment","personal_cloudrun"
cr-deploy-personal: generate-proto generate-external-api
	gcloud builds submit --project ${PROJECT_NAME} --substitutions=_PROJECT_NAME="${PROJECT_NAME}" --config=build/package/cloudbuild.yaml
integration-test:
	TAGS=$(TAGS) ENV_FILE=test/testEnv/.env.$(NETWORK) go test -v $(TEST)

