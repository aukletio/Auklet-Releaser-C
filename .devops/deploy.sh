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
echo 'OS/Arch: linux/amd64'
GOOS=linux GOARCH=amd64 go build -o release-$VERSION-linux-amd64 ./release
echo 'OS/Arch: windows/amd64'
GOOS=windows GOARCH=amd64 go build -o release-$VERSION-windows-amd64.exe ./release
echo

echo 'Preparing for cross compilation...'
echo 'deb http://emdebian.org/tools/debian/ jessie main' | sudo tee /etc/apt/sources.list.d/crosstools.list
curl -sS http://emdebian.org/tools/debian/emdebian-toolchain-archive.key | sudo apt-key add -
echo

echo 'Compiling wrapper/library combinations...'
echo
export GOOS=linux
while IFS=, read arch cc ar pkg
do
  echo "OS/Arch: $GOOS/$arch"
  if [[ "$pkg" != "" ]]; then
    echo "Installing $pkg cross compilation toolchain..."
    sudo dpkg --add-architecture $pkg
    sudo apt-get -qq update
    sudo apt-get -y install crossbuild-essential-$pkg > /dev/null
    echo "$pkg cross compilation toolchain installed; proceeding with compilation..."
  fi
  if [[ "$arch" == "arm" ]]; then
    ARM_FAM=(5 6 7)
    for fam in "${ARM_FAM[@]}"; do
      echo "ARM family: $fam"
      echo 'Compiling wrapper...'
      GOARCH=arm GOARM=$fam go build -v -o wrap-$VERSION-$GOOS-arm$fam ./wrap
      echo 'Compiling library...'
      CC=$cc AR=$ar TARNAME="libauklet-$VERSION-$GOOS-arm$fam.tgz" make libauklet.tgz
    done
  else
    echo 'Compiling wrapper...'
    GOARCH=$arch go build -v -o wrap-$VERSION-$GOOS-$arch ./wrap
    echo 'Compiling library...'
    CC=$cc AR=$ar TARNAME="libauklet-$VERSION-$GOOS-$arch.tgz" make libauklet.tgz
  fi
  echo "DONE: $GOOS/$arch"
  echo
done < packaging-grid.csv
mv -t deploy release-* wrap-* libauklet-*

echo 'Installing AWS CLI...'
sudo apt-get -y install awscli

echo 'Uploading profiler to S3...'
cd deploy
for f in *; do aws s3 cp $f s3://auklet-profiler/$ENVDIR/$VERSION/$f; done
