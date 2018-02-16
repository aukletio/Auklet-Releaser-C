#!/bin/bash
set -e
# Get the version and drop all build metadata from it.
VERSION=$(cat ~/.version | xargs | cut -f1 -d"+")
# Revert to a pristine checkout.
echo 'Reverting to pristine checkout...'
git reset --hard HEAD
git clean -qfdx || true
# Create release in JIRA.
JIRA_PROJECT=$(cat .esg | jq -r .jiraProject)
echo "Creating release in JIRA project '$JIRA_PROJECT'..."
JIRA_DATE=$(date +'%d/%b/%Y')
PAYLOAD="{\"name\": \"$CIRCLE_PROJECT_REPONAME $VERSION\", \"released\": true, \"userReleaseDate\": \"$JIRA_DATE\", \"project\": \"$JIRA_PROJECT\"}"
HEADERS_FILE="/tmp/jira-$RANDOM"
BODY_FILE="/tmp/jira-$RANDOM"
set +e
curl -D $HEADERS_FILE -sS -u "$JIRA_USERNAME:$JIRA_PASSWORD" -H "Content-Type: application/json" -X POST -d "$PAYLOAD" "$JIRA_URL/rest/api/2/version" > $BODY_FILE
RESULT=$?
set -e
cat -s $HEADERS_FILE
cat $BODY_FILE | jq .
if [[ "$RESULT" != "0" ]]; then
  exit $RESULT
fi
# Make and push the tag.
echo
echo 'Tagging release...'
git checkout $CIRCLE_BRANCH
git tag -a $VERSION -m "$VERSION"
git push origin $VERSION
TAG_SHA=$(git show-ref -s $VERSION | xargs) # This is NOT the same as the SHA-1 of the commit to which the tag points. Remember that the tag created above is annotated, not lightweight.
# Wait until the GitHub API recognizes the new tag.
# This ensures that GCG will see it.
echo
echo 'Waiting for new tag to appear in GitHub API...'
TAG_RESULT=''
while [ ! "$TAG_RESULT" == "$VERSION" ]; do
  sleep 5
  TAG_RESULT=$(curl -s -H "Authorization: Token $CHANGELOG_GITHUB_TOKEN" https://api.github.com/repos/$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME/git/tags/$TAG_SHA | jq -r .tag)
done
# Switch to the changelog branch.
echo 'Generating changelog updates...'
gem install github_changelog_generator -v 1.14.3 > /dev/null 2>&1
git checkout changelog
git branch -u origin/changelog changelog
git pull
# Generate the changelogs.
UNRELEASED_LABEL='Upcoming Changes'
github_changelog_generator --no-verbose --cache-file /tmp/github-changelog-http-cache-$RANDOM --cache-log /tmp/github-changelog-logger-$RANDOM.log --unreleased-label "$UNRELEASED_LABEL" --exclude-tags-regex '\d+\.\d+\.\d+-rc\.\d+' --date-format '%c %Z' --exclude-labels release -o README.md "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME"
github_changelog_generator --no-verbose --cache-file /tmp/github-changelog-http-cache-$RANDOM --cache-log /tmp/github-changelog-logger-$RANDOM.log --no-compare-link --date-format '%c %Z' --exclude-labels release -o README-WITH-RC.md "$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME"
# Modify the HEAD links in the non-RC changelog to point to "master".
sed -i -E "s/(## \[$UNRELEASED_LABEL\].+)\/HEAD\)/\1\/master)/g" README.md
sed -i -E 's/(\[Full Changelog\].+\.\.\.)HEAD\)/\1master)/g' README.md
# Shrink the headers of the RC tags in the RC changelog.
sed -i -E 's/## \[([[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+-.+)\]/### [\1]/g' README-WITH-RC.md
# If the topmost tag in the RC changelog is an RC tag, add an "Upcoming Changes" header.
perl -0777 -i -pe "s:# Change Log\s+### :# Change Log\n\n## $UNRELEASED_LABEL\n### :gsm" README-WITH-RC.md
# Collapse duplicate blank lines.
cat -s README.md > README.final
cat -s README-WITH-RC.md > README-WITH-RC.final
rm README.md README-WITH-RC.md
mv README.final README.md
mv README-WITH-RC.final README-WITH-RC.md
# Push the changelog to GitHub.
echo
echo 'Updating changelog and pushing to GitHub...'
git add README.md README-WITH-RC.md
git commit -m "$VERSION [skip ci]"
git push
