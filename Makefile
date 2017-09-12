all: build test

build: wrapper releaser instrument

wrapper:
	go install ./wrapper

releaser:
	go install ./releaser

instrument:
	make -C instrument/

test:
	make -C test/src && cd test/target && ./run.sh

.PHONY: all build wrapper releaser instrument test
