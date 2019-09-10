#!/bin/bash
set -e
mkdir ./tmpclone
cd ./tmpclone
git clone https://${GH_TOKEN}@github.com/${GH_ORG}/homebrew-appsody.git

cd homebrew-appsody
git checkout -b test${TRAVIS_TAG}${TRAVIS_BUILD_NUMBER}
cp ../../package/appsody.rb .
git add appsody.rb
git commit -m "Travis build: $TRAVIS_TAG$TRAVIS_BUILD_NUMBER" --author="Appsody Devops <appsodydev@gmail.com>"
git push --set-upstream origin test${TRAVIS_TAG}${TRAVIS_BUILD_NUMBER}
# clean up
cd ../..
rm -rf tmpclone