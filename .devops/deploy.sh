#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
ENVDIR=$1
VERSION="$(cat VERSION)"
LIBTAR="libauklet-$VERSION.tgz"

echo 'Compiling profiler...'
make go libauklet.a

echo 'Packaging profiler...'
tar czv -f $LIBTAR libauklet.a

echo 'Installing AWS CLI...'
sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
aws s3 cp $LIBTAR s3://auklet-profiler/$ENVDIR/$LIBTAR
aws s3 cp $GOPATH/bin/wrap s3://auklet-profiler/$ENVDIR/wrap-$VERSION
aws s3 cp $GOPATH/bin/release s3://auklet-profiler/$ENVDIR/release-$VERSION
