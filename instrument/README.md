# Auklet Instrument

A C library that your program is linked against at compile time.

## Linking a C/C++ program against the instrument library

Build and install the instrument by running

	make

Use the following toolchain arguments when compiling your program. For instance,
if using a Makefile, you might add something like

	CFLAGS += -finstrument-functions
	LDLIBS += -lauklet

If using CMake, your CMakeLists.txt might have

	set(CMAKE_CXX_FLAGS "-finstrument-functions")
	target_link_libararies(my_executable auklet)

See `test/` for an example project.

## How to run unit tests

	make test
