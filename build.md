
# Building from Source

Prerequisites:

* go version 1.12 or higher is installed and in the path
* docker is installed and running
* wget is installed

After setting the `GOPATH` env var correctly, just run `make <action...>` from the command line, within the same directory where `Makefile` resides. 

For example - if you want to obtain a fully built Appsody CLI for all the supported platform run:
```
make clean package
```
which will run the `clean` action, and then the `package` action. The latter produces the built binaries (as described in the next section).

If you inspect the `Makefile`, you'll notice that it invokes a number of scripts - and you'll notice a few conditional paths, because certain Linux commands behave differently on OS/X and elsewhere.



### What gets produced by the build?
Quite a bit of stuff. 

If you run `make clean package`, you will find the artifacts in the `package` sub-directory, under the folder that contains the `Makefile`. The same artifacts are published in the release page, when Appsody is released.

Here's a description of the various artifacts that are produced by the build:

* The actual RPM package for RHEL/Centos (to be `yum`med or `rpm -i`)

* The binaries tarred up in the way homebrew loves them

* The plain binaries tarred up as they come out of the build process

** for OS/X

** for Linux

** for Windows

* The homebrew Formula (which we should push to some git repo, once we go "public")

* The Debian package for Ubuntu-like Linux (to be `apt-get install`ed)

### Running the CLI
So, you built from source and you would like to run it. 

* The first thing you need to do is to extract the binary for your OS from the `./package` directory. Un-tar the file that matches your OS.

* Next, you need to copy that file to some place that's in your $PATH (for example, /usr/bin or /usr/local/bin) or %PATH% (any Win folder, then add it to the %PATH%). You may also want to call it `appsody`, so that you can run the CLI just by typing `appsody <command>`.

* If you are replacing an old installation, you may want to delete the appsody home directory (`$HOME/.appsody`).

### The Appsody Controller
In order to enable rapid app development, the Appsody CLI relies on a component called the `Appsody controller`, which is built separately, and which is meant to be run within the container that hosts the Appsody app itself (as opposed to being executed on the developers' system). However, the good news is that the Appsody CLI automatically downloads the Appsody controller when necessary, in the form of a Docker image.

When the Appsody CLI is built using the `Makefile`, the CLI "knows" which version of the controller needs to be obtained. If you just compile the Appsody CLI binary using `go build`, the Appsody CLI will pull the `latest` version of the controller image (which may or may not be a good match - so, be aware of that).

As a developer, you can also override which version of the Appsody controller the CLI uses, in two different ways:
1) By setting the `APPSODY_CONTROLLER_VERSION` env var prior to launching the Appsody CLI. For example, `export APPSODY_CONTROLLER_VERSION=0.3.0`
2) By setting the `APPSODY_CONTROLLER_IMAGE` env var prior to launching the Appsody CLI. For example, `export APPSODY_CONTROLLER_IMAGE=mydockeraccount/my-controller:1.0`.

If you specify both, `APPSODY_CONTROLLER_IMAGE` wins. 

Using these env vars, you can test the Appsody CLI with various levels of the controller binaries, which might be useful if you want to contribute to the Appsody project.

# Travis build
The project is instrumented with Travis CI and with an appropriate `Makefile`. Most of the build logic is triggered from within the `Makefile`.

When you push your code to GitHub, only the `test` and `lint` actions are executed by Travis.

In order for Travis to go all the way to `package` and `deploy`, you need to create a *new* release (one that is tagged with a never seen before tag). When you create a new release, a Travis build with automatically run, and the resulting artifacts will be posted on the `Releases` page. 
