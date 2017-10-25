# APM Profiler

The profiler consists of three components:

## `rt.o`

A C library that your program is linked against at compile time.

## `release`

A deploy-time command-line tool that sends symbol information from the profiled
program to the backend.

## `wrap`

A command-line program that runs the program to be profiled and continuously
sends live profile data to the backend.

# Build

	make

# Run Unit Tests

	./lib_test

# Configure

An Auklet configuration is a directory with the following structure, where
`staging` is the name of the configuration:

	staging/
		broker
		event
		prof
		url
		cert/
			ck_ca
			ck_cert
			ck_private_key

It is suggested to keep all Auklet configurations in one place, such as
`~/.auklet/`. 

## broker

A newline-delimited list of Kafka broker addresses (one address per line)

## event

The Kafka topic that the wrapper sends event data to.

## prof

The Kafka topic that the wrapper sends profile data to.

## url

The base URL to be used when creating and checking releases (used by both
wrap and release commands).

## cert/

A directory containing SSL certs.

# Assign a Configuration

Point the profiler to the desired configuration directory with the environment variable
`AUKLET_ENDPOINT`. For convenience, your project folder could contain a file
called `auklet.conf` that you source to update the configuration.

	$ cat auklet.conf
	AUKLET_ENDPOINT=$HOME/.auklet/staging
	AUKLET_APP_ID=<...>
	AUKLET_API_KEY=<...>

Update the configuration:

	. auklet.conf

# Release an App

	release x

# Run an App

	wrap ./x

