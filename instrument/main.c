#include <stdio.h>

__attribute__ ((destructor)) void
cleanup(void)
{
	printf("I'm a destructor.\n");
}

void
hello(void)
{
	printf("hello, world\n");
}

int
main(int argc, char **argv)
{
	for (int i = 0; i < 1000000000; i++)
		if (!(i % 10000))
			hello();
	return 0;
}
