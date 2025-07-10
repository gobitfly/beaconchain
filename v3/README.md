# beaconchain-api

# Initial definition
PoC for what a new repo which implements a scalable maintainable API service might look like

Uses https://github.com/golang-standards/project-layout for the initial file structure. 
For more info on structure, see: https://github.com/golang-standards/project-layout/blob/master/README.md

# Goals

Implement a core set of APIs following the principles of DDD (Domain Driven Development), specifically the Repository Pattern.  

* https://threedots.tech/post/repository-pattern-in-go/ 

Good examples of the repository examples in practice:
* https://github.com/jorzel/go-repository-pattern



# Initial Setup

Before you can begin developing with this package, there are a few things you'll need to setup.

## Setup Buf

### What is Buf?
> Buf builds tooling to make schema-driven, Protobuf-based API development reliable and user friendly for service producers and consumers. Your organization shouldn't have to reinvent the wheel to work with Protobuf—our tools simplify your Protobuf management strategy so you can focus on what matters.

More Info: 
* https://buf.build/docs/ecosystem/
* https://buf.build/docs/tutorials/getting-started-with-buf-cli/

### Setup Steps

1. Install the Buf Cli: https://buf.build/docs/installation/. If on Mac, just run:

```
brew install bufbuild/buf/buf
```

To verify this is working correctly, you should be able to run `make generate-proto` from the workspace's root directory, and no error should be returned.

## Build and Run locally

```
cp configs/service/local.yaml.example configs/service/local.yaml
make generate-proto
docker compose -f deployments/docker-compose.yml up -d
```

In your browser, navigate to `http://localhost:8080/swagger-ui/#/BeaconchainService` to interact with the service. You can interact using curls against port 8080, or you can [grpcurl](https://github.com/fullstorydev/grpcurl) against 9090 (`grpcurl -plaintext localhost:9090 ExternalService/ExecutionBlock`)

## Debugging
Using visual studio code, navigate to main.go and click "Run" then "Start Debugging". Set break points before sending any requests (i.e. via the swagger link above or via curl).

## What is grpc and grpc-gateway?

https://github.com/grpc-ecosystem/grpc-gateway


## Run in CloudRun

You can deploy and run the service in either your personal environment or against the real staging environment

To deploy it to your personal project, note that you must initialize it first:
1. Fill in your .env and run `source .env`. Then create a cloud storage bucket to store terraform state & some configs in, and fill in `deployments/backend.hcl`
2. Create a container artifact registry: `make cr-create-registry`
3. Fill out configs (`default.yaml`: copy&paste example, `personal_cloudrun.yaml`: insert your project id), then push the service image: `make cr-deploy-personal` (if this step fails, try again after a few minutes for permission updates to propagate)
4. Install [terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli), enter hash from prev step in `terraform.tfvars:image` and run: `cd deployments && terraform init -backend-config=backend.hcl && terraform apply`

TODO
- Could combine step 1+2 into another small setup terraform
- Need to enable CI/CD in cloud run for changes to go live automatically on push. Until then you need to run steps 3 & 4 manually to update

To deploy it to staging, note that you are connecting to the shared staging database, so be careful of any modifying changes your service might execute.

```
gcloud run deploy --source . --project michael-test-454110 --network mono-vpc --region us-central1 --add-cloudsql-instances=staging-456307:us-central1:hoodi-d7ab6b26,staging-456307:us-central1:users-024879cd --args="--environment","staging"
```

## Run a Hybrid Setup
Hybrid refers to the Service running locally, while connected to resources in the cloud (i.e. a real psql database in Gcloud)

1. Set up the cloud proxies, which establish a secure connection to the databases 

```
./cloud-sql-proxy --auto-iam-authn staging-456307:us-central1:hoodi-d7ab6b26 --port 5432
./cloud-sql-proxy --auto-iam-authn staging-456307:us-central1:users-024879cd --port 5433
```

2. Build and run the service, specifying the environment as "hybrid". This uses the Hybrid config file specified at `configs/service/hybrid.yaml`.

```
make build-service && make run ARGS="--environment hybrid"
```

