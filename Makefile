INSTFLAGS = -finstrument-functions
INSTLIBS = -lauklet
INST_INSTALL = /usr/local/lib
CFLAGS = -g -std=c99 -D_POSIX_SOURCE=200809L -D_GNU_SOURCE

all: profiler test

.PHONY: all

# CircleCI commands
depend:
	mkdir -p ~/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME} && \
	ln -s ${HOME}/${CIRCLE_PROJECT_REPONAME} ${HOME}/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME} && \
	go get -t -d -v ./...

.PHONY: depend

# Profiler components
profiler: wrapper releaser instrument

wrapper: ${WRAPPER}
	go install ./wrapper

releaser: releaser/*.go
	go install ./releaser

instrument: instrument/libauklet.a
	sudo cp $< ${INST_INSTALL}

instrument-clean:
	rm -f instrument/libauklet.a instrument/instrument.o

instrument-uninstall:
	sudo rm -f ${INST_INSTALL}/libauklet.a

.PHONY: profiler wrapper releaser instrument instrument-clean instrument-uninstall

instrument/libauklet.a: instrument/instrument.o
	ar rcs $@ $<

instrument/instrument.o: instrument/inst.c
	gcc ${CFLAGS} -o $@ -c -g $<

# Test program components
test: test-release test-install

test-run: test
	cd test/target && ./run.sh

test-release: test/src/snellius test/src/snellius-debug
	. test/src/secrets.sh &&\
	export AUKLET_RELEASE_ENDPOINT=production &&\
	releaser -appid $$APP_ID\
	         -apikey $$API_KEY\
	         -debug test/src/snellius-debug\
	         -deploy test/src/snellius

test-install: test/src/snellius
	sudo cp $< /usr/local/bin/

test-clean:
	rm -f test/src/snellius test/src/snellius-debug

test-uninstall:
	sudo rm -f /usr/local/bin/snellius

.PHONY: test test-release test-install test-clean test-uninstall

test/src/snellius: test/src/snellius-debug
	cp $< $@
	strip $@

test/src/snellius-debug: test/src/snellius.c
	gcc -o $@ $< ${CFLAGS} ${INSTFLAGS} ${INSTLIBS}

clean: test-clean instrument-clean

uninstall: test-uninstall instrument-uninstall

.PHONY: clean
