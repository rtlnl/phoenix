# Phoenix project

The Project is divided in two main parts:

- Public APIs
- Internal APIs

## How to start

Assuming that you have `go`, `docker` and `docker-compose` installed in your machine, run `docker-compose up -d` to spin up Redis and localstack (for local S3).

After having the services up and running, assuming that you have your Go environment in your `PATH`, you should be able to start directly with `go run main.go --help`. This command should print the `help` message.

Proceed by running `go run main.go internal` for the internal APIs (or `go run main.go public` for the public APIs)

If you need to upload some files to the local S3, use the following commands after `localstack` has been created:

- `aws --endpoint-url=http://localhost:4572 s3 mb s3://test` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket test --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp ~/Desktop/data.csv s3://test/content/20190713/` to copy a file to local S3

## How to run tests

To run all the tests, use the following command:

```bash
$: go clean -testcache && go test -race ./...
```

The first part is to avoid that Go will cache the result of the tests. This could lead to some evaluation errors
if you change some tests. Better without cache.

## How to perform manual tests

For manual testing of endpoints on the different environments, a [Postman collection](docs/postman/Phoenix.postman_collection.json) has been included in the `docs` directory.

## Batch Upload

The APIs gives the possibility to read a file from S3 and upload it to Redis. To avoid timeouts and having the client hanging waiting for the response, the APIs has a simple `checking` mechanism. The picture below explain how the process works

![](/docs/images/batch_upload.png)

The process of uploading the file from S3 to Redis is delegated to a separate `go routine`. The client should store the `batchID` that is returned from the initial request `(POST /v1/batch)` and ask for the status with `GET /v1/batch/status/:id`.

Time taken to upload **1.6M unique keys** from S3 is `3m 33secs`. Check this [PR](https://github.com/rtlnl/phoenix/pull/5) for more information
