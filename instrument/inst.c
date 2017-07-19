/* headers */
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/un.h>

#define SOCK_PATH "socket"

/* types */
typedef struct {
	void *fn, *cs;
} Frame;

typedef struct node {
	Frame frame;
	unsigned ncalls;
	struct node **callee, *parent;
	int cap, len;
} Node;

/* function declarations */
__attribute__ ((constructor)) void init(void);
static Node *newNode(Node *parent, Frame f);
static void sample(int n);
static void zeroNode(Node *n);
static int dumpstack(Node *n, char *buf);
static int printNode(int indent, Node *n);
static void push(Frame);
static Node *hasCallee(Node *n, Frame f);
static Node *appendCallee(Node *n, Frame f);
static void pop(Frame);
static void nop(Frame);
void __cyg_profile_func_enter(void *fn, void *cs);
void __cyg_profile_func_exit(void *fn, void *cs);

/* global variables */
static void (*enter_action)(Frame f) = nop;
static void (*exit_action)(Frame f) = nop;

static Node *tree, *tp;

static int sock;

/* macros */
#define SAMPLE_PD 10000

/* function definitions */
__attribute__ ((constructor)) void
init(void)
{
	struct sockaddr_un remote;
	struct itimerval new = {
		.it_interval = {0, SAMPLE_PD},
		.it_value = {0, SAMPLE_PD},
	}, old;
	int length;

	/* initialize the tree */
	tree = newNode(NULL, (Frame){0, 0});
	tp = tree;

	/* Enable instrumentation. */
	enter_action = push;
	exit_action = pop;

	if ((sock = socket(AF_UNIX, SOCK_STREAM, 0)) == -1) {
		printf("init: no socket\n");
		return;
	}

	remote.sun_family = AF_UNIX;
	strcpy(remote.sun_path, SOCK_PATH);
	length = strlen(remote.sun_path) + sizeof(remote.sun_family);
	if (connect(sock, (struct sockaddr *)&remote, length) == -1) {
		printf("init: no connection\n");
		return;
	}

	/* Turn on the sampler only if we're connected. This is useful when
	 * profiling the instrument with gprof, which must be able to install
	 * its own SIGPROF handler. */
	setitimer(ITIMER_PROF, &new, &old);
	signal(SIGPROF, sample);
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
	new->parent = parent;

	/* Callee lists are allocated lazily, when we attempt to append to
	 * them. */
	new->callee = NULL;
	new->cap = 0;
	new->len = 0;

	return new;
}

void
sample(int n)
{
	char samp[256];
	int length = dumpstack(tp, samp);
	length += sprintf(samp + length, "\r\n");

	/* Send a stack sample over the socket. */
	if (send(sock, samp, length, 0) == -1) {
		/* Could not send a sample. No big deal. */
	}
}

static int
dumpstack(Node *n, char *buf)
{
	int wc = 0;

	/* If the current node is not an (immediate) child of the root of the
	 * tree, dump its parent first. This both avoids dumping the meaningless
	 * tree root and formats the sample so that the stack grows rightward.
	 */

	if (n->parent != tree) {
		wc += dumpstack(n->parent, buf);
	}

	/* All values are hex encoded to save a little space. */
	wc += sprintf(buf + wc, "%lx:%lx:%x ",
	              (unsigned long)n->frame.fn,
	              (unsigned long)n->frame.cs,
	              n->ncalls);

	/* It's convenient, but hacky, to clear the counters here, while we're
	 * walking the tree. Counters are cleared only after samples are taken.
	 * Only counters which get sampled get cleared; some counters will
	 * rarely be sampled and therefore rarely cleared, but that should not
	 * be a problem. If there is risk of overflow, `unsigned long`s could
	 * be used. */
	n->ncalls = 0;

	return wc;
}

/* printNode is useful when debugging; it prints a Node and all of its children
 * to stdout in a human-readable way. */
static void
printNode(int indent, Node *n)
{
	if (n == tp)
		printf("→");
	else
		printf(" ");

	printf("%8d ", n->ncalls);
	for (int i = 0; i < indent; ++i)
		printf(".   ");

	printf("%p:%p\n", n->frame.fn, n->frame.cs);

	for (int i = 0; i < n->len; ++i)
		printNode(1 + indent, n->callee[i]);
}

static void
push(Frame f)
{
	Node *callee = hasCallee(tp, f);
	Frame *new = NULL;
	
	/* Move tp to the current node. */
	tp = callee;
}

static Node *
hasCallee(Node *n, Frame f)
{
	Node *callee = NULL;
	/* Search the callee list for the given frame. */
	for (int i = 0; i < n->len; ++i) {
		if (n->callee[i]->frame.cs == f.cs) {
			callee = n->callee[i];
			break;
		}
	}

	if (!callee) {
		/* This is the first time visiting this part of the program; we
		 * need to grow the list of callees at the current location. */
		callee = appendCallee(tp, f);
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

		n->callee = new;
		++n->cap;
	}

	++n->len;
	node = newNode(n, f);
	n->callee[n->len - 1] = node;
	return node;
}

static void
pop(Frame f)
{
	++tp->ncalls;
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
