CFLAGS += -finstrument-functions -g

all: main main-instd

main-instd: main.c instrument.o
	gcc -v -o $@ $^ ${CFLAGS}

main: main.c
	gcc -v -o $@ $<

instrument.o: inst.c
	gcc -v -o $@ -c -g $<

clean:
	rm -f main main-instd instrument.o

.PHONY: all clean
