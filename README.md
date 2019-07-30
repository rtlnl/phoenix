# Data Personalization APIs

The Project is divided in two main parts:

- Public APIs
- Internal APIs

## How to start

Here are some commands

- `docker-compose up -d` to spin up aerospike and localstack
- `aws --endpoint-url=http://localhost:4572 s3 mb s3://test` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket test --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp ~/Desktop/personalization.csv s3://test/content/20190713/` to copy a file to local S3
