#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>

__attribute__ ((destructor)) void
init(void)
{
	printf("I'm a destructor.\n");
}

int
fib(int n)
{
	if (n < 2) {
		sleep(1);
		return n;
	} else
		return fib(n - 1) + fib(n - 2);
}

void
g(void)
{
	printf("g = %ld\n", g);
	sleep(1);
}

void
f(void)
{
	printf("f = %ld\n", f);
	g();
}

int
main(int argc, char **argv)
{
	printf("main = %ld\n", main);
	int val = 0;
	if (argc > 1)
		val = atoi(argv[1]);
	for (int i = 0; i < val; i++)
		f();
	return 0;
}
