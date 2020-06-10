![Docker](https://github.com/rtlnl/phoenix/workflows/Docker/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/rtlnl/phoenix)](https://goreportcard.com/report/github.com/rtlnl/phoenix)

# Phoenix

Phoenix is the **delivery recommendation systems** that is used at RTL Nederland. These APIs are able to deliver millions of recommendations per day. We use Phoenix for powering [Videoland](https://www.videoland.com/) and [RTL Nieuws](https://www.rtlnieuws.nl/). Our data science team works very hard to generate tailored recommendations to each user and we, as the platform team, make sure that these recommendations are actually delivered.

## How to start

The Project is divided in two main parts:

- Public APIs
- Internal APIs
- Worker

Assuming that you have `go`, `docker` and `docker-compose` installed in your machine, run `docker-compose up -d` to spin up Redis and localstack (for local S3).

After having the services up and running, assuming that you have your Go environment in your `PATH`, you should be able to start directly with `go run main.go --help`. This command should print the `help` message.

Proceed by running `go run main.go internal` for the internal APIs. In another terminal run `go run main.go public` for the public APIs and in a third terminal run `go run main.go worker` for the Worker service.

If you need to upload some files to the local S3, use the following commands after `localstack` has been created:

- `aws --endpoint-url=http://localhost:4572 s3 mb s3://my-bucket` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket my-bucket --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp /your/path/data.jsonl s3://my-bucket/data/` to copy a file to local S3

To know more in details how to use the system, go to the [wiki](https://github.com/rtlnl/phoenix/wiki) section in this repository. If you have any question, open an Issue and we will help you :rocket:

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
