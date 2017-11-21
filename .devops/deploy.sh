#!/bin/bash
set -e
VERSION="$(cat VERSION)"
LIBTAR="libauklet-$VERSION.tgz"

echo 'Compiling profiler...'
make deploy

echo 'Packaging profiler...'
tar czv -f $LIBTAR libauklet.a

echo 'Installing AWS CLI...'
sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
aws s3 cp $LIBTAR s3://auklet-profiler/$LIBTAR
aws s3 cp $GOPATH/bin/wrap s3://auklet-profiler/wrap-$VERSION
aws s3 cp $GOPATH/bin/release s3://auklet-profiler/release-$VERSION
