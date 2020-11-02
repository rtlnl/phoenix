![Docker](https://github.com/rtlnl/phoenix/workflows/Docker/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/rtlnl/phoenix)](https://goreportcard.com/report/github.com/rtlnl/phoenix)

# Phoenix

<a href="https://ibb.co/j6Mv7b5"><img src="https://i.ibb.co/j6Mv7b5/gopher-phoenix.png" align="right" alt="gopher-phoenix" border="0"></a>

Phoenix is the **delivery recommendation systems** that is used at RTL Nederland. These APIs are able to deliver millions of recommendations per day. We use Phoenix for powering [Videoland](https://www.videoland.com/) and [RTL Nieuws](https://www.rtlnieuws.nl/). Our data science team works very hard to generate tailored recommendations to each user and we, as the platform team, make sure that these recommendations are actually delivered.

Simple, yet powerful API for delivery recommendations

* **Easy to understand** - push and get data from Redis very quickly
* **Fast in deliverying** - the combination of Go, Redis and Allegro cache makes the project blazing fast
* **Smart in storing** - the APIs avoid the overload of the Redis database by using a worker for bulk uplaod

We have being used in production since December 2019 and we haven't had a single downtime since. So far, we have delivered more than 350M recommendations to our users. The average request latency is `35ms`.

## Quick start

Assuming that you have `go`, `docker` and `docker-compose` installed in your machine, you need to have 3 terminals open that points to the directory where the project is. Do the following

1. In terminal number 1, run `docker-compose up -d` to spin up Redis and localstack (for local S3)
2. In terminal number 1 run `go run main.go worker` for the Worker service
3. In terminal number 2 run `go run main.go internal` for the Internal APIs
4. In terminal number 3 run `go run main.go public` for the Public APIs

Now you are ready to go :rocket:

## Docs

The documentation for developing and using Phoenix is available in the [wiki](https://github.com/rtlnl/phoenix/wiki)

## Join the Phoenix Community
In order to contribute to Phoenix, see the [CONTRIBUTING](CONTRIBUTING.md) file for how to go get started.
If your company or your product is using Phoenix, please let us know by adding yourself to the Phoenix [users](USERS.md) file.

## License
Phoenix is licensed under MIT license as found in the [LICENSE](LICENSE.md) file.