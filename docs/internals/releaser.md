# Releaser

The releaser is a command-line tool that you use to inform the Auklet backend
that you want to start profiling a particular release. You run this tool after
you have compiled your code. The releaser allows the backend to make sense of
your program's profile data.

Conventional profilers symbolize callgraphs either when writing them to a file
or when reading them through some kind of presentation interface. They do this
by reading DWARF entries and ELF symbols embedded in the executable. Since this
information expands the size of the executable, it is common to save space on
target devices by stripping this information out.

The releaser requires that you provide a version of your program that has
debugging symbols. It allows the deployable binary to be stripped, but compares
the two exectuables to make sure that they have the same instructions. If
everything goes well, then the releaser copies debug and symbol information and
sends it to Auklet's backend, where it will be used to symbolize the callgraphs
for the user interface.

The releaser also computes a checksum of your deployable executable, so that the
backend can identify incoming callgraph data and associate it with the correct
release.
