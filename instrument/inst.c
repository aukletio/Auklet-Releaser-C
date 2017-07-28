/* headers */
#include <errno.h>
#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/un.h>
#include <time.h>
#include <unistd.h>

/* macros */
#define SAMPLE_PD 10000

/* types */
typedef struct {
	void *fn, *cs;
} Frame;

typedef struct node {
	Frame frame;
	unsigned ncalls, nsamples;
	struct node **callee, *parent;
	int cap, len;

	/* A Node is empty if ncalls == nsamples == 0 and all of its callees are
	 * empty. This flag allows marshal() to skip irrelevant branches of the
	 * tree. */
	int empty;
} Node;

typedef struct {
	char *buf;
	int cap, len;
} String;

/* function declarations */
static __attribute__ ((constructor)) void init(void);
static Node *newNode(Node *parent, Frame f);
static void sample(int n);
static void emit(int n);
static void marshal(Node *n, String *s);
static String *newString(unsigned size);
static void grow(String *s);
static int sappend(String *s, const char *format, ...);
static void push(Frame);
static Node *hasCallee(Node *n, void *cs);
static Node *appendCallee(Node *n, Frame f);
static void setnotempty(Node *n);
static void pop(Frame);
static void nop(Frame);
void __cyg_profile_func_enter(void *fn, void *cs);
void __cyg_profile_func_exit(void *fn, void *cs);

/* global variables */
static void (*enter_action)(Frame f) = nop;
static void (*exit_action)(Frame f) = nop;
static Node *tree, *tp;
static int sock;

