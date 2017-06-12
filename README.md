# APM Profiler Library

## Build

	make

## Use in a C/C++ program

	cp profiler.a /your/build/directory

Use these arguments when compiling. For instance, if using a Makefile

	CFLAGS += -finstrument-functions
	LDLIBS += -pthread
	OBJ += profiler.a

When in doubt, look at the provided Makefile, which builds a simple C program.

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

	grep Tick profiler.go


