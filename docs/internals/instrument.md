# Instrument

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
names and source code locations.^[A basic callgraph generator and symbolizer
script can be found [here][3]]

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
