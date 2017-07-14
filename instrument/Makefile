CFLAGS += -finstrument-functions -g

main: main.c instrument.o
	gcc -o $@ $^ ${CFLAGS}

instrument.o: inst.c
	gcc -o $@ -c $<

clean:
	rm -f main instrument.o

.PHONY: clean
