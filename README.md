# Data Personalization APIs

The Project is divided in two main parts:

- Public APIs
- Internal APIs

## How to start

Assuming that you have `docker` and `docker-compose` installed in your machine, run `docker-compose up -d` to spin up aerospike and localstack (for local S3).

After having the services up and running, assuming that you have your Go environment in your `PATH`, you should be able to start directly with `go run main.go --help`. This command should print the `help` message.

If you need to upload some files to the local S3, use the following commands after `localstack` has been created:

- `aws --endpoint-url=http://localhost:4572 s3 mb s3://test` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket test --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp ~/Desktop/data.csv s3://test/content/20190713/` to copy a file to local S3

## Aerospike

For testing we use two different namespaces specified in the `./conf/aerospike.conf` file. We use a custom settings-file to create a secondary namespace a logically divide testing from development.

## TODO

- [ ] Improve SWAG definitions for public APIs
- [ ] Improve SWAG definitions for internal APIs
- [ ] Add more tests with edge cases
