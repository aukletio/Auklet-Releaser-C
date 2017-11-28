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
export GOFLAGS="-ldflags \"-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP\""

echo 'Compiling releaser...'
echo 'Releaser: linux/amd64'
GOOS=linux GOARCH=amd64 go build -o release-$VERSION-linux-amd64 ./release
echo 'Releaser: windows/amd64'
GOOS=windows GOARCH=amd64 go build -o release-$VERSION-windows-amd64.exe ./release

echo 'Compiling wrapper and library...'
export GOOS=linux
while IFS=, read arch cc ar pkg
do
  if [[ "$pkg" != "" ]]; then
    apt-get -y install $pkg
  fi
  if [[ "$a" == "arm" ]]; then
    ARM_FAM=(5 6 7)
    for f in "${ARM_FAM[@]}"; do
      echo "Wrapper: linux/arm$f"
      GOARCH=arm GOARM=$f go build -o wrap-$VERSION-$GOOS-arm$f ./wrap
      CC=$cc AR=$ar TARNAME="libauklet-$VERSION-$GOOS-arm$f.tgz" make libauklet.tgz
    done
  else
    echo "Wrapper: linux/$a"
     GOARCH=$arch go build -o wrap-$VERSION-$GOOS-$arch ./wrap
     CC=$cc AR=$ar TARNAME="libauklet-$VERSION-$GOOS-$arch.tgz" make libauklet.tgz
  fi
done < packaging-grid.csv
mv -t deploy release-* wrap-* libauklet-*

echo 'Installing AWS CLI...'
# sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
# aws s3 cp $LIBTAR s3://auklet-profiler/$ENVDIR/$LIBTAR
# cd deploy
# for f in *; do
#   aws s3 cp $f s3://auklet-profiler/$ENVDIR/$f
# done
