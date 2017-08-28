#!/bin/sh
mkdir -p ~/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME}
ln -s ${HOME}/${CIRCLE_PROJECT_REPONAME} ${HOME}/.go_project/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
go get -t -d -v ./...

cd wrapper
go build -v

cd ../releaser
go build -v
cd test
make

cd ../../instrument
make
