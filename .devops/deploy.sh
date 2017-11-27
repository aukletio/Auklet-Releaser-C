#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
ENVDIR=$1
VERSION="$(cat VERSION)"
TIMESTAMP="$(date --rfc-3339=seconds | sed 's/ /T/')"
mkdir deploy

echo 'Compiling wrapper and releaser...'
export GOFLAGS="-ldflags \"-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP\""
echo 'Releaser: linux/amd64'
GOOS=linux GOARCH=amd64 go build -o release-$VERSION-linux-amd64 ./release
echo 'Releaser: windows/amd64'
GOOS=windows GOARCH=amd64 go build -o release-$VERSION-windows-amd64.exe ./release
WRAPPER_ARCHS=( amd64 arm arm64 mips64 mips64le )
for a in "${WRAPPER_ARCHS[@]}"; do
  if [[ "$a" == "arm" ]]; then
    ARM_FAM=( 5 6 7 )
    for f in "${ARM_FAM[@]}"; do
      echo "Wrapper: linux/arm$f"
      GOOS=linux GOARCH=arm GOARM=$f go build -o wrap-$VERSION-linux-arm$f ./wrap
    done
  else
    echo "Wrapper: linux/$a"
    GOOS=linux GOARCH=$a go build -o wrap-$VERSION-linux-$a ./wrap
  fi
done
mv -t deploy release-* wrap-*

echo 'Compiling/packaging profiler...'
LIBTAR="libauklet-$VERSION.tgz"
make libauklet.a
tar cz -f $LIBTAR libauklet.a

echo 'Installing AWS CLI...'
# sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
# aws s3 cp $LIBTAR s3://auklet-profiler/$ENVDIR/$LIBTAR
# cd deploy
# for f in *; do
#   aws s3 cp $f s3://auklet-profiler/$ENVDIR/$f
# done
