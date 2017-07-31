/* Unit tests for inst.c */
#include "inst.c"

#define len(x) (sizeof(x)/sizeof(x[0]))
typedef struct {
	Node *n;
	String *s;
} Marshal;

/* A test case is implemented by writing a function that sets up input-output
 * pairs. Every test case has the same function signature, so all cases can be
 * iterated over in a loop.
 */
void
case_empty_marshal(Marshal *m)
{
	m->n = newNode(NULL, (Frame){0, 0});

	m->s = newString(100);
	sappend(m->s, "{}");
}

void
case_simple_marshal(Marshal *m)
{
	m->n = newNode(NULL, (Frame){0, 0});
	appendCallee(m->n, (Frame){0, 0});
	m->n->callee[0]->ncalls = 1;
	m->n->callee[0]->empty = 0;

	m->s = newString(100);
	sappend(m->s, "{\"callees\":[{\"ncalls\":1}]}");
}

void
case_more_marshal(Marshal *m)
{
	m->n = newNode(NULL, (Frame){0, 0});

	appendCallee(m->n, (Frame){0, 0});
	m->n->callee[0]->ncalls = 1;
	m->n->callee[0]->empty = 0;

	appendCallee(m->n, (Frame){(void *)1, 0});
	m->n->callee[1]->ncalls = 1;
	m->n->callee[1]->empty = 0;

	m->s = newString(100);
	sappend(m->s, "{\"callees\":[{\"ncalls\":1},{\"fn\":1,\"ncalls\":1}]}");
}

int
test_marshal(Marshal m)
{
	int result;
	String *out = newString(100);
	marshal(m.n, out);
	switch (strncmp(m.s->buf, out->buf, 100)) {
	case 0:
		/* pass */
		result = 1;
		break;
	default:
		/* fail */
		result = 0;
		printf("test_marshal: wanted %s\n"
		       "              got    %s\n", m.s->buf, out->buf);
		break;
	}
	return result;
}

int
main(void)
{
	int result = 0;
	void (*make_case[])(Marshal *) = {
		case_empty_marshal,
		case_simple_marshal,
		case_more_marshal,
	};
	Marshal m[len(make_case)];

	for (int i = 0; i < len(make_case); ++i) {
		make_case[i](&m[i]);
		if (!test_marshal(m[i])) {
			printf("case %d failed\n", i);
			result = 1;
		}
	}
	return result;
}
