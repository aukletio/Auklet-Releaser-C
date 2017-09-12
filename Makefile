all: build test

build: wrapper releaser instrument depend

depend:
	mkdir -p ~/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME} && \
	ln -s ${HOME}/${CIRCLE_PROJECT_REPONAME} ${HOME}/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME} && \
	go get -t -d -v ./...

wrapper:
	go install ./wrapper

releaser:
	go install ./releaser

instrument:
	make -C instrument/

test:
	make -C test/src && cd test/target && ./run.sh

.PHONY: all depend build wrapper releaser instrument test
