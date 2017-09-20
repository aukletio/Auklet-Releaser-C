# Installing

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
