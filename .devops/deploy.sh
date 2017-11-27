#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
ENVDIR=$1
VERSION="$(cat VERSION)"
TIMESTAMP="$(date --rfc-3339=seconds | sed 's/ /T/')"
LIBTAR="libauklet-$VERSION.tgz"

echo 'Compiling wrapper and releaser...'
mkdir deploy
export GOFLAGS="-ldflags \"-X main.Version=$VERSION main.BuildDate=$TIMESTAMP\""
echo 'Releaser: linux/amd64'
GOOS=linux GOARCH=amd64 make release && cp $GOPATH/bin/release deploy/release-$VERSION-linux-amd64
echo 'Releaser: windows/amd64'
GOOS=windows GOARCH=amd64 make release && cp $GOPATH/bin/release deploy/release-$VERSION-windows-amd64.exe
echo 'Wrapper: linux/amd64'
GOOS=linux GOARCH=amd64 make wrap && cp $GOPATH/bin/release deploy/wrap-$VERSION-linux-amd64
echo 'Wrapper: linux/arm'
GOOS=linux GOARCH=arm make wrap && cp $GOPATH/bin/release deploy/wrap-linux-arm
echo 'Wrapper: linux/arm64'
GOOS=linux GOARCH=arm64 make wrap && cp $GOPATH/bin/release deploy/wrap-$VERSION-linux-arm64
echo 'Wrapper: linux/mips64'
GOOS=linux GOARCH=mips64 make wrap && cp $GOPATH/bin/release deploy/wrap-$VERSION-linux-mips64
echo 'Wrapper: linux/mips64le'
GOOS=linux GOARCH=mips64le make wrap && cp $GOPATH/bin/release deploy/wrap-$VERSION-linux-mips64le

echo 'Compiling/packaging profiler...'
make libauklet.a
tar cz -f $LIBTAR libauklet.a

echo 'Installing AWS CLI...'
sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
aws s3 cp $LIBTAR s3://auklet-profiler/$ENVDIR/$LIBTAR
aws s3 cp $GOPATH/bin/wrap s3://auklet-profiler/$ENVDIR/wrap-$VERSION
aws s3 cp $GOPATH/bin/release s3://auklet-profiler/$ENVDIR/release-$VERSION
