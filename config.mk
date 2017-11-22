# flags
CFLAGS = -pedantic -std=c99 -D_POSIX_C_SOURCE=200809L -D_GNU_SOURCE

# profiler runtime installation path
INSTALL = /usr/local/lib

# for compiling with the profiler runtime
PFLAGS = -g -finstrument-functions
PLIBS = -lpthread

# overridden by CircleCI to use govvv
GOCMD ?= go