/* function definitions */
static __attribute__ ((constructor)) void
init(void)
{
	struct sockaddr_un remote;
	struct itimerval sample_pd = {
		.it_interval = {0, SAMPLE_PD},
		.it_value = {0, SAMPLE_PD},
	}, tree_pd = {
		.it_interval = {1, 0},
		.it_value = {1, 0},
	};
	struct sigaction sample_act = {
		.sa_handler = sample,
	}, tree_act = {
		.sa_handler = emit,
	};
	int length;

	tree = newNode(NULL, (Frame){0, 0});
	tp = tree;

	/* Enable instrumentation, even before we know whether we're
	 * connected. */
	enter_action = push;
	exit_action = pop;

	/* Configure timers. */
	setitimer(ITIMER_VIRTUAL, &tree_pd, NULL);
	setitimer(ITIMER_PROF, &sample_pd, NULL);

	/* Configure signal handlers (that are triggered by timer expiration). */

	/* Don't allow any signals to interrupt our signal handlers. This is
	 * done especially to prevent marshal from being interrupted and
	 * generating invalid JSON. */
	sigfillset(&sample_act.sa_mask);
	sigfillset(&tree_act.sa_mask);

	sigaction(SIGPROF, &sample_act, NULL);
	sigaction(SIGVTALRM, &tree_act, NULL);

	if ((sock = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		return;
	}

	remote.sun_family = AF_UNIX;
	sprintf(remote.sun_path, "socket-%d", getppid());
	length = strlen(remote.sun_path) + sizeof(remote.sun_family);
	if (connect(sock, (struct sockaddr *)&remote, length) == -1) {
		return;
	}
}

static Node *
newNode(Node *parent, Frame f)
{
	Node *new = (Node *)malloc(sizeof(Node));
	if (!new) {
		printf("newNode: malloc failed\n");
		exit(1);
	}

	new->frame = f;
	new->ncalls = 0;
	new->nsamples = 0;
	new->parent = parent;

	/* Callee lists are allocated lazily, when we attempt to append to
	 * them. */
	new->callee = NULL;
	new->cap = 0;
	new->len = 0;
	new->empty = 1;
	return new;
}

static void
sample(int n)
{
	++tp->nsamples;
	setnotempty(tp);
}

static void
emit(int n)
{
	static String *s = NULL;
	if (!s)
		s = newString(8000);
	marshal(tree, s);
	sappend(s, "\n");

	/* Send a tree over the socket. */
	if (send(sock, s->buf, s->len, 0) == -1) {
		/* Could not send a tree. No big deal. */
		printf("%s", s->buf);
		printf("wc = %d\n", s->len);
	}
	s->len = 0;
}

static void
marshal(Node *n, String *s)
{
	sappend(s, "{");

	if (n->frame.fn)
		sappend(s, "\"fn\":%ld,", (unsigned long)n->frame.fn);

	if (n->frame.cs)
		sappend(s, "\"cs\":%ld,", (unsigned long)n->frame.cs);

	if (n->ncalls)
		sappend(s, "\"ncalls\":%u,", n->ncalls);

	if (n->nsamples)
		sappend(s, "\"nsamples\":%u,", n->nsamples);

	/* It's convenient, but hacky, to clear the counters here while we're
	 * walking the tree. All counters are cleared after a tree is emitted,
	 * so that the next tree begins at zero.  We also set the empty flag
	 * so that future calls to marshal can skip empty branches. */
	n->ncalls = 0;
	n->nsamples = 0;
	n->empty = 1;

	if (n->len) {
		sappend(s, "\"callees\":[");
		for (int i = 0; i < n->len; ++i) {
			if (n->callee[i]->empty)
				continue;
			marshal(n->callee[i], s);
			sappend(s, ",");
		}
		if (',' == s->buf[s->len - 1])
			s->len -= 1;
		sappend(s, "]");
	} else {
		if (',' == s->buf[s->len - 1])
			s->len -= 1;
	}

	sappend(s, "}");
}

static String *
newString(unsigned size)
{
	String *s = (String *)malloc(sizeof(String));
	char *new = (char *)malloc(size * sizeof(char));
	if (!s || !new) {
		printf("newString: allocation failed\n");
		exit(1);
	}

	s->buf = new;
	s->cap = size;
	s->len = 0;
	return s;
}

static void
grow(String *s)
{
	char *new = (char *)realloc(s->buf, s->cap * 2 * sizeof(char));
	if (!new) {
		printf("grow: allocation failed\n");
		exit(1);
	}
	s->buf = new;
	s->cap *= 2;
}

static int
sappend(String *s, const char *format, ...)
{
	/* Wrap snprintf so that the underlying buffer grows if it runs out of
	 * space. */
	va_list ap;
	va_start(ap, format);
	int wc;

retry:
	wc = vsnprintf(s->buf + s->len, s->cap - s->len, format, ap);
	if (s->cap - s->len <= wc) {
		/* We ran out of buffer space. Allocate more and try again. */
		grow(s);
		goto retry;
	}

	/* Successful write. Increment the slice length to indicate the
	 * current length. */
	va_end(ap);
	s->len += wc;
	return wc;
}

static void
push(Frame f)
{
	Node *callee = hasCallee(tp, f.cs);
	if (!callee) {
		/* This is the first time visiting this part of the program; we
		 * need to grow the list of callees at the current location. */
		callee = appendCallee(tp, f);
	}

	/* Move tp to the current Node. */
	tp = callee;
}

static Node *
hasCallee(Node *n, void *cs)
{
	Node *callee = NULL;

	/* Search the callee list for the given callsite.  It suffices to check
	 * the callsite alone, since two different functions will never share a
	 * callsite. */
	for (int i = 0; i < n->len; ++i) {
		if (n->callee[i]->frame.cs == cs) {
			callee = n->callee[i];
			break;
		}
	}

	return callee;
}

static Node *
appendCallee(Node *n, Frame f)
{
	Node **new, *node;
	if (n->cap == n->len) {
		/* Grow the slice to accommodate one more reference to a callee
		 * Node. */
		if (n->cap)
			new = (Node **)realloc(n->callee, (1+n->cap)*sizeof(Node *));
		else
			new = (Node **)malloc(sizeof(Node *));

		if (!new) {
			printf("appendCallee: slice allocation failed\n");
			exit(1);
		}

		++n->cap;
		n->callee = new;
	}

	++n->len;
	node = newNode(n, f);
	n->callee[n->len - 1] = node;
	return node;
}

static void
setnotempty(Node *n)
{
	n->empty = 0;
	if (n->parent && n->parent->empty)
		setnotempty(n->parent);
}

static void
pop(Frame f)
{
	++tp->ncalls;
	setnotempty(tp);

	if (!tp->parent) {
		printf("pop: called with null parent\n");
		exit(1);
	}

	tp = tp->parent;
}

static void
nop(Frame f)
{
}

void
__cyg_profile_func_enter(void *fn, void *cs)
{
	enter_action((Frame){fn, cs});
}

void
__cyg_profile_func_exit(void *fn, void *cs)
{
	exit_action((Frame){fn, cs});
}
