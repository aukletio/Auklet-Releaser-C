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
#define len(x) (sizeof(x)/sizeof(x[0]))

/* fault injector */
#if defined(FAULT_RATE)
	static void *
	fault_inject(void *ptr, size_t size)
	{
		if (rand() < RAND_MAX/FAULT_RATE) {
			printf("fault injected\n");
			return NULL;
		} else {
			return realloc(ptr, size);
		}
	}

	#define malloc(size)       fault_inject(NULL, (size))
	#define realloc(ptr, size) fault_inject((ptr), (size))
#endif

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
static void run_enablers(void);
static void run_disablers(void);
static int enable_inst(void);
static void disable_inst(void);
static int enable_timers(void);
static void disable_timers(void);
static int enable_dtor_cb(void);
static void disable_dtor_cb(void);
static int enable_sigactions(void);
static void disable_sigactions(void);
static int alloc_storage(void);
static void free_storage(void);
static int enable_socket(void);
static void disable_socket(void);

static Node *newNode(Node *parent, Frame f);
static void killNode(Node *n);
static Node *hasCallee(Node *n, Frame f);
static Node *appendCallee(Node *n, Frame f);
static void setnotempty(Node *n);

static int marshal(Node *n, String *s);
static int newString(String *s, unsigned size);
static int grow(String *s);
static int sappend(String *s, const char *format, ...);

static void push(Frame);
static void pop(Frame);
static void inst_nop(Frame);

static __attribute__ ((constructor)) void init(void);
static __attribute__ ((destructor)) void fini(void);
void __cyg_profile_func_enter(void *fn, void *cs);
void __cyg_profile_func_exit(void *fn, void *cs);
static void sig_sample(int n);
static void sig_emit(int n);
static int emit(void);
static void sig_cleanup(int n);
static void dtor_cleanup(void);
static void dtor_nop(void);

/* global variables */
static int termsig[] = {SIGHUP, SIGINT, SIGPIPE, SIGTERM, SIGUSR1, SIGUSR2};
static struct sigaction old_sample_act, old_emit_act, old_term_act[len(termsig)];
static String s;
static void (*dtor_cb)(void) = dtor_nop;
static void (*enter_action)(Frame f) = inst_nop;
static void (*exit_action)(Frame f) = inst_nop;
static Node *tree, *tp;
static int sock;
static struct {
	char *name;
	int (*enabler)(void);
	void (*disabler)(void);
} config[] = {
	{"socket",     enable_socket,     disable_socket    },
	{"storage",    alloc_storage,     free_storage      },
	{"inst",       enable_inst,       disable_inst      },
	{"dtor_cb",    enable_dtor_cb,    disable_dtor_cb   },
	{"timers",     enable_timers,     disable_timers    },
	{"sigactions", enable_sigactions, disable_sigactions},
};
static int cp = 0;

/* function definitions */
static void
run_enablers(void)
{
	for (cp = 0; cp < len(config); ++cp) {
		if (config[cp].enabler())
			goto error;
	}
	return;
error:
	run_disablers();
}

static void
run_disablers(void)
{
	while (cp > 0) {
		--cp;
		config[cp].disabler();
	}
}

/* Each of the following function pairs is responsible for enabling and
 * disabling a particular aspect of global state necessary for the instrument to
 * run. They allow us to systematically handle any errors by backing out of
 * whatever action was underway, whether initialization or profiling. After an
 * error has occurred, the instrument's goal is to completely shutdown. */

static int
enable_inst(void)
{
	enter_action = push;
	exit_action = pop;
	return 0;
}

static void
disable_inst(void)
{
	enter_action = inst_nop;
	exit_action = inst_nop;
}

static int
enable_timers(void)
{
	struct timeval sample_pd = {.tv_sec = 0, .tv_usec = 10000};
	struct timeval emit_pd = {.tv_sec = 1, .tv_usec = 0};
	struct itimerval sample_timer = {
		.it_interval = sample_pd,
		.it_value = sample_pd,
	}, emit_timer = { 
		.it_interval = emit_pd,
		.it_value = emit_pd,
	};

	if (setitimer(ITIMER_VIRTUAL, &emit_timer, NULL) == -1)
		goto error;
	if (setitimer(ITIMER_PROF, &sample_timer, NULL) == -1)
		goto error;
	return 0;
error:
	perror("enable_timers");
	return 1;
}

