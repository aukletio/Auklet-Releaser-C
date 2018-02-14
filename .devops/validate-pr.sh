#!/bin/bash
set -e
THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# When running CircleCI locally, don't do anything.
if [[ "$CIRCLE_BUILD_NUM" != '' && "$CIRCLE_BRANCH" != 'HEAD' ]]; then
  echo 'Initializing...'
  cd ~ # Prevents codebase contamination.
  rm -rf node_modules prnum.txt
  npm install --no-spin request request-promise > /dev/null 2>&1
  node $THIS_DIR/validatePr.js
# Done.
fi
