# Auklet Profiler

Auklet is an IoT C/C++ profiler that runs on Linux.

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

`wrap` and `release` need at least Go 1.8. See the [getting started page][1] to
download Go. Then see [How to Write Go Code - Organization][2] to set up your
system.

[1]: https://golang.org/doc/install
[2]: https://golang.org/doc/code.html#Organization

Conventionally, your `~/.profile` should contain the following:

	export GOPATH=$HOME/go
	export PATH=$PATH:$GOPATH/bin

The first line tells Go where your workspace is located. The second makes sure
that the shell will know about executables built with `go install`.

`wrap` needs [package sarama][3], a Kafka client. Install it with

[3]: https://github.com/Shopify/sarama

	go get github.com/Shopify/sarama

# Build

To build and install all components, run

	make

In particular, this installs the commands `wrap` and `release` to `$GOPATH/bin`,
and the static library `libauklet.a` to `/usr/local/lib/`.

It also builds test executables `x` and `x-dbg`.

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

A plaintext file containing a newline-delimited list of Kafka broker addresses
(one address per line). For example:

	broker1
	broker2
	broker3

## event

A plaintext file containing the Kafka topic that the wrapper sends event data to.

## prof

A plaintext file containing the Kafka topic that the wrapper sends profile data to.

## url

A plaintext file containing the base URL, without a trailing slash, to be used
when creating and checking releases. It is accessed by both `wrap` and `release`
commands.  For example:

	https://api-staging.auklet.io/v1

## cert/

A directory containing SSL certs that `wrap` uses to authenticate its Kafka
connection. These must be [PEM][1] files, which must end in a blank line (double
newline). For example:

[1]: https://en.wikipedia.org/wiki/Privacy-enhanced_Electronic_Mail

	-----BEGIN CERTIFICATE-----
	MIICLDCCAdKgAwIBAgIBADAKBggqhkjOPQQDAjB9MQswCQYDVQQGEwJCRTEPMA0G
	A1UEChMGR251VExTMSUwIwYDVQQLExxHbnVUTFMgY2VydGlmaWNhdGUgYXV0aG9y
	aXR5MQ8wDQYDVQQIEwZMZXV2ZW4xJTAjBgNVBAMTHEdudVRMUyBjZXJ0aWZpY2F0
	ZSBhdXRob3JpdHkwHhcNMTEwNTIzMjAzODIxWhcNMTIxMjIyMDc0MTUxWjB9MQsw
	CQYDVQQGEwJCRTEPMA0GA1UEChMGR251VExTMSUwIwYDVQQLExxHbnVUTFMgY2Vy
	dGlmaWNhdGUgYXV0aG9yaXR5MQ8wDQYDVQQIEwZMZXV2ZW4xJTAjBgNVBAMTHEdu
	dVRMUyBjZXJ0aWZpY2F0ZSBhdXRob3JpdHkwWTATBgcqhkjOPQIBBggqhkjOPQMB
	BwNCAARS2I0jiuNn14Y2sSALCX3IybqiIJUvxUpj+oNfzngvj/Niyv2394BWnW4X
	uQ4RTEiywK87WRcWMGgJB5kX/t2no0MwQTAPBgNVHRMBAf8EBTADAQH/MA8GA1Ud
	DwEB/wQFAwMHBgAwHQYDVR0OBBYEFPC0gf6YEr+1KLlkQAPLzB9mTigDMAoGCCqG
	SM49BAMCA0gAMEUCIDGuwD1KPyG+hRf88MeyMQcqOFZD0TbVleF+UsAGQ4enAiEA
	l4wOuDwKQa+upc8GftXE2C//4mKANBC6It01gUaTIpo=
	-----END CERTIFICATE-----
	

# Assign a Configuration

The profiler reads the path defined in the environment variable
`AUKLET_ENDPOINT` to determine which configuration to use. For convenience,
your project folder could contain a file called `auklet.conf` that you source to
update the configuration, app ID, and API key all at once.

	$ cat auklet.conf
	AUKLET_ENDPOINT=$HOME/.auklet/staging
	AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	AUKLET_API_KEY=SM49BAMCA0...

Update the configuration:

	. auklet.conf

# Release an App

To release an executable called `x`, create an executable in the same directory
called `x-dbg` that contains debug information. (`x` is not required to contain
debug info.) Then run

	release x

# Run an App

	wrap ./x