static void
disable_timers(void)
{
}

static int
enable_dtor_cb(void)
{
	dtor_cb = dtor_cleanup;
	return 0;
}

static void
disable_dtor_cb(void)
{
	dtor_cb = dtor_nop;
}

static int
enable_sigactions(void)
{
	struct sigaction sample_act, emit_act, term_act;
	if (sigaction(SIGVTALRM, NULL, &old_emit_act) == -1)
		goto error;
	if (sigaction(SIGPROF, NULL, &old_sample_act) == -1)
		goto error;

	sample_act = old_sample_act;
	emit_act = old_emit_act;
	sample_act.sa_handler = sig_sample;
	emit_act.sa_handler = sig_emit;

	if (sigfillset(&sample_act.sa_mask) == -1)
		goto error;
	if (sigfillset(&emit_act.sa_mask) == -1)
		goto error;

	if (sigaction(SIGVTALRM, &emit_act, NULL) == -1)
		goto error;
	if (sigaction(SIGPROF, &sample_act, NULL) == -1)
		goto error;

	for (int i = 0; i < len(termsig); ++i) {
		if (sigaction(termsig[i], NULL, &old_term_act[i]) == -1)
			goto error;

		term_act = old_term_act[i];
		term_act.sa_handler = sig_cleanup;
		if (sigfillset(&term_act.sa_mask) == -1)
			goto error;

		if (SIG_DFL == old_term_act[i].sa_handler)
			if (sigaction(termsig[i], &term_act, NULL) == -1)
				goto error;
	}
	return 0;
error:
	perror("enable_sigaction");
	return 1;
}

static void
disable_sigactions(void)
{
	old_emit_act.sa_handler = SIG_IGN;
	old_sample_act.sa_handler = SIG_IGN;
	if (sigaction(SIGVTALRM, &old_emit_act, NULL) == -1)
		goto error;
	if (sigaction(SIGPROF, &old_sample_act, NULL) == -1)
		goto error;
	for (int i = 0; i < len(termsig); ++i)
		if (sigaction(termsig[i], &old_term_act[i], NULL) == -1)
			goto error;
	return;
error:
	perror("disable_sigactions");
}

static int
alloc_storage(void)
{
	tree = newNode(NULL, (Frame){0, 0});
	if (!tree)
		return 1;
	tp = tree;

	if (newString(&s, 8000))
		return 1;
	return 0;
}

static void
free_storage(void)
{
	killNode(tree);
	tree = NULL;
	tp = NULL;

	if (s.cap)
		free(s.buf);
}

static int
enable_socket(void)
{
	struct sockaddr_un remote;
	int length;
	if ((sock = socket(AF_UNIX, SOCK_STREAM, 0)) == -1)
		goto error;

	remote.sun_family = AF_UNIX;
	sprintf(remote.sun_path, "socket-%d", getppid());
	length = strlen(remote.sun_path) + sizeof(remote.sun_family);
	if (connect(sock, (struct sockaddr *)&remote, length) == -1)
		goto error;

	return 0;
error:
	perror("enable_socket");
	return 1;
}

static void
disable_socket(void)
{
	if (close(sock) == -1)
		goto error;
	return;
error:
	perror("disable_socket");
}

