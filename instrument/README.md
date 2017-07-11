# APM Instrument

A C library that your program is linked against at compile time.

## Linking a C/C++ program against the instrument library

Build the instrument

	cd instrument
	make

Then copy `instrument.a` to the build directory of the C/C++ program you want to
profile.

	cp instrument.a /your/build/directory

Use the following toolchain arguments when compiling your program. For instance,
if using a Makefile, you might add something like

	CFLAGS += -finstrument-functions -g
	LDLIBS += -pthread
	OBJ += instrument.a

When in doubt, look at `instrument/Makefile`, which builds a simple C program.
