#!/bin/bash
set -e
cd $(dirname "$0")
cp ../$1 ./$1.static
tar -cvf appsody-deb.tar ./debian
docker run -it -v $PWD:/input -v $PWD:/output -e CMD_NAME=$1 \
    -e VERSION=$5 -e CONTROLLER_BASE_URL=$4 $2
mv *.deb $3/
rm $1.static
rm *.tar
cd -