#!/bin/bash
set -e
# When running CircleCI locally, don't do anything and use a dummy app version.
# This skips unnecessary behaviors in a local build that won't work.
NEW_VERSION=
if [[ "$CIRCLE_BUILD_NUM" == '' || "$CIRCLE_BRANCH" == 'HEAD' ]]; then
  echo 'This is a local CircleCI build.'
  NEW_VERSION="0.1.0-a.local.circleci.build"
else
  # Initialize.
  echo 'Initializing...'
  cd ~ # Prevents codebase contamination.
  rm -rf node_modules bug.txt enhancement.txt breaking.txt
  npm install --no-spin semver semver-extra > /dev/null
  gem install github_changelog_generator -v 1.14.3 > /dev/null
  # 1. Get all tags in the remote repo. Strip duplicate results for annotated tags.
  echo 'Getting latest production version...'
  TAGS=$(eval cd $CIRCLE_WORKING_DIRECTORY ; git ls-remote -q --tags | sed -E 's/[0-9a-f]{40}\trefs\/tags\/(.+)/\1/g;s/.+\^\{\}//g' | sed ':a;N;$!ba;s/\n/ /g')
  # 2. Get latest non-prerelease version, or default to 0.0.0.
  BASE_VERSION=$(node -e "var semver = require('semver-extra'); console.log(semver.maxStable(process.argv.slice(1)) || '0.0.0');" $TAGS)
  if [ "$BASE_VERSION" == '0.0.0' ]; then
    echo 'No production version yet (new codebase).'
    # No need to check changelogs/PRs, per SemVer.
    NEW_VERSION='0.1.0'
  else
    echo "Current production version: $BASE_VERSION"
    # 3. Generate three changelogs - bugs, bugs + enhancements, bugs + enhancements + breaking changes. Calculate their checksums.
    echo 'Calculating next version based on closed issues/PRs since the last production release...'
    github_changelog_generator --no-verbose --cache-file /tmp/github-changelog-http-cache-$RANDOM --cache-log /tmp/github-changelog-logger-$RANDOM.log --since-tag $BASE_VERSION --include-labels bug --no-compare-link --header-label '' --simple-list --date-format '' -o bug.txt "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME" > /dev/null
    github_changelog_generator --no-verbose --cache-file /tmp/github-changelog-http-cache-$RANDOM --cache-log /tmp/github-changelog-logger-$RANDOM.log --since-tag $BASE_VERSION --include-labels bug,enhancement --no-compare-link --header-label '' --simple-list --date-format '' -o enhancement.txt "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME" > /dev/null
    github_changelog_generator --no-verbose --cache-file /tmp/github-changelog-http-cache-$RANDOM --cache-log /tmp/github-changelog-logger-$RANDOM.log --since-tag $BASE_VERSION --include-labels bug,enhancement,breaking --no-compare-link --header-label '' --simple-list --date-format '' -o breaking.txt "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME" > /dev/null
    SHASUM_REGEX='s/([0-9a-f]{40}).*/\1/'
    BUG=$(sha1sum bug.txt | sed -E $SHASUM_REGEX)
    ENHANCEMENT=$(sha1sum enhancement.txt | sed -E $SHASUM_REGEX)
    BREAKING=$(sha1sum breaking.txt | sed -E $SHASUM_REGEX)
    # 4. Determine which version element to bump.
    if [ "$BUG" == "$ENHANCEMENT" ] && [ "$ENHANCEMENT" == "$BREAKING" ]; then
      MODE=patch
    elif [ "$ENHANCEMENT" == "$BREAKING" ]; then
      MODE=minor
    else
      MODE=major
    fi
    echo "Resulting version change: $MODE"
    # 5. Bump the version.
    NEW_VERSION=$(./node_modules/.bin/semver -i $MODE $BASE_VERSION)
  fi
  # 6. Add a prerelease identifier to the version if necessary.
  GIT_SHA=$(eval cd $CIRCLE_WORKING_DIRECTORY ; git rev-parse --short HEAD | xargs)
  if [ "$CIRCLE_BRANCH" == 'edge' ]; then
    echo 'This is a staging release.'
    NEW_VERSION="${NEW_VERSION}-beta.${CIRCLE_BUILD_NUM}+${GIT_SHA}"
  elif [ "$CIRCLE_BRANCH" == 'master' ]; then
    echo 'This is a QA release.'
    # Get the new RC version for this version (1 or greater).
    NEW_RC_VERSION=$(node -e "var semver = require('semver-extra'); var rcVersion = semver.maxPrerelease(process.argv.slice(1).filter(v => v.startsWith('$NEW_VERSION'))) || '.0'; rcVersion = rcVersion.substring(rcVersion.lastIndexOf('.') + 1); rcVersion = parseInt(rcVersion) + 1; console.log(rcVersion);" $TAGS)
    NEW_VERSION="${NEW_VERSION}-rc.${NEW_RC_VERSION}+${GIT_SHA}"
  elif [ "$CIRCLE_BRANCH" == 'production' ]; then
    echo 'This is a production release.'
    NEW_VERSION="${NEW_VERSION}+${GIT_SHA}"
  else
    # Assume this is a PR.
    PR_NUM=$CIRCLE_PR_NUMBER
    if [[ "$PR_NUM" == "" ]]; then
      # This might be a PR from another branch in the ESG-USA repo.
      # Use the GitHub API to extract the PR number, if an open PR exists.
      echo 'No PR number reported by CircleCI. Checking GitHub...'
      PR_NUM=$(curl -sS -H "User-Agent: esg-usa-bot" -H "Authorization: Token $CHANGELOG_GITHUB_TOKEN" "https://api.github.com/repos/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/pulls?base=edge&head=$CIRCLE_PROJECT_USERNAME:$CIRCLE_BRANCH" | jq '.[] | .number' | xargs)
    fi
    if [[ "$PR_NUM" == "" ]]; then
      echo 'No open PR from this branch to edge.'
      echo 'This is a branch build.'
      BRANCH_NAME=$(echo $CIRCLE_BRANCH | sed -E 's/[^a-zA-Z0-9]/\-/g')
      NEW_VERSION="${NEW_VERSION}-a.branch.${BRANCH_NAME}+${GIT_SHA}"
    else
      echo 'This is a PR build.'
      NEW_VERSION="${NEW_VERSION}-a.pr.${PR_NUM}+${GIT_SHA}"
    fi
  fi
# Done.
fi
echo "New codebase version: $NEW_VERSION"
echo $NEW_VERSION > ~/.version
rm -rf node_modules bug.txt enhancement.txt breaking.txt
