#!/bin/bash
set -e
cd $(dirname "$0")
VERSION=$4
FILE_NAME="appsody-$VERSION-windows.tar"
FILE_POSTFIX="windows"
CMD_NAME=$2
CMD_NAME=${CMD_NAME%.*}
wget $3/$CMD_NAME-controller
cp ../$2 .
cp ../LICENSE .
cp ../README.adoc .
tar cfz $FILE_NAME $2 $CMD_NAME-controller appsody-setup.bat LICENSE README.adoc

mv $FILE_NAME $1/

rm $2 $CMD_NAME-controller LICENSE README.adoc

cd -