# User Documentation

## Installing

The system that builds your code will need to have the instrument library,
`libauklet.a`. For convenience, it can be installed to your linker's default
search path, so it can be found automatically when building your code.

Auklet can work with GCC-like compilers, including Clang.

If in doubt as to where your linker looks, run your compiler with a `-v` flag,
which should print out the directories searched by the linker.

On Ubuntu, you can put `libauklet.a` in `/usr/local/lib`.

The system that deploys your code (probably the same as what builds it) needs to
have the releaser installed. For convenience, it should appear in `$PATH`. On
Ubuntu, you can put it in `/usr/local/bin`.

## Building

C and C++ programs are traditionally built in two stages: compilation and
linking.

### Compiling

At compile-time, pass the flags `-finstrument-functions -g` to your compiler.
The releaser needs debug information for your program; that's why the `-g` flag
is required. 

### Linking

At link-time, pass `-libauklet` to the compiler, but be aware that when using
GCC, [the order of arguments matters][1].

[1]: https://stackoverflow.com/questions/6247926/gcc-command-line-argument-pickiness

### Releasing

If you want to release a stripped executable (one without debug info), copy the
debuggable executable before running `strip`:

	cp x_debug x_stripped
	strip x_stripped

Then you can create a release.

	releaser -appid $APP_ID -apikey $API_KEY -debug x_debug -deploy x_stripped

If you want to release a debuggable executable, give `-debug` and `-deploy` the
same filename.

	releaser -appid $APP_ID -apikey $API_KEY -debug x_debug -deploy x_debug

### Troubleshooting

Verify that the `-finstrument-functions` flag is working. Build your program
with debug info (use `-g`). The command

	nm x_debug | grep __cyg_profile_func_

should show two functions.

The instrument library may not be getting linked. Try changing the argument
position of the `-libauklet` flag.

### Caveats

-   If your program forks, the child process will not be profiled.
-   Multi-threaded programs are currently unsupported. They may run without issue,
    but the callgraphs generated may be inaccurate.
-   The profiler library uses the following POSIX facilities:

    - `SIGPROF`
    - `SIGVTALRM`
    - `ITIMER_VIRTUAL`
    - `ITIMER_PROF`

    Programs that depend on these facilities will disturb the profiler.

-   If your program is linked together from multiple object (`.o`) files, it's
    possible to not instrument certain compilation units by not passing
    `-finstrument-functions` when compiling those units. This would result in
    missing profile information for the un-instrumented compilation units, which
    is likely to cause the callgraph data to be incomplete. Incompleteness
    affects the accuracy of the information, because functions which are not
    instrumented are in effect invisible, so time taken up in them will be
    attributed to callers.

## Running

### Requirements

- The wrapper assumes that it is the parent process of your program, and the
  instrument assumes that it is in the child process of the wrapper. If you need
  to execute your program from a script, invoke the wrapper in the script:

      #!/bin/sh
      ENVAR='some important setting'
      wrapper -n executable arg1 arg2

- The wrapper assumes that it has permission to create files in the current
  working directory.

- The wrapper assumes that it has Internet access.

### Execution

Usage is

	wrapper -n executable arg1 arg2 ...

### Caveats

`stdin` is currently not passed through to your program.
