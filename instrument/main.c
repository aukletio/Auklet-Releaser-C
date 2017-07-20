/* headers */
#include <stdlib.h>
#include <time.h>

/* global variables */
int salt1 = 0, salt2 = 0;
int ns;

/* function definitions */
long
now(void)
{
	struct timespec t;
	clock_gettime(CLOCK_REALTIME, &t);
	return t.tv_sec * 1000000000 + t.tv_nsec;
}

void
cpuHogger(void (*f)(void), long dur)
{
	long t0 = now();
	for (int i = 0; i < 500 || now() - t0 < dur; ++i)
		f();
}

void
cpuHog1(void)
{
	int foo = salt1;
	for (int i = 0; i < 10000; ++i) {
		if (foo > 0) {
			foo *= foo;
		} else {
			foo *= foo + 2;
		}
	}
	salt1 = foo;
}

void
cpuHog2(void)
{
	int foo = salt2;
	for (int i = 0; i < 10000; ++i) {
		if (foo > 0) {
			foo *= foo;
		} else {
			foo *= foo + 2;
		}
	}
	salt2 = foo;
}

__attribute__ ((destructor)) void
cleanup(void)
{
	cpuHogger(cpuHog2, ns);
}

int
main(int argc, char **argv)
{
	ns = 0;
	if (argc > 1)
		ns = atoi(argv[1]);
	cpuHogger(cpuHog1, ns);
	return 0;
}
