#!/bin/bash
set -e
mkdir ./tmpdocclone
cd ./tmpdocclone
git clone https://${GH_TOKEN}@github.com/${GH_ORG}/website.git

cd website

set +e
diff ../../build/cli-commands.md ./content/docs/using-appsody/cli-commands.md
if [ $? -ne 0 ]
then
    set -e
    git checkout -b test${VERSION}
    cp ../../build/cli-commands.md ./content/docs/using-appsody/cli-commands.md

    git add content/docs/using-appsody/cli-commands.md

    git commit -m "Travis build: $VERSION" --author="Kyle G. Christianson <christik@us.ibm.com>"

    git push --set-upstream origin test${VERSION}
else
    echo "No changes were found in the appsody doc generation process between releases, no appsody/website branch will be created."
fi
# clean up
cd ../..
rm -rf tmpdocclone

 
