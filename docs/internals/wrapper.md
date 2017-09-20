# Wrapper

The wrapper manages the execution of your instrumented program.^[See [this
discussion][1]] It verifies that the backend is prepared to receive callgraph
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
