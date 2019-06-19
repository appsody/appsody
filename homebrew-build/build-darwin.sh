#!/bin/bash
set -e
cd $(dirname "$0")
VERSION=$4
FILE_PREFIX="appsody-homebrew"
REPO_NAME=${TRAVIS_REPO_SLUG}
RELEASE_TAG=${TRAVIS_TAG}
echo $VERSION

cp appsody-formula-template.rb ./$2.rb
wget $3/$2-controller
cp ../$2 .
tar -cvzf $FILE_PREFIX-$VERSION.tar.gz $2 $2-controller
SHA_256=`shasum -a 256 $FILE_PREFIX-$VERSION.tar.gz | cut -c -64`
if [[ "$OSTYPE" == "darwin"* ]]; then
echo "FOR LOCAL DEBUG ONLY..."
sed -i "" "s/VERSION_NUMBER/$VERSION/g" $2.rb
sed -i "" "s/SHA_256/$SHA_256/g" $2.rb
sed -i "" "s/FILE_PREFIX/$FILE_PREFIX/g" $2.rb
sed -i "" "s/REPO_NAME/$REPO_NAME/g" $2.rb
sed -i "" "s/RELEASE_TAG/$RELEASE_TAG/g" $2.rb
else
echo "Travis only..."
sed -i "s/VERSION_NUMBER/$VERSION/g" $2.rb
sed -i "s/SHA_256/$SHA_256/g" $2.rb
sed -i "s/FILE_PREFIX/$FILE_PREFIX/g" $2.rb
sed -i "s/REPO_NAME/$REPO_NAME/g" $2.rb
sed -i "s/RELEASE_TAG/$RELEASE_TAG/g" $2.rb   
fi

mv $FILE_PREFIX-$VERSION.tar.gz $1/
mv ./$2.rb $1/
rm $2 $2-controller

cd -