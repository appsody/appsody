# Dependencies
The Appsody CLI depends on a number of assets:
1. The `Appsody Controller`, which is available [here](https://github.com/appsody/controller). The controller version to pull is located in the `Makefile` as `CONTROLLER_VERSION` variable. The version is updated manually as needed, there is no automatic update when a new controller is release. However, the dependent controller version does need to be released first, as the build process for the Appsody CLI retrieves the controller binaries and does not build them. Once the dependent controller has been released, a Pull Request with the `CONTROLLER_VERSION` changes to the `Makefile` needs to be created and merged before continuing on with the CLI release.
1. The `debian-builder` Docker image, which is built separately from [this repo](https://github.com/appsody/debian-builder). The `Makefile` has a variable that points to the latest version of this image, called `DOCKER_IMAGE_DEB`. This image is used during the deploy stage to generate the Debian installer for Appsody. 
1. The `rpmbuilder` Docker image, which builds the RPM installer for Appsody. Currently, we use the image provided by `alectolytic`, and we point to it through the `Makefile` env var `DOCKER_IMAGE_RPM`.

We expect the Docker images for the Debian and RPM installers to change very rarely. The Appsody controller may be released more frequently. Generally, a new release of the Controller requires a new release of the Appsody CLI.

## Downstream dependencies
When the Appsody CLI is released, the build process also updates the Appsody website repository and the Appsody Homebrew formula, as we mentioned earlier. The update creates a new branch in those repos, and it is expected that the maintainers create new pull requests for those branches against the master branch. Keep in mind that the Appsody CLI release is complete only when those PRs are merged.

# Appsody CLI Release Process
Follow this process to create a new release of the Appsody CLI.

### Create a GitHub release
The Appsody CLI is made available by creating a tagged GitHub release
1. If there is a new dependent controller, a Pull Request with the `CONTROLLER_VERSION` changes to the `Makefile` needs to be created and merged before continuing
1. Navigate to https://github.com/appsody/appsody/releases
1. Click _Draft a new release_
1. Ensure the target branch is __master__
1. Define a tag in the format of x.y.z (example: 0.2.4). Use the tag also for the title.
1. Describe the release with your release notes, including a list of the features added by the release, a list of the major issues that are resolved by the release, caveats, known issues.
    * To see a comparison between master and the last release, navigate to https://github.com/appsody/appsody/compare/0.0.0...master replacing 0.0.0 with the last release.
1. Check the box for _This is a pre-release_
1. Click _Publish release_

### Monitor the build
1. Watch the [Travis build](https://travis-ci.com/appsody/appsody) for the release and ensure it passes. The build will include the `deploy` stage of the build process as defined in `.travis.yml`. The `deploy` stage, if successful, will produce the following results:
    * The release page will be populated with the build artifacts (see next step)
    * A new branch will be created in the homebrew-appsody repo (further steps below)
    * A new branch will be created in the website repo (further steps below)
1. Check the release artifacts to ensure these all exist (again, _x.y.z_ indicates the release - for example 0.2.4):
    * appsody-x.y.z-1.x86_64.rpm
    * appsody-x.y.z-darwin-amd64.tar.gz
    * appsody-x.y.z-linux-amd64.tar.gz
    * appsody-x.y.z-windows.tar.gz
    * appsody-homebrew-x.y.z.tar.gz
    * appsody.rb
    * appsody_x.y.z_amd64.deb
1. Edit the release and uncheck _This is a pre-relsease_ then click _Update relsease_

### Create a PR in the homebrew repo
1. Go to the [appsody/homebrew-appsody](https://github.com/appsody/homebrew-appsody/branches) repo and create a PR for the new Travis build branch.
1. Review and merge the PR.

### Create a PR in the website repo
1. Go to the [appsody/website](https://github.com/appsody/website/branches) repo and create a PR for the new Travis build branch.
1. Review and merge the PR.

# Release schedule
We plan to release the Appsody CLI at the end of each sprint - approximately every two weeks.


