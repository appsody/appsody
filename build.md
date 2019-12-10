
# Building locally from source

Prerequisites:

* go version 1.12 or higher is installed and in the path
* docker is installed and running
* wget is installed

After setting the `GOPATH` env var correctly, just run `make <action...>` from the command line, within the same directory where `Makefile` resides.

## What is supported by the makefile to build?

For the actual CLI binaries, three main target systems are supported:

* `darwin` (macOS)
* `linux`
* `windows`

For each of these you can set a component target to be:

* `build` the CLI binary, placing it in a sub-directory called "build". It will be named to match the target system architecture (e.g. `appsody-0.0.0-darwin-amd64`)
* `tar` up the binary from the build directory, along with the LICENSE and README files, and place the resulting tar file in a sub-directory called "package", again named to match the target system architecture
* create an *installable* package, in a form relevant for the given OS (i.e. `brew`, `deb` or `rpm`)

These core makefile targets take the form of:

{component}-{target system}

For instance, to build the CLI for macOS, you would enter:

```bash
make build-darwin
```

There are many other options including `all`, `package` (which builds, tars and packages everything) `build-docs` etc. For example, to clean, build and package all available targeted system binaries enter:

```bash
make clean package
```

After running this, you will find the artifacts in the `package` sub-directory, under the folder that contains the `Makefile`. The same artifacts are published in the release page, when Appsody is released.

Since the makefile uses the cross architecture capabilities of `go build`, as well as docker images to build appropriate installable packages for each architecture, you can build and package for all target systems from any one system.

The makefile supports a `help` target (which is, in fact, the default target), so to get the full list of targets simply enter:

```bash
make
```

## Running the locally-built CLI

As mentioned above, the `build` makefile target will produce a binary for the requested system, named accordingly. While you can just execute that locally from the `build` sub-directory, the makefile also supports an additional target that will copy this to a `bin` sub-directory and rename it `appsody`, for example to do this on macOS:

```bash
make build-darwin localbin-darwin
```

You can simply execute the newly build CLI by referencing it directly (`./bin/appsody`) or add this `bin` directory to your $PATH (or %PATH% if on Windows) and then just use the `appsody` command.

The Appsody CLI relies on a component called the `Appsody controller`, which is built separately, and which is meant to be run within the container that hosts the Appsody app itself (as opposed to being executed on the developers' system). However, the good news is that the Appsody CLI automatically downloads the Appsody controller when necessary, in the form of a Docker image.

If your CLI changes are not dependent on changes to the appsody controller, then there is nothing else you need to do to test locally. If your changes do need a matching appsody controller, as a developer, you can also override which version of the Appsody controller the CLI uses, in two different ways:

1) By setting the `APPSODY_CONTROLLER_VERSION` env var prior to launching the Appsody CLI. For example, `export APPSODY_CONTROLLER_VERSION=0.3.0`
2) By setting the `APPSODY_CONTROLLER_IMAGE` env var prior to launching the Appsody CLI. For example, `export APPSODY_CONTROLLER_IMAGE=mydockeraccount/my-controller:1.0`.

If you specify both, `APPSODY_CONTROLLER_IMAGE` wins.

Using these env vars, you can test the Appsody CLI with various levels of the controller binaries, which might be useful if you want to contribute to the Appsody project.

# Travis build within the project

The project is instrumented with Travis CI and with an appropriate `Makefile`. Most of the build logic is triggered from within the `Makefile`.

When you push your code to GitHub, only the `test` and `lint` actions are executed by Travis.

In order for Travis to go all the way to `package` and `deploy`, you need to create a *new* release (one that is tagged with a never seen before tag). When you create a new release, a Travis build with automatically run, and the resulting artifacts will be posted on the `Releases` page. 
