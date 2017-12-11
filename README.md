# Auklet C/C++ Profiler

Auklet's IoT C/C++ profiler is built to run on any POSIX operating system. It
has been validated on:

- Ubuntu 16.04

It consists of three components:

## `libauklet.a`

A C library that your program is linked against at compile time.

## `release`

A deploy-time command-line tool that sends symbol information from the profiled
program to the backend.

## `wrap`

A command-line program that runs the program to be profiled and continuously
sends live profile data to the backend.

# Go Setup

`wrap` and `release` need at least Go 1.8. See the [getting started page][gs] to
download Go. Then see [How to Write Go Code - Organization][org] to set up your
system.

[gs]: https://golang.org/doc/install
[org]: https://golang.org/doc/code.html#Organization

Conventionally, your `~/.profile` should contain the following:

	export GOPATH=$HOME/go
	export PATH=$PATH:$GOPATH/bin

The first line tells Go where your workspace is located. The second makes sure
that the shell will know about executables built with `go install`.

`wrap` needs several third-party packages. Install them with

	go get ./wrap

# Build

To build and install all components, run

	make

In particular, this installs the commands `wrap` and `release` to `$GOPATH/bin`,
and build test executables `x` and `x-dbg`.

To install the static library `libauklet.a` to `/usr/local/lib/`, run

	make install

# Run Unit Tests

	./lib_test

# Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY
	AUKLET_BASE_URL
	AUKLET_BROKERS
	AUKLET_PROF_TOPIC
	AUKLET_EVENT_TOPIC
	AUKLET_CA
	AUKLET_CERT
	AUKLET_PRIVATE_KEY

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage multiple configurations, it is suggested to define
the envars in a shell script named after the configuration; for example,
`.env.staging`.

The variables `AUKLET_API_KEY` and `AUKLET_APP_ID` are likely to be different
among developers, so it is suggested that they be defined in a separate
file, `.auklet`, and sourced from within `.env.staging`. For example:

	$ cat .auklet
	export AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	export AUKLET_API_KEY=SM49BAMCA0...

	$ cat .env.staging
	. .auklet
	export AUKLET_BASE_URL=https://api-staging.auklet.io/v1
	export AUKLET_BROKERS=broker1,broker2,broker3
	export AUKLET_PROF_TOPIC=z8u1-profiler
	export AUKLET_EVENT_TOPIC=z8u1-events
	export AUKLET_CA=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS...
	export AUKLET_CERT=LS0tLS1CRUdJTiBDRVJUSUZJQ0FU...
	export AUKLET_PRIVATE_KEY=LS0tLS1CRUdJTiBQUklW...

## `AUKLET_BROKERS`

A comma-delimited list of Kafka broker addresses. For example:

	broker1,broker2,broker3

## `AUKLET_EVENT_TOPIC`, `AUKLET_PROF_TOPIC`

Kafka topics to which `wrap` should send event and profile data, respectively.

## `AUKLET_BASE_URL`

A URL, without a trailing slash, to be used when creating and checking releases.
It is accessed by both `wrap` and `release` commands. For example:

	https://api-staging.auklet.io/v1

If not defined, `wrap` and `release` default to the production endpoint.

## `AUKLET_CA`, `AUKLET_CERT`, `AUKLET_PRIVATE_KEY`

Base64-encoded [PEM][pem]-format certs that `wrap` uses to authenticate its Kafka
connection.

[pem]: https://en.wikipedia.org/wiki/Privacy-enhanced_Electronic_Mail

# Assign a Configuration

	. .env.staging

# Release an App

To release an executable called `x`, create an executable in the same directory
called `x-dbg` that contains debug information. (`x` is not required to contain
debug info.) Then run

	release x

# Run an App

	wrap ./x

