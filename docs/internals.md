# Internals

## Instrument

The link-time library component of Auklet is called _the instrument,_ because it
is responsible for instrumenting your program and generating profile data during
its execution.

The basic mechanism for this is the compile-time GCC-compatible flag,
`-finstrument-functions` (see [here][1]), which injects a pair of function calls
into every function in the current compilation unit. These are used as "hooks"
that enable the instrument to track your program's stack.

[1]: https://gcc.gnu.org/onlinedocs/gcc-4.3.3/gcc/Code-Gen-Options.html

The goal of the instrument is to generate a [callgraph][2], which is like a
roadmap of your program's execution. In Auklet, a callgraph is a tree in
which each node corresponds to a particular function. Auklet's callgraph is more
detailed than traditional directed-acyclic-graph (DAG) callgraphs, because it
preserves callsite and stacktrace information. This means that a node in the
callgraph is not necessarily unique.  There may be multiple nodes with the same
function address and callsite, if they differ in their stacktrace.

[2]: https://en.wikipedia.org/wiki/Call_graph

Callgraphs refer to functions by their addresses in the virtual memory space
provided by the kernel. These addresses can later be "symbolized" with debugging
and symbol information, which associates the function addresses with function
names and source code locations. A basic callgraph generator and symbolizer
script can be found [here][3].

[3]: https://git.2f30.org/callgraph/files.html

Conventional profilers, such as gprof and pprof, wait until your program
terminates to write callgraph data to a file. Since Auklet is supposed to run in
production and provide constantly-updating metrics, the instrument generates
callgraphs at periodic intervals in addition to the moment of termination.
Instead of writing it to a file, it is sent over a Unix domain socket to the
wrapper instance that executed it. The wrapper does some processing and sends it
to the backend.

Another goal of the instrument is to avoid disturbing the execution of your
program. This means that in the event of an unrecoverable error, the instrument
should fall back to a dormant state rather than terminate your program. It does
this by using function pointers that are switched to empty functions after an
error happens. After shutdown, some of the instrument's functions are still
called, but do nothing and return immediately.

It's still possible that the instrument could cause your program to terminate or
segfault, but we try hard to avoid it.

## Releaser

The releaser is a command-line tool that you use to inform the Auklet backend
that you want to start profiling a particular release. You run this tool after
you have compiled your code. The releaser allows the backend to make sense of
your program's profile data.

Conventional profilers symbolize callgraphs either when writing them to a file
or when reading them through some kind of presentation interface. They do this
by reading DWARF entries and ELF symbols embedded in the executable. Since this
information expands the size of the executable, it is common to save space on
target devices by stripping this information out.

The releaser requires that you provide an executable that has debugging symbols.
It allows your deployable executable to be stripped, but compares the two
exectuables to make sure that they have the same instructions. If everything
goes well, then the releaser copies debug and symbol information and sends it to
Auklet's backend, where it will be used to symbolize the callgraphs for the user
interface.

The releaser also computes a checksum of your deployable executable, so that the
backend can identify incoming callgraph data and associate it with the correct
release.

## Wrapper

The wrapper manages the execution of your instrumented program (see [this
discussion][1]). It verifies that the backend is prepared to receive callgraph
data from your program by computing a checksum of your executable and asking the
backend if a release exists for that checksum. If so, it will send profile data
to the backend; otherwise, not. In either case your program should be executed,
since the absence of a release shouldn't prevent you from running it.

[1]: https://groups.google.com/d/msg/golang-nuts/qBQ0bK2zvQA/W-GQviEvVSUJ

The wrapper creates a Unix domain socket and binds it to a local address that
uniquely identifies the wrapper instance. The instrument uses this socket to
send JSON-formatted callgraphs to the wrapper.

The wrapper adds your executable's checksum to each callgraph so that the
backend can associate the callgraph with a particular release of your program.
