CFLAGS += -finstrument-functions -g

all: main main-instd

main-instd: main.c instrument.o
	gcc -o $@ $^ ${CFLAGS}

main: main.c
	gcc -o $@ $<

instrument.o: inst.c
	gcc -o $@ -c -g -pg $<

clean:
	rm -f main main-instd instrument.o

.PHONY: all clean
