/*
 * profiler runtime
 */

/* headers */
#include "lib.c"

#include <fcntl.h>
#include <signal.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/un.h>
#include <unistd.h>

/* function declarations */
static void sigprof(int n);
static void signals(void);
static void emit(void);
static void *timer(void *);
static void timers(void);
static void setup(void);
static void cleanup(void);

static void enternop(N **sp, F f) {}
static F exitnop(N **sp) {}

void __cyg_profile_func_enter(void *fn, void *cs);
void __cyg_profile_func_exit(void *fn, void *cs);

/* global variables */
static void (*instenter)(N **sp, F f) = enternop;
static F (*instexit)(N **sp) = exitnop;
static N root = {
	.f = {0, 0},
	.parent = NULL,
	.nsamp = 0,
	.llist = PTHREAD_MUTEX_INITIALIZER,
	.lsamp = PTHREAD_MUTEX_INITIALIZER,
	.lcall = PTHREAD_MUTEX_INITIALIZER,
	.callee = NULL,
	.cap = 0,
	.len = 0,
	.ncall = 0,
};
static __thread N *sp = &root;
static int sock;

/* function definitions */
/* Increment sample counters in the stack of the current thread. */
static void
sigprof(int n)
{
	sample(sp);
}

/* Send a JSON-encoded profile tree to the wrapper. */
static void
emit(void)
{
	B b = {0, 0, 0};
	//dumpN(&root, 0);
	marshal(&b, &root);
	append(&b, "\n");
	//dprintf(log, "emit: %s", b.buf);
	if (send(sock, b.buf, b.len, 0) == -1) {
		//dprintf(log, "emit: send: %s\n", strerror(errno));
		//exit(1);
	}
	reset(&root);
	free(b.buf);
}

/* Set up signal handlers. */
static void
signals(void)
{
	struct sigaction prof, emit;
	sigaction(SIGPROF, NULL, &prof);
	prof.sa_handler = sigprof;

	/* sigfillset prevents sigprof from getting interrupted, but this
	 * does not avoid races between handlers in different threads. */
	sigfillset(&prof.sa_mask);
	sigaction(SIGPROF, &prof, NULL);
}

/* Emit profile data periodically. Implementing this as a interval timer +
 * signal handler combo would interrupt--and therefore deadlock--instrumented
 * threads, which must acquire locks. To avoid this, it is a dedicated thread. */
static void *
timer(void *p)
{
	struct timespec t = {.tv_sec = 1, .tv_nsec = 0};
	sigset_t s;
	sigfillset(&s);
	pthread_sigmask(SIG_BLOCK, &s, NULL);
	while (1) {
		clock_nanosleep(CLOCK_PROCESS_CPUTIME_ID, 0, &t, NULL);
		emit();
	}
}

/* Start timers for stack sampling and profile tree emission. */
static void
timers(void)
{
#define SP {.tv_sec = 0, .tv_usec = 10000}
	struct itimerval prof = {SP, SP};
	setitimer(ITIMER_PROF, &prof, NULL);
	pthread_t t;
	pthread_create(&t, NULL, timer, NULL);
}

/* Set up a communicaiton channel with the wrapper. */
static int
comm(int type, char *prefix)
{
	struct sockaddr_un remote;
	int l, fd;
	if ((fd = socket(AF_UNIX, type, 0)) == -1) {
		return 0;
		//dprintf(log, "comm: socket: %s\n", strerror(errno));
		//exit(1);
	}
	remote.sun_family = AF_UNIX;
	sprintf(remote.sun_path, "%s-%d", prefix, getppid());
	l = strlen(remote.sun_path) + sizeof(remote.sun_family);
	if (connect(fd, (struct sockaddr *)&remote, l) == -1) {
		return 0;
		//dprintf(log, "comm: connect: %s\n", strerror(errno));
		//exit(1);
	}

	return fd;
}

/* Initialize the profiler runtime. */
__attribute__ ((constructor (101)))
static void
setup(void)
{
	log = comm(SOCK_SEQPACKET, "log");
	dprintf(log, "auklet: version: %s\n", AUKLET_VERSION);
	sock = comm(SOCK_STREAM, "data");
	if (!sock)
		return;
	signals();
	timers();
	instenter = push;
	instexit = pop;
}

/* Clean up the profiler runtime. */
__attribute__ ((destructor (101)))
static void
cleanup(void)
{
	instenter = enternop;
	instexit = exitnop;
	killN(&root, 1);
}

/* instrumentation interface */
void
__cyg_profile_func_enter(void *fn, void *cs)
{
	F f = {
		.fn = (uintptr_t)fn,
		.cs = (uintptr_t)cs,
	};
	instenter(&sp, f);
}

void
__cyg_profile_func_exit(void *fn, void *cs)
{
	instexit(&sp);
}
