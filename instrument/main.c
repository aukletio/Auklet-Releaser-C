#include <stdio.h>
#include <stdlib.h>

int
fib(int n)
{
	if (n < 2)
		return n;
	else
		return fib(n - 1) + fib(n - 2);
}

int
main(int argc, char **argv)
{
	int val = 0;
	if (argc > 1)
		val = atoi(argv[1]);
	fib(val);
	return 0;
}
