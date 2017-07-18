#include <stdio.h>

__attribute__ ((destructor)) void
cleanup(void)
{
	printf("I'm a destructor.\n");
}

void
hello(void)
{
	printf("%p: hello, world\n", hello);
}

void
goodbye(void)
{
	printf("%p: goodbye, world\n", goodbye);
}

int
main(int argc, char **argv)
{
	printf("%p: starting up\n", main);
	hello();
	goodbye();
	return 0;
}
