CFLAGS += -finstrument-functions -g

all: main main-instd inst_test

main-instd: main.c instrument.o
	gcc -o $@ $^ ${CFLAGS}

main: main.c
	gcc -o $@ $<

instrument.o: inst.c
	gcc -o $@ -c -g $<

inst_test: inst_test.c
	gcc -o $@ -I. $<

test: inst_test
	@./$<

clean:
	rm -f main main-instd instrument.o inst_test

.PHONY: all clean test
