#!/bin/bash
set -e
mkdir ./tmpclone
cd ./tmpclone
git clone https://github.com/${GH_ORG}/homebrew-appsody.git
cd homebrew-appsody
git checkout -b test${TRAVIS_BUILD_NUMBER}
cp ../../package/appsody.rb .
git add appsody.rb
git commit -m "Travis build: $TRAVIS_BUILD_NUMBER"
git remote add origin-2 https://${GH_TOKEN}@github.com/${GH_ORG}/homebrew-appsody.git
git push --set-upstream origin-2 test${TRAVIS_BUILD_NUMBER}
# clean up
cd ../..
rm -rf tmpclone