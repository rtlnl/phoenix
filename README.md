![Docker](https://github.com/rtlnl/phoenix/workflows/Docker/badge.svg?branch=master) 

# Phoenix project

The Project is divided in two main parts:

- Public APIs
- Internal APIs
- Worker

Go to the [wiki](https://github.com/rtlnl/phoenix/wiki) for more information about these services.

## How to start

Assuming that you have `go`, `docker` and `docker-compose` installed in your machine, run `docker-compose up -d` to spin up Redis and localstack (for local S3).

After having the services up and running, assuming that you have your Go environment in your `PATH`, you should be able to start directly with `go run main.go --help`. This command should print the `help` message.

Proceed by running `go run main.go internal` for the internal APIs. In another terminal run `go run main.go public` for the public APIs and in a third terminal run `go run main.go worker` for the Worker service.

If you need to upload some files to the local S3, use the following commands after `localstack` has been created:

- `aws --endpoint-url=http://localhost:4572 s3 mb s3://my-bucket` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket my-bucket --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp /your/path/data.jsonl s3://my-bucket/data/` to copy a file to local S3

## How to run tests

To run all the tests, use the following command:

```bash
$: go clean -testcache && go test -race ./...
```

The first part is to avoid that Go will cache the result of the tests. This could lead to some evaluation errors
if you change some tests. Better without cache.

## How to perform manual tests

For manual testing of endpoints on the different environments, a [Postman collection](docs/postman/Phoenix.postman_collection.json) has been included in the `docs` directory.

## Deploy to Kubernetes

We made a convenient Helm chart so that you can deploy the project in your own cluster. Go to the folder `chart` for more information
