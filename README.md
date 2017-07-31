# APM Profiler

The profiler consists of three components:

- instrument: a C library that your program is linked against at
  compile time
- releaser: a deploy-time command-line tool that stores symbol information
  from your program in the backend
- wrapper: a command-line program that executes and communicates with your
  program in production

See the component directories for specific information.

# How to Build

	cd wrapper
	go build

	cd ../releaser
	go build
	cd test
	make

	cd ../../instrument
	make

# How to Test

	cd releaser
	./test.sh

	cd ../instrument
	./test.sh
