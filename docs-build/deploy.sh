#!/bin/bash
set -e
mkdir ./tmpdocclone
cd ./tmpdocclone
git clone git@github.com:${GH_ORG}/docs.git

cd docs
git checkout -b test${TRAVIS_BUILD_NUMBER}

cp ../../build/cli-commands.md ./docs/using-appsody/cli-commands.md

git add docs/using-appsody/cli-commands.md

git commit -m "Travis build: $TRAVIS_BUILD_NUMBER"

git remote add origin-2 https://${GH_TOKEN}@github.com/${GH_ORG}/docs.git
git push --set-upstream origin-2 test${TRAVIS_BUILD_NUMBER}
# clean up
cd ../..
rm -rf tmpdocclone

