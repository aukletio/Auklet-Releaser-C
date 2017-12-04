/*
 * profiler library, separated from runtime for testing
 */

/* headers */
#include <errno.h>
#include <pthread.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* types */
/* Type F represents a stack frame holding function address and callsite. */
typedef struct {
	uintptr_t fn, cs;
} F;

/* Type N represents a node in a stacktree. A stacktree is an aggregation of
 * stacktraces. Stack pointers (type (N *)) should be declared thread-specific
 * (__thread). */
typedef struct n { 
	F f; 
	struct n *parent;

	/* If two sigprof calls run at the same time (in different threads),
	 * there is a possible data race on nsamp. */
	pthread_mutex_t lsamp;
	unsigned nsamp;

	/* push locks l to prevent races in the callee list. Contention, but not
	 * deadlock, is possible if two threads push at the same stack pointer. */
	pthread_mutex_t llist;
	struct n **callee;
	unsigned cap, len;

	/* push modifies ncall only when it has acquired l, to avoid races
	 * between threads. */
	pthread_mutex_t lcall;
	unsigned ncall;

	int empty;
} N;

/* Type B represents a resizable text buffer. */
typedef struct {
	char *buf;
	unsigned cap, len;
} B;

/* function declarations */
static void push(N **sp, F f);
static F pop(N **sp);
static int eqF(F a, F b);
static N *newN(F f);
static void killN(N *n, int root);
static N *hascallee(N *n, F f);
static N *addcallee(N *n, F f);
static void sample(N *sp);
static void setnotempty(N *n);

static int growB(B *b);
static int append(B *b, char *fmt, ...);

static int marshaln(B *b, N *n);
static int marshalc(B *b, N *n);
static int marshal(B *b, N *n);

static int log = 0; // stdout, initially

/* function definitions */
void
mcheck(int (*op)(pthread_mutex_t *), pthread_mutex_t *m)
{
	int ret = op(m);
	if (ret) {
		char *msg = strerror(ret);
		dprintf(log, "mcheck: ret = %d, msg = %s\n", ret, msg);
		exit(1);
	}
}

/* Push a frame f onto a stack defined by sp. */
static void
push(N **sp, F f)
{
	mcheck(pthread_mutex_lock, &(*sp)->llist);
	N *c = hascallee(*sp, f);
	if (!c) {
		c = addcallee(*sp, f);
		if (!c) {
			dprintf(log, "push: couldn't add callee\n");
			exit(1);
		}
	}
	mcheck(pthread_mutex_unlock, &(*sp)->llist);
	mcheck(pthread_mutex_lock, &c->lcall);
	++c->ncall;
	mcheck(pthread_mutex_unlock, &c->lcall);
	*sp = c;
	setnotempty(*sp);
}

/* Pop a frame off a stack sp. */
static F
pop(N **sp)
{
	F f = (*sp)->f;
	if (!(*sp)->parent) {
		dprintf(log, "pop: called with NULL parent\n");
		exit(1);
	}
	*sp = (*sp)->parent;
	return f;
}

static int
eqF(F a, F b)
{
	return ((a.fn == b.fn) && (a.cs == b.cs));
}

/* Create a new node with frame f. */
static N *
newN(F f)
{
	N *n = (N *)malloc(sizeof(N));
	//dprintf(log, "%p = malloc(%d)\n", n, sizeof(N));
	if (!n)
		return NULL;

	n->f = f;
	n->ncall = n->nsamp = n->cap = n->len = 0;
	n->callee = NULL;
	n->llist = (pthread_mutex_t)PTHREAD_MUTEX_INITIALIZER;
	n->lsamp = (pthread_mutex_t)PTHREAD_MUTEX_INITIALIZER;
	n->lcall = (pthread_mutex_t)PTHREAD_MUTEX_INITIALIZER;
	return n;
}

/* Print a node n and its callees. For debugging purposes. */
static void
dumpN(N *n, unsigned ind)
{
	char tab[] = "\t\t\t\t\t\t\t\t\t\t";
	tab[ind] = '\0';
	dprintf(log, "%s%p:\n", tab, (void *)n);
	dprintf(log, "%s    f.fn = %p\n", tab, (void *)n->f.fn);
	dprintf(log, "%s    f.cs = %p\n", tab, (void *)n->f.cs);
	dprintf(log, "%s    nsamp = %u\n", tab, n->nsamp);
	dprintf(log, "%s    ncall = %u\n", tab, n->ncall);
	dprintf(log, "%s    len/cap = %u/%u\n", tab, n->len, n->cap);
	dprintf(log, "%s    callee = %p\n", tab, (void *)n->callee);
	for (int i = 0; i < n->len; ++i)
		dumpN(n->callee[i], ind + 1);
}

