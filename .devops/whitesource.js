const { exec } = require('child_process');
const https = require('follow-redirects').https;
// Prepare variables.
const projectName = process.env.WHITESOURCE_PROJECT_NAME;
const projectVersion = process.env.WHITESOURCE_PROJECT_VERSION;
const repoDir = process.argv[2];
var deps = [], depList = [], depsLength = 0;
var outputJson = {
  coordinates: {
    artifactId: projectName,
    version: projectVersion
  },
  dependencies: []
};
// Get all the Golang dependencies for this project.
console.log('Retrieving Go dependencies and assembling WhiteSource API payload...');
exec('dep status -json', { cwd: repoDir }, (error, stdout, stderr) => {
  if (error) {
    console.error(error);
    process.exitCode = 1;
    return;
  } else if (stderr) {
    console.error(stderr);
    process.exitCode = 1;
    return;
  }
  // Clean up the output and iterate.
  deps = JSON.parse(stdout);
  depsLength = deps.length;
  if (deps.length === 0) {
    console.log('No dependencies; nothing to do.');
    return;
  }
  deps.forEach(function(dep) {
    // Get the project name.
    var name = dep.ProjectRoot;
    // Get the project owner.
    var owner = name.substring(name.indexOf('/') + 1);
    owner = owner.substring(0, owner.indexOf('/'));
    // Determine the revision and version.
    var revision = dep.Revision;
    var version = dep.Version;
    if (version.indexOf('branch ') === 0) {
      version = null;
    }
    // Add the dependency to our final list.
    console.log(`${name} ${revision} (Version: ${version})`);
    // This schema comes from a request file created by FSA.
    newDep = {
      groupId: owner,
      artifactId: name,
      systemPath: `${repoDir}/Gopkg.lock`,
      optional: false,
      children: [],
      exclusions: [],
      licenses: [],
      copyrights: [],
      dependencyType: 'GO',
      checksums: {},
      commit: revision
    };
    if (version) newDep.version = version;
    depList.push(newDep);
  });
  submitToWhitesource();
});

function submitToWhitesource() {
  outputJson.dependencies = depList;
  console.log('');
  console.log('Submitting payload to WhiteSource...');
  var req = https.request({
    method: 'POST',
    host: 'saas.whitesourcesoftware.com',
    path: '/agent',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
      Charset: 'utf-8'
    }
  }, function(response) {
    var resp = '';
    response.on('data', (chunk) => { resp += chunk; });
    response.on('end', () => {
      try {
        resp = JSON.parse(resp);
        try {
          resp.data = JSON.parse(resp.data);
        } catch (de) {}
        console.log(JSON.stringify(resp, null, 2));
      } catch (e) {
        console.error(e);
        console.log(resp);
        process.exitCode = 1;
      }
    });
  });
  req.on('error', (e) => {
    console.error(e);
    process.exitCode = 1;
  });
  // Arguments here are designed to spoof the FSA.
  req.write(`type=UPDATE&agent=fs-agent&agentVersion=2.7.0&pluginVersion=18.5.1&token=${process.env.WHITESOURCE_ORG_TOKEN}&product=${process.env.WHITESOURCE_PRODUCT_TOKEN}&timeStamp=${Date.now()}&diff=[${JSON.stringify(outputJson)}]`);
  req.end();
}
