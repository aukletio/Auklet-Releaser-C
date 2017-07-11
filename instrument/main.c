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
	hello();
	return 0;
}
