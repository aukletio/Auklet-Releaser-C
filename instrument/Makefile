INSTALL = /usr/local/lib
CFLAGS = -g -std=c99 -pedantic -D_POSIX_C_SOURCE=200809L -D_GNU_SOURCE

all: install

instrument.o: inst.c
	gcc ${CFLAGS} -o $@ -c -g $<

inst_test: inst_test.c
	gcc ${CFLAGS} -o $@ -I. $<

test: inst_test
	@./$<

libauklet.a: instrument.o
	ar rcs $@ $<

install: libauklet.a
	sudo cp $< ${INSTALL}/

uninstall:
	rm -f ${INSTALL}/libauklet.a

clean:
	rm -f instrument.o inst_test libauklet.a

.PHONY: all clean test install uninstall