/* Node functions */
static Node *
newNode(Node *parent, Frame f)
{
	Node *new = (Node *)malloc(sizeof(Node));
	if (!new)
		return NULL;

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
killNode(Node *n)
{
	for (int i = 0; i < n->len; ++i)
		killNode(n->callee[i]);
	free(n->callee);
	free(n);
}

static Node *
hasCallee(Node *n, Frame f)
{
	Node *callee = NULL;

	/* Search the callee list for the given callsite. Some functions have a
	 * callsite of zero (usually callbacks), so we check both function
	 * address and callsite. */
	for (int i = 0; i < n->len; ++i) {
		if ((n->callee[i]->frame.cs == f.cs) &&
		    (n->callee[i]->frame.fn == f.fn)) {
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

		if (!new)
			return NULL;

		++n->cap;
		n->callee = new;
	}

	node = newNode(n, f);
	if (!node)
		return NULL;
	++n->len;
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

/* String functions */
static int
marshal(Node *n, String *s)
{
	if (sappend(s, "{"))
		return 1;

	if (n->frame.fn)
		if (sappend(s, "\"fn\":%ld,", (unsigned long)n->frame.fn))
			return 1;

	if (n->frame.cs)
		if (sappend(s, "\"cs\":%ld,", (unsigned long)n->frame.cs))
			return 1;

	if (n->ncalls)
		if (sappend(s, "\"ncalls\":%u,", n->ncalls))
			return 1;

	if (n->nsamples)
		if (sappend(s, "\"nsamples\":%u,", n->nsamples))
			return 1;

	/* It's convenient, but hacky, to clear the counters here while we're
	 * walking the tree. All counters are cleared after a tree is emitted,
	 * so that the next tree begins at zero.  We also set the empty flag
	 * so that future calls to marshal can skip empty branches. */
	n->ncalls = 0;
	n->nsamples = 0;
	n->empty = 1;

	if (n->len) {
		if (sappend(s, "\"callees\":["))
			return 1;
		for (int i = 0; i < n->len; ++i) {
			if (n->callee[i]->empty)
				continue;
			if (marshal(n->callee[i], s))
				return 1;
			if (sappend(s, ","))
				return 1;
		}
		if (',' == s->buf[s->len - 1])
			s->len -= 1;
		if (sappend(s, "]"))
			return 1;
	} else {
		if (',' == s->buf[s->len - 1])
			s->len -= 1;
	}

	if (sappend(s, "}"))
		return 1;
	return 0;
}

static int
newString(String *s, unsigned size)
{
	char *new = (char *)malloc(size * sizeof(char));
	if (!new)
		goto error;

	s->buf = new;
	s->cap = size;
	s->len = 0;
	return 0;
error:
	perror("newString");
	return 1;
}

static int
grow(String *s)
{
	char *new = (char *)realloc(s->buf, s->cap * 2 * sizeof(char));
	if (!new)
		goto error;
	s->buf = new;
	s->cap *= 2;
	return 0;
error:
	perror("grow");
	return 1;
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
		if (grow(s))
			return 1;
		goto retry;
	}

	/* Successful write. Increment the slice length to indicate the
	 * current length. */
	va_end(ap);
	s->len += wc;
	return 0;
}

/* Frame functions */
static void
push(Frame f)
{
	/* Which thread does this Frame belong to? */
	Node *callee = hasCallee(tp, f);
	if (!callee) {
		/* This is the first time visiting this part of the program; we
		 * need to grow the list of callees at the current location. */
		callee = appendCallee(tp, f);
		if (!callee)
			goto error;
	}

	/* Move tp to the current Node. */
	tp = callee;
	return;
error:
	printf("push: appendCallee failed\n");
	run_disablers();
}

static void
pop(Frame f)
{
	++tp->ncalls;
	setnotempty(tp);

	if (!tp->parent)
		goto error;

	tp = tp->parent;
	return;
error:
	printf("pop: called with null parent\n");
	run_disablers();
}

static void
inst_nop(Frame f)
{
}

/* init, fini, __cyg_profile_func_enter, __cyg_profile_func_exit, sig_sample, and
 * sig_emit implement the main features of the instrument library. */
static __attribute__ ((constructor)) void
init(void)
{
#if defined(FAULT_RATE)
	srand(FAULT_RATE);
#endif
	run_enablers();
}

static __attribute__ ((destructor)) void
fini(void)
{
	dtor_cb();
}

static void
dtor_cleanup(void)
{
	emit();
	run_disablers();
}

static void
dtor_nop(void)
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

static void
sig_sample(int n)
{
	++tp->nsamples;
	setnotempty(tp);
}

static void
sig_emit(int n)
{
	if (emit()) {
		printf("sig_emit: emit failed\n");
		run_disablers();
	}
}

static int
emit(void)
{
	if (marshal(tree, &s)) {
		printf("emit: marshal failed\n");
		goto error;
	}

	if (sappend(&s, "\n")) {
		printf("emit: sappend failed\n");
		goto error;
	}

	/* Send a tree over the socket. */
	if (send(sock, s.buf, s.len, 0) == -1) {
		perror("emit");
		goto error;
	}

	s.len = 0;
	return 0;
error:
	return 1;
}

static void
sig_cleanup(int n)
{
	/* This function is responsible for handling any signal that causes
	 * termination, namely, one that is listed in termsig. It flushes the
	 * profile, shuts down the instrument, and exits. */

	if (emit())
		printf("sig_cleanup: emit failed\n");
	run_disablers();
	exit(1);
}
