#!/bin/bash
set -e
if [[ "$1" == "" ]]; then
  echo "ERROR: env not provided."
  exit 1
fi
TARGET_ENV=$1
VERSION="$(cat ~/.version)"
VERSION_SIMPLE=$(cat VERSION | xargs | cut -f1 -d"+")
export TIMESTAMP="$(date --rfc-3339=seconds | sed 's/ /T/')"

echo 'Gathering license files for dependencies...'
REPO_DIR=$(eval cd $CIRCLE_WORKING_DIRECTORY ; pwd)
LICENSES_DIR="$REPO_DIR/cmd/release/licenses"
cp LICENSE $LICENSES_DIR
cd .devops
npm install --no-spin follow-redirects@1.5.0 > /dev/null 2>&1
node licenses.js "$REPO_DIR" "$LICENSES_DIR"
rm -rf node_modules package-lock.json
cd ..
echo

echo 'Generating packed resource files...'
curl -sSL https://github.com/gobuffalo/packr/releases/download/v1.11.0/packr_1.11.0_linux_amd64.tar.gz | tar -xz packr
./packr -v -z
echo

echo 'Compiling releaser...'
GO_LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP"
PREFIX='auklet-releaser'
S3_PREFIX='auklet/c/releaser'
GOOS=linux GOARCH=amd64 go build -ldflags "$GO_LDFLAGS" -o $PREFIX-linux-amd64-$VERSION_SIMPLE ./cmd/release
echo

echo 'Installing AWS CLI...'
sudo apt -y install awscli > /dev/null 2>&1

if [[ "$TARGET_ENV" == "release" ]]; then
  echo 'Erasing production C releaser binaries in S3...'
  aws s3 rm s3://$S3_PREFIX/latest/ --recursive
fi

echo 'Uploading C releaser binaries to S3...'
# Iterate over each file and upload it to S3.
for f in ${PREFIX}-*; do
  # Upload to the internal bucket.
  S3_LOCATION="s3://$S3_PREFIX/$VERSION_SIMPLE/$f"
  aws s3 cp $f $S3_LOCATION
  # Copy to the "latest" dir for production builds.
  if [[ "$TARGET_ENV" == "release" ]]; then
    LATEST_NAME="${f/$VERSION_SIMPLE/latest}"
    aws s3 cp $S3_LOCATION s3://$S3_PREFIX/latest/$LATEST_NAME
  fi
done
