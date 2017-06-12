#include <stdlib.h>
#include <unistd.h>

void f(void)
{
	sleep(1);
}

void g(void)
{
	f();
	sleep(1);
	f();
}

int
main(int argc, char **argv)
{
	int val = 0;
	sleep(1);
	g();
	return 0;
}
