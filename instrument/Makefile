CFLAGS += -finstrument-functions -g

main: main.c instrument.o
	gcc -o $@ $^ ${CFLAGS}

main.ll: main.c instrument.a
	clang -o $@ $^ -S -emit-llvm ${CFLAGS}

instrument.o: inst.c
	gcc -o $@ -c $<

clean:
	rm -f main main instrument.o

.PHONY: clean
