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
export GOFLAGS="-ldflags \"-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP\""
echo 'Releaser: linux/amd64'
GOOS=linux GOARCH=amd64 go build -o release-$VERSION-linux-amd64 ./release
echo 'Releaser: windows/amd64'
GOOS=windows GOARCH=amd64 go build -o release-$VERSION-windows-amd64.exe ./release
echo 'Wrapper: linux/amd64'
GOOS=linux GOARCH=amd64 go build -o wrap-$VERSION-linux-amd64 ./wrap
echo 'Wrapper: linux/arm'
GOOS=linux GOARCH=arm go build -o wrap-$VERSION-linux-arm ./wrap
echo 'Wrapper: linux/arm64'
GOOS=linux GOARCH=arm64 go build -o wrap-$VERSION-linux-arm64 ./wrap
echo 'Wrapper: linux/mips64'
GOOS=linux GOARCH=mips64 go build -o wrap-$VERSION-linux-mips64 ./wrap
echo 'Wrapper: linux/mips64le'
GOOS=linux GOARCH=mips64le go build -o wrap-$VERSION-linux-mips64le ./wrap

echo 'Compiling/packaging profiler...'
make libauklet.a
tar cz -f $LIBTAR libauklet.a

echo 'Installing AWS CLI...'
#sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
#aws s3 cp $LIBTAR s3://auklet-profiler/$ENVDIR/$LIBTAR
#aws s3 cp $GOPATH/bin/wrap s3://auklet-profiler/$ENVDIR/wrap-$VERSION
#aws s3 cp $GOPATH/bin/release s3://auklet-profiler/$ENVDIR/release-$VERSION
