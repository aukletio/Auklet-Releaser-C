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

/* function declarations */
__attribute__ ((constructor)) void init(void);
static void sample(int n);
static int dumpstack(char *samp);
static void push(Frame);
static void pop(Frame);
static void nop(Frame);
void __cyg_profile_func_enter(void *fn, void *cs);
void __cyg_profile_func_exit(void *fn, void *cs);

/* global variables */
static void (*enter_action)(Frame f) = nop;
static void (*exit_action)(Frame f) = nop;

static Frame *stack;
static int cap, len;

static int sock;

/* function definitions */
__attribute__ ((constructor)) void
init(void)
{
	struct sockaddr_un remote;
	struct itimerval new = {
		.it_interval = {0, 100},
		.it_value = {0, 100},
	}, old;
	int length;

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
	signal(SIGPROF, sample);

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

void
sample(int n)
{
	char samp[256];
	int length = dumpstack(samp);

	/* Send a stack sample over the socket. */
	if (send(sock, samp, length, 0) == -1) {
		/* Could not send a sample. No big deal. */
	}
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
push(Frame f)
{
	Frame *new = NULL;
	
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
}

static void
pop(Frame f)
{
	/* Don't bother resizing the stack, because we're going to need it
	 * later. The stack is freed totally by fini. */
	if (len > 0)
		--len;
	else {
		/* stack corruption! */
		printf("pop: called with stack len <= 0");
		exit(1);
	}
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
