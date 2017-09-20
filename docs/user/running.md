# Running

Usage is

	wrapper -n executable arg1 arg2 ...

## Caveats

`stdin` is currently not passed through to your program.

The wrapper assumes that it is the parent process of your program, and the
instrument assumes that it is in the child process of the wrapper. If you need
to execute your program from a script, invoke the wrapper in the script:

	#!/bin/sh

	ENVAR='some important setting'

	wrapper -n executable arg1 arg2

The wrapper assumes that it has permission to create files in the current
working directory.

The wrapper assumes that it has Internet access.
