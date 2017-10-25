/*
 * unit tester for profiler library
 */

#include "lib.c"
#define len(x) (sizeof(x)/sizeof(x[0]))

F z = {0, 0};
F f = {0xaced, 0xfade};

int
callee_test(void)
{
	N *root = newN(z);
	N *c = addcallee(root, f);
	//dumpN(root, 0);
	N *g = hascallee(root, f);
	int ret = 1;
	if (c != g) {
		printf("callee_test: expected %p, got %p\n", (void *)c, (void *)g);
		ret = 0;
	}
	killN(root, 0);
	return ret;
}

int
marshal_test(void)
{
	N *root = newN(z);
	B b = {0, 0, 0};
	char *e = "{}";
	int ret = 1;
	if (!marshal(&b, root)) {
		printf("marshal_test: marshal failed\n");
		ret = 0;
	}
	if (strcmp(b.buf, e)) {
		printf("marshal_test: expected \"%s\", got \"%s\"\n", e, b.buf);
		ret = 0;
	}
	killN(root, 0);
	free(b.buf);
	return ret;
}

int
main()
{
	int (*test[])(void) = {
		callee_test,
		marshal_test,
	};
	for (int i = 0; i < len(test); ++i)
		if (!test[i]())
			exit(1);
	return 0;
}
