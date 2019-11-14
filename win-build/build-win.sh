#!/bin/bash
set -e
cd $(dirname "$0")
VERSION=$3
FILE_NAME="appsody-$VERSION-windows.tar.gz"
FILE_POSTFIX="windows"
CMD_NAME=$2
CMD_NAME=${CMD_NAME%.*}
#
cp ../$2 .
cp ../LICENSE .
cp ../README.md .
tar cfz $FILE_NAME $2 appsody-setup.bat LICENSE README.md

mv $FILE_NAME $1/

rm $2 LICENSE README.md

cd -