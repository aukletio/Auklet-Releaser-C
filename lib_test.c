/*
 * unit tester for profiler library
 */

#include "lib.c"

#define str(x) #x

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
		printf("%s: expected %p, got %p\n", __func__, (void *)c, (void *)g);
		ret = 0;
	}
	killN(root, 0);
	return ret;
}

int
marshal_test(void)
{
	int err;
	B b = {0, 0, 0};
	char *e = "{}";
	int ret = 1;
	err = setjmp(nomem);
	if (err) {
		printf("%s: marshal failed\n", __func__);
		ret = 0;
		goto end;
	}
	N *root = newN(z);
	marshal(&b, root);
	if (strcmp(b.buf, e)) {
		printf("%s: expected \"%s\", got \"%s\"\n", __func__, e, b.buf);
		ret = 0;
	}
end:
	killN(root, 0);
	free(b.buf);
	return ret;
}

int
marshal_test2(void)
{
	int err;
	B b = {0, 0, 0};
	char *e = "{\"callees\":[{"
		"\"fn\":44269,"
		"\"cs\":64222,"
		"\"ncalls\":1"
	"}]}";
	int ret = 1;

	err = setjmp(nomem);
	if (err) {
		printf("%s: marshal failed\n", __func__);
		ret = 0;
		goto end;
	}
	N *root = newN(z);
	N *sp = root;
	push(&sp, f);
	//addcallee(root, f);
	marshal(&b, root);
	if (strcmp(b.buf, e)) {
		printf("%s:\n"
		       "expected \"%s\"\n"
		       "got      \"%s\"\n", __func__, e, b.buf);
		ret = 0;
	}
end:
	killN(root, 0);
	free(b.buf);
	return ret;
}

int
sample_test(void)
{
	N *root = newN(z);
	int ret = 1;
	N *sp = addcallee(addcallee(root, f), f);
	sample(sp);
	ret = sane(root);
	killN(root, 0);
	return ret;
}

int
marshals_test(void)
{
	int err;
	B b = {0, 0, 0};
	char *e = "{\"signal\":11,\"stack_trace\":[{"
		"\"fn\":44269,"
		"\"cs\":64222"
	"},{"
		"\"fn\":44269,"
		"\"cs\":64222"
	"},{"
		"\"fn\":0,"
		"\"cs\":0"
	"}]}";
	int ret = 1;

	err = setjmp(nomem);
	if (err) {
		printf("%s: marshals failed\n", __func__);
		ret = 0;
		goto end;
	}
	N *root = newN(z);
	N *sp = addcallee(addcallee(root, f), f);
	marshals(&b, sp, 11);
	if (strcmp(b.buf, e)) {
		printf("%s:\n"
		       "    expected \"%s\"\n"
		       "    got      \"%s\"\n", __func__, e, b.buf);
		ret = 0;
	}
end:
	killN(root, 0);
	free(b.buf);
	return ret;
}

int
main()
{
	struct {
		int (*run)(void);
		char *name;
	} test[] = {
#define TEST(f) {f, str(f)}
		TEST(callee_test),
		TEST(marshal_test),
		TEST(marshal_test2),
		TEST(marshals_test),
		TEST(sample_test),
	};
	int ret = 0;
	for (int i = 0; i < len(test); ++i) {
		int pass = test[i].run();
		printf("%d/%lu %15s: %s\n", i+1, len(test),
		       test[i].name,
		       pass ? "pass" : "fail");
		if (!pass)
			ret = 1;
	}
	return ret;
}