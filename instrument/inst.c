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
static int dumpstack(char *samp);
static void printNode(int indent, Node *n);
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

static Node *tree;
static Node *tp;

static Frame *stack;
static int cap, len;

static int sock;

/* function definitions */
__attribute__ ((constructor)) void
init(void)
{
	struct sockaddr_un remote;
	struct itimerval new = {
		.it_interval = {0, 10000},
		.it_value = {0, 10000},
	}, old;
	int length;

	/* initialize the tree */
	tree = (Node *)malloc(sizeof(Node));
	if (!tree) {
		printf("init: tree allocation failed\n");
		exit(1);
	}

	tree = newNode(NULL, (Frame){0, 0});
	tp = tree;

	/* initialize the stack */
	stack = (Frame *)malloc(10 * sizeof(Frame));
	if (stack) {
		cap = 10;
		len = 0;
	} else {
		printf("init: stack allocation failed\n");
		exit(1);
	}

	/* Enable instrumentation and sampling. */
	enter_action = push;
	exit_action = pop;
	setitimer(ITIMER_PROF, &new, &old);
	//signal(SIGPROF, sample);

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
	int length = dumpstack(samp);

	/* Send a stack sample over the socket. */
	if (send(sock, samp, length, 0) == -1) {
		/* Could not send a sample. No big deal. */
	}
	printNode(0, tree);
	printf("\n");
}

static int
dumpstack(char *buf)
{
	int wc = 0;
	for (int i = 0; i < len; ++i) {
		wc += sprintf(buf + wc, "%lx:%lx ",
		              (unsigned long)stack[i].fn,
		              (unsigned long)stack[i].cs);
	}
	wc += sprintf(buf + wc, "\n");
	return wc;
}

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
	//printf("push: top\n");

	Node *callee = hasCallee(tp, f);
	Frame *new = NULL;
	
	/* Increment tp to the current node. */
	tp = callee;

	if (cap == len) {
		/* The stack is full; we need to grow it. */
		new = (Frame *)realloc(stack, (cap + 1) * sizeof(Frame));
		if (new) {
			stack = new;
			++cap;
		} else {
			printf("push: realloc failed\n");
			exit(1);
		}
	}
	
	++len;
	stack[len - 1] = f;
	//printf("push: bottom\n");
}

static Node *
hasCallee(Node *n, Frame f)
{
	//printf("hasCallee: n = %p, slice %d/%d %p\n", n, n->len, n->cap, n->callee);
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
	//printf("pop: top:\n");

	++tp->ncalls;
	if (!tp->parent) {
		printf("pop: called with null parent\n");
		exit(1);
	}

	tp = tp->parent;

	/* Don't bother resizing the stack, because we're going to need it
	 * later. We currently do not free the stack because of a possible race
	 * between our destructor and the instrumentee's destructors. */
	if (len > 0)
		--len;
	else {
		/* stack corruption! */
		printf("pop: called with stack len <= 0");
		exit(1);
	}
	//printf("pop: bottom:\n");
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
