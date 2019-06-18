#!/bin/bash
set -e
VERSION=$5
cd $(dirname "$0")
cp ../$1 .
cp rpm-appsody.spec.tpl rpm-$1.spec
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "FOR LOCAL DEBUG ONLY..."
    sed -i "" "s/PACKAGE_NAME/$1/g" rpm-$1.spec
    sed -i "" "s/PACKAGE_VERSION/$VERSION/g" rpm-$1.spec
    sed -i "" "s+CONTROLLER_BASE_URL+$4+g" rpm-$1.spec
else
    echo "Travis only..."
    sed -i  "s/PACKAGE_NAME/$1/g" rpm-$1.spec
    sed -i  "s/PACKAGE_VERSION/$VERSION/g" rpm-$1.spec
    sed -i  "s+CONTROLLER_BASE_URL+$4+g" rpm-$1.spec  
fi
docker run -v $PWD:/sources -v $PWD:/output:Z $2:centos-7
# Get rid of the source RPM package
rm -f $1*.src.rpm
mv $1*.rpm $3/
rm -f $1 rpm-$1.spec
cd -