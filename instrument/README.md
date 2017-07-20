# APM Instrument

A C library that your program is linked against at compile time.

## Linking a C/C++ program against the instrument library

Build the instrument by running

	make

Then copy `instrument.o` to the build directory of the C/C++ program you want to
profile.

	cp instrument.o /your/build/directory

Use the following toolchain arguments when compiling your program. For instance,
if using a Makefile, you might add something like

	CFLAGS += -finstrument-functions
	OBJ += instrument.o

When in doubt, look at `Makefile`, which builds a simple C program.

## How to Test

Build the wrapper and instrument, then run

	cd instrument
	./test.sh
