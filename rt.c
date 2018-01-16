/*
 * profiler runtime
 */

/* headers */
#include "lib.c"

#include <semaphore.h>
#include <signal.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <sys/un.h>
#include <unistd.h>

/* macros */
#define SIGRT(n) (SIGRTMIN + (n))

/* types */
enum {
	REAL,
	VIRT,
};

/* function declarations */
static void sigprof(int n);
static void signals(void);
static void siginstall(int sig, void (*handler)(int));
static void emit(void);
static void *timer(void *);
static void timers(void);
static void mktimers(void);
static void settimers(void);
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
static int sock, stack;
static sem_t sem;
static struct {
	clockid_t clk;
	struct timespec value;
	timer_t tid;
} tmr[] = {
	/*        clk                       value */
	[REAL] = {CLOCK_REALTIME,           {.tv_sec = 10, .tv_nsec = 0}},
	[VIRT] = {CLOCK_PROCESS_CPUTIME_ID, {.tv_sec =  1, .tv_nsec = 0}},
};


/* function definitions */
/* Increment sample counters in the stack of the current thread. */
static void
sigprof(int n)
{
	sample(sp);
}

static void
sigemit(int n)
{
	dprintf(log, "sigemit(%d)\n", n);
	settimers();
	sem_post(&sem);
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
	if (write(sock, b.buf, b.len) == -1) {
		dprintf(log, "emit: write: %s\n", strerror(errno));
		//exit(1);
	}
	free(b.buf);
}

/* Send a JSON-encoded stacktrace to the wrapper. */
static void
stacktrace(int sig)
{
	B b = {0, 0, 0};
	marshals(&b, sp, sig);
	append(&b, "\n");
	//dprintf(log, "stacktrace: %s", b.buf);
	if (write(stack, b.buf, b.len) == -1) {
		dprintf(log, "stacktrace: write: %s\n", strerror(errno));
	}
	free(b.buf);
}

/* Handle some kind of error signal. */
static void
sigerr(int n)
{
	stacktrace(n);
	_exit(EXIT_FAILURE);
}

/* Set up signal handlers. */
static void
signals(void)
{
	struct {
		int sig;
		void (*handler)(int);
	} s[] = {
		{SIGSEGV,     sigerr },
		{SIGILL,      sigerr },
		{SIGFPE,      sigerr },
		{SIGPROF,     sigprof},
		{SIGRT(REAL), sigemit},
		{SIGRT(VIRT), sigemit},
	};
	sem_init(&sem, 0, 0);
	for (int i = 0; i < len(s); ++i)
		siginstall(s[i].sig, s[i].handler);
}

static void
siginstall(int sig, void (*handler)(int))
{
	struct sigaction sa;
	sigaction(sig, NULL, &sa);
	sa.sa_handler = handler;

	/* sigfillset prevents handler from getting interrupted, but this
	 * does not avoid races between handlers in different threads. */
	sigfillset(&sa.sa_mask);
	sigaction(sig, &sa, NULL);
}

/* Emit profile data periodically. Implementing this as a interval timer +
 * signal handler combo would interrupt--and therefore deadlock--instrumented
 * threads, which must acquire locks. To avoid this, it is a dedicated thread. */
static void *
timer(void *p)
{
	sigset_t s;
	sigfillset(&s);
	pthread_sigmask(SIG_BLOCK, &s, NULL);
	while (1) {
		sem_wait(&sem);
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
	mktimers();
	settimers();
	pthread_t t;
	pthread_create(&t, NULL, timer, NULL);
}

static void
mktimers(void)
{
	for (int i = 0; i < len(tmr); ++i) {
		timer_create(tmr[i].clk, &(struct sigevent){
			.sigev_notify = SIGEV_SIGNAL,
			.sigev_signo  = SIGRT(i),
		}, &tmr[i].tid);
	}
}

static void
settimers(void)
{
	for (int i = 0; i < len(tmr); ++i) {
		timer_settime(tmr[i].tid, 0, &(struct itimerspec){
			.it_interval = {0, 0},
			.it_value    = tmr[i].value,
		}, NULL);
	}
}

/* Set up a communication channel with the wrapper. */
static int
comm(int type, char *prefix)
{
	struct sockaddr_un remote;
	int l, fd;
	if ((fd = socket(AF_UNIX, type, 0)) == -1) {
		dprintf(log, "comm: socket: %s\n", strerror(errno));
		return 0;
	}
	remote.sun_family = AF_UNIX;
	sprintf(remote.sun_path, "%s-%d", prefix, getppid());
	l = strlen(remote.sun_path) + sizeof(remote.sun_family);
	if (connect(fd, (struct sockaddr *)&remote, l) == -1) {
		dprintf(log, "comm: connect: %s\n", strerror(errno));
		return 0;
	}

	return fd;
}

/* Initialize the profiler runtime. */
__attribute__ ((constructor (101)))
static void
setup(void)
{
	log = comm(SOCK_SEQPACKET, "log");
	dprintf(log, "Auklet Instrument version %s (%s)\n", AUKLET_VERSION, AUKLET_TIMESTAMP);
	sock = comm(SOCK_STREAM, "data");
	stack = comm(SOCK_STREAM, "stacktrace");
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
