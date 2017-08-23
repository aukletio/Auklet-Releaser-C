CFLAGS = -g -std=c99 -pedantic -D_POSIX_C_SOURCE=200809L
INSTFLAGS = -finstrument-functions

all: main main-instd

main-instd: main.c instrument.o
  gcc -v -o $@ $^ ${INSTFLAGS} ${CFLAGS}

main: main.c
  gcc -v -o $@ $< ${CFLAGS}

instrument.o: inst.c
  gcc -v -o $@ -c -g $< ${CFLAGS}

clean:
  rm -f main main-instd instrument.o

.PHONY: all clean
