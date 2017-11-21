#!/bin/bash

echo 'Compiling profiler...'
make
echo 'Packaging profiler...'
VERSION="$(cat VERSION)"
LIBTAR="libauklet-$VERSION.tgz"
tar czv -f $LIBTAR libauklet.a
echo 'Uploading profiler to S3...'
sudo apt-get -y install awscli
aws s3 cp $LIBTAR s3://auklet-profiler/$LIBTAR
aws s3 cp $GOPATH/bin/wrap s3://auklet-profiler/wrap-$VERSION
aws s3 cp $GOPATH/bin/release s3://auklet-profiler/release-$VERSION
