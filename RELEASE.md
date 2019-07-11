# How to make this asset available
The Appsody CLI is made available by creating a tagged GitHub release:
* Go to the _Releases_ page of the repo
* Click _Draft a new release_
* Define a tag in the format of x.y.z (example: 0.2.4). Use the tag also for the title.
* Describe the release with your release notes, including a list of the features added by the release, a list of the major issues that are resolved by the release, caveats, known issues.
* Click _Publish release_

These steps will trigger the `deploy` stage of the build process, as defined in `.travis.yml`. The `deploy` stage, if successful, will produce the following results:
* The release page will be populated with the build artifacts (installers, and binary tar files)
* A new branch will be created in the [appsody/homebrew-appsody repo](https://github.com/appsody/homebrew-appsody), containing the updated Homebrew formula. The branch is named after the travis build number. You will then have to create a pull request for that branch in that repository to have the formula released.
* A new branch will be created in the [appsody/website repo](https://github.com/appsody/website) with the newly generated CLI documentation. The branch is named after the travis build number. You will then have to create a pull request for that branch in that repository to have the new docs released.

# Release schedule
We plan to release the Appsody CLI at the end of each sprint - approximately every two weeks.

# Dependencies
The Appsody CLI depends on a number of assets:
1) The `Appsody Controller`, which is available [here](https://github.com/appsody/controller). This component needs to be released first, as the build process for the Appsody CLI retrieves the controller binaries. In the `Makefile`, you find the `CONTROLLER_BASE_URL` variable, which determines the version of the controller that gets pulled by the build process.
1) The `debian-builder` Docker image, which is built separately from [this repo](https://github.com/appsody/debian-builder). The `Makefile` has a variable that points to the latest version of this image, called `DOCKER_IMAGE_DEB`. This image is used during the deploy stage to generate the Debian installer for Appsody. 
1) The `rpmbuilder` Docker image, which builds the RPM installer for Appsody. Currently, we use the image provided by `alectolytic`, and we point to it through the `Makefile` env var `DOCKER_IMAGE_RPM`.

We expect the Docker images for the Debian and RPM installers to change very rarely. The Appsody controller may be released more frequently. Generally, a new release of the Controller requires a new release of the Appsody CLI.

## Downstream dependencies
When the Appsody CLI is released, the build process also updates the Appsody website repository and the Appsody Homebrew formula, as we mentioned earlier. The update creates a new branch in those repos, and it is expected that the maintainers create new pull requests for those branches against the master branch. Keep in mind that the Appsody CLI release is complete only when those PRs are merged.