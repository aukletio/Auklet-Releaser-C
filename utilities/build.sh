#!/bin/sh
cd wrapper
go get
go build

cd ../releaser
go get
go build
cd test
make

cd ../../instrument
make
