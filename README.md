![Docker](https://github.com/rtlnl/phoenix/workflows/Docker/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/rtlnl/phoenix)](https://goreportcard.com/report/github.com/rtlnl/phoenix)

# Phoenix

Phoenix is the **delivery recommendation systems** that is used at RTL Nederland. These APIs are able to deliver millions of recommendations per day. We use Phoenix for powering [Videoland](https://www.videoland.com/) and [RTL Nieuws](https://www.rtlnieuws.nl/). Our data science team works very hard to generate tailored recommendations to each user and we, as the platform team, make sure that these recommendations are actually delivered.

Simple, yet powerful API for delivery recommendations

* **Easy to understand** - push and get data from Redis very quickly
* **Fast in deliverying** - the combination of Go, Redis and Allegro cache makes the project blazing fast
* **Smart in storing** - the APIs avoid the overload of the Redis database by using a worker for bulk uplaod

We have being used in production since December 2019 and we haven't had a single downtime since. So far, we have delivered more than 350M recommendations to our users. The average request latency is `35ms`.

## Docs

The documentation for developing and using Phoenix is available in the [wiki](https://github.com/rtlnl/phoenix/wiki)

## Join the Phoenix Community
In order to contribute to Phoenix, see the [CONTRIBUTING](CONTRIBUTING.md) file for how to go get started.
If your company or your product is using Phoenix, please let us know by adding yourself to the Phoenix [users](USERS.md) file.

## License
Phoenix is licensed under MIT license as found in the [LICENSE](LICENSE.md) file.
