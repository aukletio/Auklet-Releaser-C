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
if [[ "$1" == "staging" ]]; then
  BASE_URL='https://api-staging.auklet.io'
elif [[ "$1" == "qa" ]]; then
  BASE_URL='https://api-qa.auklet.io'
else
  BASE_URL='https://api.auklet.io'
fi

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
GO_LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$TIMESTAMP -X github.com/ESG-USA/Auklet-Releaser-C/config.StaticBaseURL=$BASE_URL"
PREFIX='auklet-releaser'
S3_BUCKET='auklet'
S3_PREFIX='releaser'
GOOS=linux GOARCH=amd64 go build -ldflags "$GO_LDFLAGS" -o $PREFIX-linux-amd64-$VERSION ./cmd/release
echo

echo 'Installing AWS CLI...'
sudo apt -y install awscli > /dev/null 2>&1

if [[ "$ENVDIR" == "production" ]]; then
  echo 'Erasing production releaser binaries in public S3...'
  aws s3 rm s3://$S3_BUCKET/$S3_PREFIX/latest/ --recursive
fi

echo 'Uploading releaser binaries to S3...'
# Iterate over each file and upload it to S3.
for f in ${PREFIX}-*; do
  # Upload to the internal bucket.
  S3_LOCATION="s3://auklet-profiler/$ENVDIR/$S3_PREFIX/$VERSION/$f"
  aws s3 cp $f $S3_LOCATION
  # Upload to the public bucket for production builds.
  if [[ "$ENVDIR" == "production" ]]; then
    # Copy to the public versioned directory.
    VERSIONED_NAME="${f/$VERSION/$VERSION_SIMPLE}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/$VERSION_SIMPLE/$VERSIONED_NAME
    # Copy to the public "latest" directory.
    LATEST_NAME="${f/$VERSION/latest}"
    aws s3 cp $S3_LOCATION s3://$S3_BUCKET/$S3_PREFIX/latest/$LATEST_NAME
  fi
done

# Push to public GitHub repo.
# The hostname "aukletio.github.com" is intentional and it matches the "ssh-config-aukletio" file.
if [[ "$ENVDIR" == "production" ]]; then
  echo 'Pushing production branch to github.com/aukletio...'
  mv ~/.ssh/config ~/.ssh/config-bak
  cp .devops/ssh-config-aukletio ~/.ssh/config
  chmod 400 ~/.ssh/config
  git remote add aukletio git@aukletio.github.com:aukletio/Auklet-Releaser-C.git
  git push aukletio HEAD:master
  git remote rm aukletio
  rm -f ~/.ssh/config
  mv ~/.ssh/config-bak ~/.ssh/config
fi
