# Auklet C/C++ Profiler

Auklet's IoT C/C++ profiler is built to run on any POSIX operating system. It
has been validated on the following systems:

Board                         | Distribution  | Architecture
------------------------------|---------------|-------------
N/A                           | Ubuntu 16.04  | x86-64
ARTIK 530                     | Fedora 24     | ARM7
ARTIK 710                     | Fedora 24     | ARM7
BeagleBone Black Wireless     | Debian 8.6    | ARM7
ConnectCore 6 i.MX6Quad       | Yocto 2.2-r2  | ARM7
DragonBoard 410c              | Debian 8.6    | ARM64
Seeeduino Cloud (Arduino Yún) | OpenWRT 3.8.3 | MIPS

It consists of three components:

## `libauklet.a`

A C library that your program is linked against at compile time.

## `release`

A deploy-time command-line tool that sends symbol information from the profiled
program to the backend.

## `wrap`

A command-line program that runs the program to be profiled and continuously
sends live profile data to the backend.

# TODO: Install

# Configure

An Auklet configuration is defined by the following environment variables.

	AUKLET_APP_ID
	AUKLET_API_KEY
	AUKLET_BROKERS
	AUKLET_PROF_TOPIC
	AUKLET_EVENT_TOPIC
	AUKLET_CA
	AUKLET_CERT
	AUKLET_PRIVATE_KEY

To view your current configuration, run `env | grep AUKLET`.

To make it easier to manage multiple configurations, it is suggested to define
the envars in a shell script named after the configuration; for example,
`.env`.

The variables `AUKLET_API_KEY` and `AUKLET_APP_ID` are likely to be different
among developers, so it is suggested that they be defined in a separate
file, `.auklet`, and sourced from within `.env`. For example:

	$ cat .auklet
	export AUKLET_APP_ID=5171dbff-c0ea-98ee-e70e-dd0af1f9fcdf
	export AUKLET_API_KEY=SM49BAMCA0...

	$ cat .env
	. .auklet
	export AUKLET_BROKERS=broker1,broker2,broker3
	export AUKLET_PROF_TOPIC=z8u1-profiler
	export AUKLET_EVENT_TOPIC=z8u1-events
	export AUKLET_CA=LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS...
	export AUKLET_CERT=LS0tLS1CRUdJTiBDRVJUSUZJQ0FU...
	export AUKLET_PRIVATE_KEY=LS0tLS1CRUdJTiBQUklW...

## `AUKLET_BROKERS`

A comma-delimited list of broker addresses. For example:

	broker1,broker2,broker3

## `AUKLET_EVENT_TOPIC`, `AUKLET_PROF_TOPIC`

Topics to which `wrap` should send event and profile data, respectively.

## `AUKLET_CA`, `AUKLET_CERT`, `AUKLET_PRIVATE_KEY`

Base64-encoded [PEM][pem]-format certs that `wrap` uses to authenticate its
connection.

[pem]: https://en.wikipedia.org/wiki/Privacy-enhanced_Electronic_Mail

# Assign a Configuration

	. .env

# Integrate with Your App

When compiling source files to object files, pass the flags
`-finstrument-functions -g` to your compiler.  When linking object files into an
executable, pass the flag `-libauklet` to your compiler.

# Run an App

	wrap ./x

