# Data Personalization APIs

The Project is divided in two main parts:

- Public APIs
- Internal APIs

Public APIs are meant to be the frontend for the clients where they can ask the recommended items for a given user.
The request will be processed against the Redis Cluster so that the client can have a quick result due to the fast
retrival capabilities of Redis. The Public APIs are meant to be Read-Only.

To retrieve the information from our service the workflow is as follow:

1) `POST` request containing the `user_ID`
2) Given the date of the request we `query` the `history` key space with `history:YYYMMDD` that returns the `UUID` `[Time complexity O(1)]`
3) Combining `UUID:user:ID` we can access the values (if any) `[Time complexity O(1)]`

The time complexity for these operations are:

```txt
O(1) + O(1) = O(1)
```

which is the fastest possible access in terms of complexity.
If there are no values, the API will return a 404 not found which will tell the client to serve random content.

The Internal APIs have the important task of populating and deleting recommendations. The process of population and
deletion is strictly connected to the data-structure that we are using. There are two main data structures:

1) History
2) Actual recommendations

The History is defined as follow:

```txt
Key   => history:YYYYMMDD
Value => UUID
```

Redis doesn't have a concept of table but we can use part of the key to create a `"key space"`. This will help indexing
the values for a quick retrival. The reason why we add a date is to identify which are the latest inserted recommendations.

The Actual recommendations data structure is defined as follow:

```txt
Key   => UUID:user:ID
Value => [ ... ]
```

The UUID is the same as defined above and the ID comes from the file that contains the recommendations. We expect that the
file contains a structure similar to `user_id,[ ... ]`.

The efficiency in this storing system stays in the moment that we need to insert and delete the recommended items.
The insertion workflow is defined as follow:

1) Create a new entry in the `history` key space with date time of the request
2) Pull the file from S3 (if any)
3) Iterate each line and add an entry in the database

The insertion part is rather easy to do. Deletion instead is a bit more complicated

The deletion part needs some thinking in order to avoid collision in the key space as well as to avoid deleting the wrong keys.
The major issues stands in selecting the correct `UUID` to delete and avoid long `LOCKS` on the cluster when deleting that.
Fortunately, Redis comes with a command called `SCAN` that will return part of the `KEYS` that follows a certain patter helping us in
`DELETING` effeciently without long `LOCKS`. Although, we need to pay attention to the `Time complexity`. Here is why:

- `SCAN` operation is `O(n)`
- `DEL` operation is `O(1)`

Given the above, the final time complexity is still `O(n)` which is an acceptable time considering that SCAN will only read a certain
amount of items given the `keys` pattern. To improve reliability we use `PIPELINE` command to chain `SCAN + DEL` of the items. The
`PIPELINE` basically allows the server to execute multiple commands without returning all the `OK+` messages. This speeds up  the query time.

The deletion operation is done in the following manner:

1) Deletion request arrives at a particular `YYYY-MM-DD`
2) Scan the `history:*` `KEYS` and sort the list by `ASC`
3) Pick the first `history:YYYYMMDD` and `GET` `UUID` value
4) Create a `PIPELINE` and
 4.1) Execute a `SCAN` for `UUUID:user:*` for max 1000 items
 4.2) Execute `DEL` on `UUUID:user:n`
5) return `OK+` to client

Failover is always something that we need to take care of. Redis claims that due to the Atomicity and `LOCK` system on the server when
executing operations errors are very difficult to get. The possible errors are in the data format.
Having a strong validation on line-by-line will allow us to store all the new data incoming without breaking the workflow. In fact,
Redis *doesn't have* a `roll-back` mechanisms because it is not Transactional. Although, atomicity guarantees that the information is
stored properly without errors.

## How to start

Here are some commands

- `brew install redis` to install the redis-cli
- `docker-compose up -d` to spin up redis-cluster and localstack
- `aws --endpoint-url=http://localhost:4572 s3 mb s3://test` to create a bucket in local S3
- `aws --endpoint-url=http://localhost:4572 s3api put-bucket-acl --bucket test --acl public-read` to set up a policy for testing with local s3
- `aws --endpoint-url=http://localhost:4572 s3 cp ~/Desktop/personalization.csv s3://test/content/20190713/` to copy a file to local S3
