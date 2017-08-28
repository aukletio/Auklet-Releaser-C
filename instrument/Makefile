CFLAGS = -g -std=c99 -pedantic -D_POSIX_C_SOURCE=200809L
INSTFLAGS = -finstrument-functions

all: main main-instd inst_test

main-instd: main.c instrument.o
	gcc ${CFLAGS} -o $@ $^ ${INSTFLAGS}

main: main.c
	gcc ${CFLAGS} -o $@ $<

instrument.o: inst.c
	gcc ${CFLAGS} -o $@ -c -g $<

inst_test: inst_test.c
	gcc ${CFLAGS} -o $@ -I. $<

test: inst_test
	@./$<

clean:
	rm -f main main-instd instrument.o inst_test

.PHONY: all clean test
