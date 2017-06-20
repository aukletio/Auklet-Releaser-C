#include <stdio.h>
#include <unistd.h>

int f(int n)
{
	return n + 1;
}

int
main(int argc, char **argv)
{
	int x;
	for (int i = 0; i < 10; i++) {
		x = f(i);
		printf("%d\n", x);
		sleep(1);
	}
	return 0;
}
