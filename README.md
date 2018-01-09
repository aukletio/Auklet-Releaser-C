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

`wrap` and `release` need at least Go 1.8 and [dep][godep] 0.3.2. See the [getting started page][gs] to
download Go. Then see [How to Write Go Code - Organization][org] to set up your
system.

[godep]: https://github.com/golang/dep
[gs]: https://golang.org/doc/install
[org]: https://golang.org/doc/code.html#Organization

Conventionally, your `~/.profile` should contain the following:

	export GOPATH=$HOME/go
	export PATH=$PATH:$GOPATH/bin

The first line tells Go where your workspace is located. The second makes sure
that the shell will know about executables built with `go install`.

After setting up Go on your system, install `dep` by running:

```
curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
chmod +x $GOPATH/bin/dep
```

If you want to build `wrap` and `release` on Mac OS X, you can install `dep` via Homebrew by running `brew install dep`, or by changing the above `curl` command to download `dep-darwin-amd64`.

# Development Tools

`autobuild` is an optional script that can be run in a separate terminal window.
When source files change, it runs `make`, allowing the developer to find
compile-time errors immediately without needing an IDE.

`autobuild` requires [entr](http://www.entrproject.org/).

# Build

To ensure you have all the correct dependencies, run

	dep ensure

To build and install all components, run

	make

In particular, this installs the commands `wrap` and `release` to `$GOPATH/bin`,
and the static library `libauklet.a` to `/usr/local/lib/`.

It also builds test executables `x` and `x-dbg`.

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

## `AUKLET_BROKERS`

A comma-delimited list of Kafka broker addresses. For example:

	broker1,broker2,broker3

## `AUKLET_EVENT_TOPIC`, `AUKLET_PROF_TOPIC`

Kafka topics to which `wrap` should send event and profile data,
respectively.

## `AUKLET_BASE_URL`

A URL, without a trailing slash, to be used when creating and checking releases.
It is accessed by both `wrap` and `release` commands. For example:

	https://api-staging.auklet.io/v1

If not defined, `wrap` and `release` default to the production endpoint.

# Assign a Configuration

	. .env.staging

# Release an App

To release an executable called `x`, create an executable in the same directory
called `x-dbg` that contains debug information. (`x` is not required to contain
debug info.) Then run

	release x

# Run an App

	wrap ./x

