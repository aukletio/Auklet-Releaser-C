# Auklet Wrapper

A command-line program that executes and communicates with your program in
production.

## Building the Wrapper

	go install

## Running the Profiler

First install the instrument and link against your program. Then run

	wrapper [flags] your-program args ...

## Command-line flags

With no flags, the wrapper profiles your program and periodically prints
callgraphs to stdout.

- `-q`: Generate profiles, but do not print them to stdout.
- `-n`: Connect to the backend and publish profiles to it.
- `-p`: Generate a CPU profile of the wrapper.

## Output

A profile consists of a tree with the following structure:

- Fn: function address; identical to a function pointer
- Cs: callsite address; location of the `call` instruction
- Ncalls: how many times this callsite was executed
- Time: total time in nanoseconds spent at this callsite
- Callees: a list of profiles for each callee of this callsite

The tree is constructed at runtime; parts of your code that do not execute will
not show up in the tree.

The root node has no meaningful profile information aside from the list of
callees. Not all functions are called within main; your program may have
constructor-  or destructor-attributed functions that run outside of main.

### Examples

Consider the following C program:

	#include <stdio.h>

	void
	hello(void)
	{
		printf("hello, world\n");
	}

	int
	main()
	{
		hello();
		return 0;
	}

Its corresponding profile will look like this:

	{
	    "Callees": [
		{
		    "Fn": 4198857,
		    "Cs": 140340547131440,
		    "Ncalls": 1,
		    "Time": 139896,
		    "Callees": [
			{
			    "Fn": 4198806,
			    "Cs": 4198895,
			    "Ncalls": 1,
			    "Time": 18101
			}
		    ]
		}
	    ]
	}

If `hello()` were called in a loop,

	while (condition) {
		hello();
	}

the tree would have the same structure---one leaf---as the callgraph above.

If `hello()` were called twice,

	hello();
	hello();

then the tree would have two leaves, one for each callsite of the function
`hello()`.

## Notes

- A program linked against the instrument can still be run by itself without the
  wrapper. The instrumentation will still be present, but won't do anything
  (aside from slow your program down).
