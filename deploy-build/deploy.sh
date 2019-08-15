#!/bin/bash
set -e
mkdir ./tmpclone
cd ./tmpclone
git clone https://${GH_TOKEN}@github.com/${GH_ORG}/homebrew-appsody.git

cd homebrew-appsody
git checkout -b test${VERSION}
cp ../../package/appsody.rb .
git add appsody.rb
git commit -m "Travis build: $VERSION" --author="Kyle G. Christianson <christik@us.ibm.com>"
git push --set-upstream origin test${VERSION}
# clean up
cd ../..
rm -rf tmpclone