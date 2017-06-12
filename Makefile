CFLAGS += -finstrument-functions -g
LDLIBS += -pthread

main: main.c profiler.a
	gcc -o $@ $^ ${CFLAGS} ${LDLIBS}

profiler.a: profiler.go
	go build -buildmode=c-archive -o $@ $<
	cp profiler.a ~/source/st/
	#strip $@
	#ranlib $@

clean:
	rm -f main profiler.a profiler.h

.PHONY: clean
