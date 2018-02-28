#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
ENVDIR=$1
VERSION="$(cat ~/.version)"
VERSION_SIMPLE=$(cat ~/.version | xargs | cut -f1 -d"+")
export TIMESTAMP="$(date --rfc-3339=seconds | sed 's/ /T/')"
GO_LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP"

echo 'Compiling releaser...'
echo '=== linux/amd64 ==='
GOOS=linux GOARCH=amd64 go build -ldflags "$GO_LDFLAGS" -o release-linux-amd64-$VERSION ./release
echo '=== windows/amd64 ==='
GOOS=windows GOARCH=amd64 go build -ldflags "$GO_LDFLAGS" -o release-windows-amd64-$VERSION.exe ./release
echo

echo 'Installing AWS CLI...'
sudo apt -y install awscli > /dev/null 2>&1

if [[ "$ENVDIR" == "production" ]]; then
  echo 'Erasing production releaser binaries in public S3...'
  aws s3 rm s3://auklet/release/latest/ --recursive
fi

echo 'Uploading releaser binaries to S3...'
# Iterate over each file and upload it to S3.
for f in {release-}*; do
  # Upload to the internal bucket.
  S3_LOCATION="s3://auklet-profiler/$ENVDIR/$VERSION/$f"
  aws s3 cp $f $S3_LOCATION
  # Upload to the public bucket for production builds.
  if [[ "$ENVDIR" == "production" ]]; then
    # Get the component name.
    COMPONENT=$(echo $f | cut -f1 -d"-")
    # Copy to the public versioned directory.
    VERSIONED_NAME="${f/$VERSION/$VERSION_SIMPLE}"
    aws s3 cp $S3_LOCATION s3://auklet/$COMPONENT/$VERSION_SIMPLE/$VERSIONED_NAME
    # Copy to the public "latest" directory.
    LATEST_NAME="${f/$VERSION/latest}"
    aws s3 cp $S3_LOCATION s3://auklet/$COMPONENT/latest/$LATEST_NAME
  fi
done
