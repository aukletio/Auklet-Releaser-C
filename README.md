# APM Profiler

The profiler consists of two components:

- the instrument, which is a library that your program is linked against at
  compile time
- the wrapper, which executes and communicates with your program.

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

## Building the Wrapper

	cd wrapper
	go build

## Running the Profiler

	./wrapper your-program args ...

## Functionality

Periodically, the profiler prints out JSON data that quantifies the program's
execution since the last emitted profile.

	Profile
		Timestamp for this profile
		Callsites (machine addresses where functions are called)
			Function Name
			Number of completed calls
			Total time spent at the top of the stack
			Location in source code of this callsite
			Location in source code of definition

The period can be found by

	grep Tick wrapper/wrap.go

## Notes

- A program linked against the instrument can still be run by itself without the
  wrapper. The instrumentation will still be present, but won't do anything
  (aside from slow your program down).
