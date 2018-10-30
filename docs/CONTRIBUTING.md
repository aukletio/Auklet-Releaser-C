# How to Contribute

## Standards

Auklet is an edge first application performance monitor; therefore, starting 
with version 1.0.0 the following compliance levels are to be maintained:

- Automotive Safety Integrity Level B (ASIL B)

## Submissions

To submit code changes, please open a pull request that lists and explains all 
changes.

If you have found a bug please check the submitted issues. If you do not see 
your issue listed please open a new issue and we will respond as quickly as 
possible. 

If you've found a security related bug, please email our team 
directly at [hello@auklet.io](mailto:hello@auklet.io). 

# Working on the Auklet C Releaser

## Go Setup

`release` needs at least Go 1.8 and [dep][godep] 0.3.2. See the
[getting started page][gs] to download Go. Then see [How to Write Go Code -
Organization][org] to set up your system.

[godep]: https://github.com/golang/dep
[gs]: https://golang.org/doc/install
[org]: https://golang.org/doc/code.html#Organization

Conventionally, your `~/.profile` should contain the following:

	export GOPATH=$HOME/go
	export PATH=$PATH:$GOPATH/bin

The first line tells Go where your workspace is located. The second makes sure
that the shell will know about executables built with `go install`.

After setting up Go on your system, install `dep` by running:

	curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
	chmod +x $GOPATH/bin/dep

If you want to build `release` on Mac OS X, you can install `dep` via
Homebrew by running `brew install dep`, or by changing the above `curl` command
to download `dep-darwin-amd64`.

After cloning this repo and setting up your Go environment, run this command to enable pre-commit gofmt checking: `git config core.hookspath .githooks`.

## Development Tools

`autobuild` is an optional script that can be run in a separate terminal window.
When source files change, it runs `go install ./cmd/release`, allowing the
developer to find compile-time errors immediately without needing an IDE.

`autobuild` requires [entr](http://www.entrproject.org/).

## Build

To ensure you have all the correct dependencies, run

	dep ensure

To build and install the releaser to `$GOPATH/bin`, run

	go install ./release

## Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage multiple configurations, it is suggested to define
the envars in a shell script named after the configuration; for example,
`.env.staging`.

The variables `AUKLET_API_KEY` and `AUKLET_APP_ID` are likely to be different
among developers, so it is suggested that they be defined in a separate
file, `.env`. For example:

	$ cat .env
	export AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	export AUKLET_API_KEY=SM49BAMCA0...

## Assign a Configuration

	. .env

## Release an App

To release an executable called `x`, create an executable in the same directory
called `x-dbg` that contains debug information. (`x` is not required to contain
debug info.) Then, run

	release x

## Docker Setup

1. Install [Docker](www.docker.com/products/docker-desktop).
1. Build your environment with `docker-compose build`.
1. To ensure you have all the correct dependencies, run `docker-compose run auklet dep ensure`.
1. To build and install the releaser to `$GOPATH/bin`, run `docker-compose run auklet go install ./cmd/release`.
