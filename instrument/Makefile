CFLAGS += -finstrument-functions -g
LDLIBS += -pthread

main: main.c instrument.a
	gcc -o $@ $^ ${CFLAGS} ${LDLIBS}

main.ll: main.c instrument.a
	clang -o $@ $^ -S -emit-llvm ${CFLAGS}

instrument.a: inst.go
	go build -buildmode=c-archive -o $@ $^

clean:
	rm -f main instrument.a instrument.h main

.PHONY: clean
