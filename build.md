
# Building locally from source

Prerequisites:

* Go version 1.12 or higher is installed and in the path
* Docker is installed and running
* wget is installed

Set the `GOPATH` environment variable and run `make <action...>` from the command line, within the same directory as the `Makefile`.

## What is supported by the makefile to build?

For the Appsody CLI binaries, three main target systems are supported:

* `darwin` (macOS)
* `linux`
* `windows`

For each system you can set a component target to:

* `build` to build the CLI binary and place it in a subdirectory called `build`. The binary file is named to match the target system architecture, for example `appsody-0.0.0-darwin-amd64`
* `tar` to compress the CLI binary from the build directory, along with the LICENSE and README files, and place the resulting tar file in a subdirectory called `package`, again named to match the target system architecture
* `brew`, `deb` or `rpm` to create an *installable* CLI package, in a form relevant for the target system architecture

These core makefile targets take the form of:

{component}-{target system}

For instance, to build the CLI for macOS, you would enter:

```bash
make build-darwin
```

There are many other options that include `all`, `package` (which builds, tars, and packages everything), `build-docs` etc. For example, to clean, build, and package all available targeted system binaries enter:

```bash
make clean package
```

The command generates artifacts in the `package` subdirectory, within the folder that contains the makefile. When Appsody is released, the same artifacts are published in the [Appsody releases page](https://github.com/appsody/appsody/releases).

Since the makefile uses the cross architecture capabilities of `go build`, as well as Docker images to build appropriate installable packages for each architecture, you can build and package all target system architectures from any one system architecture.

The makefile supports a `help` target (which is, in fact, the default target), so to get the full list of targets enter:

```bash
make
```

## Running the locally built CLI

The `build` makefile target produces a named binary for the requested system. You can run the binary locally from the `build` subdirectory. The makefile also provides a target `localbin-{target system}` that copies the binary to a subdirectory called `bin` and renames it to `appsody`. For instance, on macOS:

```bash
make build-darwin localbin-darwin
```

You can run the newly built CLI by referencing it directly (`./bin/appsody`). Alternatively you can add the `bin` directory to your $PATH (or %PATH% on Windows) and use the `appsody` command.

The Appsody CLI relies on a component that is called the [Appsody controller](https://github.com/appsody/controller). The Appsody Controller is built separately, and is meant to be run within the container that hosts the Appsody app itself (as opposed to being executed on the developer's system). The Appsody CLI automatically downloads the Appsody controller, when necessary, in the form of a Docker image.

If your CLI changes are not dependent on changes to the Appsody Controller, then there is nothing else that you need to do to test locally. If your code changes need a matching Appsody Controller, as a developer, you can override which version of the Appsody Controller the CLI uses, in two different ways:

1) Set the `APPSODY_CONTROLLER_VERSION` environment variable before you run the Appsody CLI. For example, `export APPSODY_CONTROLLER_VERSION=0.3.0`.
2) Set the `APPSODY_CONTROLLER_IMAGE` environment variable before you run the Appsody CLI. For example, `export APPSODY_CONTROLLER_IMAGE=mydockeraccount/my-controller:1.0`.

If you specify both, `APPSODY_CONTROLLER_IMAGE` wins.

Using these environment variables, you can test the Appsody CLI with various levels of the Appsody Controller binaries, which might be useful if you want to contribute to the Appsody project.

# Travis build within the project

The Appsody project is instrumented with Travis CI and an appropriate makefile. Most of the build logic is triggered from within the makefile.

When you push your code to GitHub, only the `test` and `lint` actions are executed by Travis.

In order for Travis to go all the way to `package` and `deploy`, you need to create a *new* release (one that is tagged with a never seen before tag). When you create a new release, a Travis build is automatically run, and the resulting artifacts are posted on the `Releases` page.
