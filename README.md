# Installation
## Mac OSX
Download https://github.com/appsody/appsody/releases/download/0.0.6/appsody.rb  
Download https://github.com/appsody/appsody/releases/download/0.0.6/appsody-homebrew-0.0.6.tar.gz
```
mv ~/Downloads/appsody-homebrew-0.0.6.tar.gz /tmp
brew install -f ~/Downloads/appsody.rb
```
To uninstall `brew uninstall appsody` then `rm -r ~/.appsody`
## Debian Linux
Download https://github.com/appsody/appsody/releases/download/0.0.6/appsody_0.0.6_amd64.deb
```
sudo apt install -f ~/Downloads/appsody_0.0.6_amd64.deb
```
To uninstall `sudo apt remove appsody` then `rm -r ~/.appsody`
## Redhat Linux
Download https://github.com/appsody/appsody/releases/download/0.0.6/appsody-0.0.6-1.x86_64.rpm
```
sudo yum install ~/Downloads/appsody-0.0.6-1.x86_64.rpm
```
To uninstall `sudo yum remove appsody` then `rm -r ~/.appsody`

## Windows
Download https://github.com/appsody/appsody/releases/download/0.0.6/appsody-0.0.6-windows.tar
```
tar -xvf appsody-0.0.6-windows.tar
appsody-setup.bat
```

# Basic Usage
The appsody cli supports the `help` command or `-h` for basic help and usage information.

Run `appsody list` to see the available stacks. To start a new project from a stack, create a directory and run `appsody init <stack>`.

```
mkdir myproject
cd myproject
appsody init nodejs-express
appsody run
```

`appsody run` will run the project in a development container linked to your local source code. With this running, you can edit code with your local IDE and watch appsody update the container automatically.


# Building from Source

## Travis build
The project is instrumented with Travis CI and with an appropriate `Makefile`. Most of the build logic is triggered from within the `Makefile`.

Upon commit, only the `test` and `lint` actions are executed by Travis.

In order for Travis to go all the way to `package` and `deploy`, you need to create a *new* release (one that is tagged with a never seen before tag). When you create a new release, a Travis build with automatically run, and the resulting artifacts will be posted on the `Releases` page. 

## Manual build
You can also test the build process manually.


Prerequisites:

* docker is installed and running
* wget is installed

After setting the `GOPATH` env var correctly, just run `make <action...>` from the command line, within the same directory where `Makefile` resides. For example `make package clean` will run the `package` and then the `clean` actions.


Some of the scripts have conditional paths, because certain Linux commands behave differently on OS/X and elsewhere (fun).

### What gets produced by the build?
Quite a bit of stuff. 

Here's a description of the various artifacts as you would see them in a release page:

* The actual RPM package for RHEL/Centos (to be `yum`med or `rpm -i`)

* The binaries tarred up in the way homebrew loves them

* The plain binaries tarred up as they come out of the build process

** for OS/X

** for Linux

** for Windows

* The homebrew Formula (which we should push to some git repo, once we go "public")

* The Debian package for Ubuntu-like Linux (to be `apt-get install`ed)

* Some other stuff that's always there

### Running the CLI
So, you built from source and you would like to run it. 

* The first thing you need to do is to extract the binary for your OS from the `./package` directory. Un-tar the file that matches your OS.

* Next, you need to copy that file to some place that's in your $PATH (for example, /usr/bin or /usr/local/bin) or %PATH% (any Win folder, then add it to the %PATH%). You may also want to call it `appsody`, so that you can run the CLI just by typing `appsody <command>`.

* Next, you need to build the `appsody-controller` (sorry). It's here: https://github.com/appsody/appsody-controller. Just build the binary, and move it to the *same directory where you put `appsody`*. Call it `appsody-controller` (mandatory!) and make sure that it has +x permissions (also mandatory).

* If you are replacing an old installation, you may want to delete the appsody home directory (`$HOME/.appsody`).