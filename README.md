# Data Personalization APIs

The Project is divided in two main parts:

- Public APIs
- Internal APIs

## How to start

Assuming that you have `go`, `docker` and `docker-compose` installed in your machine, run `docker-compose up -d` to spin up aerospike and localstack (for local S3).

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

## Aerospike

For testing we use two different namespaces specified in the `./conf/aerospike.conf` file. We use a custom settings-file to create a secondary namespace a logically divide testing from development.

## How it works internally

The Project is divided in two main parts:

- Public APIs
- Internal APIs

The Public APIs have the only job of fetching the recommendations from the database given a specified model in the request body.
For more information about "How to create the request", check the endpoint "/docs". A swagger page should appear.

The Internal APIs do the hard working on handling the creation/deletion/update of models and data. The database we chose for serving
the recommendations is Aerospike. Aerospike is a Key/Value storage database similar to Redis but with steroids. It has a concept of
tables (and columns) which makes it easier to create a structured definition of both models and data.

Leaving aside how Aerospike works (check the documentation for more information [https://www.aerospike.com/docs/](https://www.aerospike.com/docs/)), it is important
to point out how we designed the internal system.
We have two concepts:

- **model**
- **data**

The model is a simple "table" in which contains the metadata of model itself. This is needed for then creating the appropriate Key in
the "data" table(s). The "data" is organized similarly but with different specifications.
In aerospike the hierarchy is the following (where `-->` means `"contains many uniques items of"`):

```bash
Namespace --> setName --> Keys --> Bins
```

### The Models

The namespace is the main `container` for all the data we want to store. Hence, it is common for both concepts. 
The setname (you can think this as the table in RDBMS) corresponds to the `publicationPoint`. Since we can have multiple campaigns for
the same publicationPoint, the Key becomes the campaign. The Bins are the `values` of the `Key`. Each Bin has also a `key/value` pair
and the entries are specified in the schema above.
Every time an action is done on the model (publish the model, stage the model, etc), the Version is either increased/decreased based
on the `SemVer` algorithm.

```bash
Namespace: personalization
SetName: publicationPoint
Key: Campaign
Bins: version => 0.1.0 				 // as start
	  stage => STAGED/PUBLISHED		 // either value
	  signalType => articleID_userID // this is an example
```

Below you can find an example of multiple models:

```bash
- Model1
	SetName = rtl_news
	Key = homepage
- Model2
	SetName = rtl_news
	Key = footer
- Model3
	SetName = videoland
	Key = profile
```

### The Data

The data is organized similarly to the Model but with a different naming convention. To make the SetName `unique` per model, we
use a combination of `publicPoint` and `campaign`. In this way, we are able to insert all the Keys we
need for that particular model

```bash
Namespace: personalization
SetName: publicationPoint#campaign
Key: signalID	// for example 111_3333
Bins: signalID => ["item1", "item2", ..., "itemN"]
```

Below you can find and example of data for a model

```bash
- Data1
	SetName = rtl_news#homepage
		- Key = 11_22
		  Bins = 11_22 = ["1","2","3"]
		- Key = 33_44
		  Bins = 33_44 = ["4","5","6"]
		- Key = 55_66
		  Bins = 55_66 = ["7","8","9"]
- Data2
	SetName = rtl_news#footer
		- Key = 3333
		  Bins = 3333 = ["a","b","c"]
		- Key = 4444
		  Bins = 4444 = ["d","e","f"]
		- Key = 5555
		  Bins = 5555 = ["g","h","i"]
```

## Batch Upload

The APIs gives the possibility to read a file from S3 and upload it to Aerospike. To avoid timeouts and having the client hanging waiting for the response, the APIs has a simple `checking` mechanism. The picture below explain how the process works

![](/docs/images/batch_upload.png)

The process of uploading the file from S3 to Aerospike is delegated to a separate `go routine`. The client should store the `batchID` that is returned from the initial request `(POST /batch)` and ask for the status with `GET /batch/status/:id`.

Time taken to upload **1.6M unique keys** from S3 is `3m 33secs`. Check this [PR](https://github.com/rtlnl/data-personalization-api/pull/5) for more information