/* Delete a node n and its callees. */
static void
killN(N *n, int root)
{
	for (int i = 0; i < n->len; ++i)
		killN(n->callee[i], 0);
	//dprintf(log, "free(%p)\n", n->callee);
	free(n->callee);

	/* allows us to avoid freeing the statically-allocated root
 	* (necessary for TLS initialization) */
	if (!root) {
		//dprintf(log, "free(%p)\n", n);
		free(n);
	}
}

/* Return a pointer to a callee of n matching frame f, otherwise NULL. */
static N *
hascallee(N *n, F f)
{
	for (int i = 0; i < n->len; ++i)
		if (eqF(n->callee[i]->f, f))
			return n->callee[i];
	return NULL;
}

/* Add to node n a callee with frame f. */
static N *
addcallee(N *n, F f)
{
	N *new;
	if (n->cap == n->len) {
		unsigned newcap = n->cap ? 2 * n->cap : 2;
		N **c = (N **)realloc(n->callee, newcap * sizeof(N *));
		//dprintf(log, "%p = realloc(%p, %d)\n", c, n->callee, newcap*sizeof(N *));
		if (!c)
			return NULL;
		n->callee = c;
		n->cap = newcap;
	}
	new = newN(f);
	if (!new)
		return NULL;
	new->parent = n;
	++n->len;
	n->callee[n->len - 1] = new;
	return new;
}

/* Increments nsamp in the stack defined by sp. */
static void
sample(N *sp)
{
	for (N *n = sp; n; n = n->parent) {
		mcheck(pthread_mutex_lock, &n->lsamp);
		++n->nsamp;
		mcheck(pthread_mutex_unlock, &n->lsamp);
	}
}

/* Grow a buffer. */
static int
growB(B *b)
{
	unsigned newcap = b->cap ? b->cap * 2 : 32;
	char *c = (char *)realloc(b->buf, newcap * sizeof(char));
	if (!c) {
		dprintf(log, "growB: realloc: %s\n", strerror(errno));
		return 0;
	}
	b->buf = c;
	b->cap = newcap;
	return 1;
}

/* Append a formatted string to buffer b. */
static int
append(B *b, char *fmt, ...)
{
	int wc;
	va_list ap;
retry:
	va_start(ap, fmt);
	wc = vsnprintf(b->buf + b->len, b->cap - b->len, fmt, ap);
	if (wc >= b->cap - b->len) {
		if (!growB(b))
			return 0;
		goto retry;
	}
	va_end(ap);
	b->len += wc;
	return wc;
}

static void
setnotempty(N *n)
{
	n->empty = 0;
	if (n->parent && n->parent->empty)
		setnotempty(n->parent);
}

static int
marshal(B *b, N *n)
{
	append(b, "{");

	if (n->f.fn)
		append(b, "\"fn\":%ld,", (unsigned long)n->f.fn);

	if (n->f.cs)
		append(b, "\"cs\":%ld,", (unsigned long)n->f.cs);

	mcheck(pthread_mutex_lock, &n->lcall);
	if (n->ncall)
		append(b, "\"ncalls\":%u,", n->ncall);
	mcheck(pthread_mutex_unlock, &n->lcall);

	mcheck(pthread_mutex_lock, &n->lsamp);
	if (n->nsamp)
		append(b, "\"nsamples\":%u,", n->nsamp);
	mcheck(pthread_mutex_unlock, &n->lsamp);

	/* It's convenient, but hacky, to clear the counters here while we're
	 * walking the tree. All counters are cleared after a tree is emitted,
	 * so that the next tree begins at zero.  We also set the empty flag
	 * so that future calls to marshal can skip empty branches. */
	n->ncall = 0;
	n->nsamp = 0;
	n->empty = 1;

	mcheck(pthread_mutex_lock, &n->llist);
	if (n->len) {
		append(b, "\"callees\":[");
		for (int i = 0; i < n->len; ++i) {
			if (n->callee[i]->empty)
				continue;
			marshal(b, n->callee[i]);
			append(b, ",");
		}
		if (',' == b->buf[b->len - 1])
			b->len -= 1;
		append(b, "]");
	} else {
		if (',' == b->buf[b->len - 1])
			b->len -= 1;
	}
	mcheck(pthread_mutex_unlock, &n->llist);

	append(b, "}");
}

static int
sane(N *n)
{
	int ok = 1;
	unsigned sum = 0;

	mcheck(pthread_mutex_lock, &n->lsamp);
	for (int i = 0; i < n->len; ++i) {
		if (!sane(n->callee[i]))
			ok = 0;
		sum += n->callee[i]->nsamp;
	}

	if (n->nsamp < sum) {
		dprintf(log, "sane: %p->nsamp = %u, sum = %u\n", (void *)n, n->nsamp, sum);
		dumpN(n, 0);
		ok = 0;
	}
	mcheck(pthread_mutex_unlock, &n->lsamp);
	return ok;
}
