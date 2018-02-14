// Setup the GitHub API connection.
const rp = require('request-promise');
const fs = require('fs');
const apiPrefix = 'https://api.github.com/';
const github = rp.defaults({
  baseUrl: apiPrefix,
  headers: {
    'User-Agent': 'esg-usa-bot',
    'Authorization': `Token ${process.env.CHANGELOG_GITHUB_TOKEN}`
  },
  json: true
});
// Grab inputs.
const org = process.env.CIRCLE_PROJECT_USERNAME;
const repo = process.env.CIRCLE_PROJECT_REPONAME;
const prNumber = process.env.CIRCLE_PR_NUMBER;
const branch = process.env.CIRCLE_BRANCH;
// If this is a PR, validate it. Otherwise, do nothing.
if (prNumber) {
  github.get({
    uri: `/repos/${org}/${repo}/pulls/${prNumber}`
  }).then(function(pr) {
    validatePr(pr);
  }).catch(handleError);
} else {
  // This might be a PR from another branch in the ESG-USA repo.
  github.get({
    uri: `/repos/${org}/${repo}/pulls?base=edge&head=${org}:${branch}`
  }).then(function(maybePr) {
    if (maybePr.length > 0) {
      validatePr(maybePr[0]);
    } else {
      console.log('Not a PR; nothing to do.');
    }
  }).catch(handleError);
}
function validatePr(pr) {
  console.log('Validating PR...');
  // Dump PR number to disk.
  var num = pr.number;
  console.log(`PR number: ${num}`);
  fs.writeFileSync('prnum.txt', `${num}`);
  // Make sure one of the required labels is set.
  var labels = pr.labels.map(function(label) { return label.name; });
  var hasRequiredLabel = labels.includes('breaking') || labels.includes('enhancement') || labels.includes('bug') || labels.includes('devops');
  if (!hasRequiredLabel) {
    console.log('ERROR: PR is missing a required label (breaking, enhancement, bug or devops).');
    process.exitCode = 1;
  }
}

// *** Utility functions ***
// Error handling function.
function handleError(err) {
  console.log(err);
  process.exitCode = 1;
}
