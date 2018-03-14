/*
The releaser is a command-line tool that is intended to be run in a CI
environment after an app is compiled into an executable file. It extracts symbol
information from the executable and sends it to Auklet's backend.

Conventional profilers symbolize callgraphs either when writing them to a file
or when reading them through some kind of presentation interface. They do this
by reading DWARF entries and ELF symbols embedded in the executable. Since this
information expands the size of the executable, it is common to save space on
target devices by stripping this information out.

The releaser requires that users provide an executable that has debugging symbols.
It allows a deployable executable to be stripped, but compares the two
exectuables to make sure that they have the same instructions. If everything
matches, then the releaser copies debug and symbol information and sends it to
Auklet's backend, where it will be used to symbolize the callgraphs for the user
interface.

The releaser also computes a checksum of the executable to be deployed, so that
the backend can identify incoming callgraph data and associate it with the
correct release.
*/
package main
