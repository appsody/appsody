# Starter Stack

A minimal, runnable starter stack that can be used as a base for developing new stacks. This stack provides a sample application template consisting of single line bash scripts, along with all the settings in Dockerfiles and manifests to allow you to use the Appsody CLI to test out init/run/debug/test/build/deploy ahead of starting to make changes to a copy of this stack for yourself.

To create a new stack called mystack based on this sample starter stack, enter, for example:

```bash
$ cd ~
$ appsody stack create mystack --copy incubator/starter
$ cd mystack
$ ls - al
total 16
drwxr-xr-x  6 henrynash  staff  192 21 Oct 00:14 .
drwxr-xr-x  3 henrynash  staff   96 21 Oct 00:14 ..
-rw-r--r--  1 henrynash  staff  621 21 Oct 00:14 README.md
drwxr-xr-x  7 henrynash  staff  224 21 Oct 00:14 image
-rw-r--r--  1 henrynash  staff  297 21 Oct 00:14 stack.yaml
drwxr-xr-x  3 henrynash  staff   96 21 Oct 00:14 templates
```

The initial version of your new stack is now ready to be packaged and made available to the Appsody CLI, by using `appsody stack package`.

> **Note** In general, Appsody will always try and look in the existing repositories first for stacks, and then in the local cache. For normal stack usage this is exactly what you want - however, when in the process of creating new stacks, by definition, the existing repositories will not yet know about your new stack. Hence it is quicker in this situation to tell Appsody to look in the local cache first. You can do this by setting the following environment variable: `export APPSODY_PULL_POLICY=IFNOTPRESENT`.

You can then check that the package works before you start to modify the stack structure to match the requirements of your new stack:

```bash
$ export APPSODY_PULL_POLICY=IFNOTPRESENT
$ appsody stack package
$ appsody list dev-local
REPO             ID       VERSION         TEMPLATES       DESCRIPTION
dev-local        mystack  0.1.0           *simple         sample stack to help...
```

Acting as an application developer, you can now use the CLI to initialize a project using this stack:

```bash
$ mkdir ~/test
$ cd ~/test
$ appsody init dev-local/mystack
$ appsody run
Running development environment...
Running command: docker[pull appsody/mystack:0.1]
Using local cache for image appsody/mystack:0.1
Running docker command: docker[run --rm -p 8080:8080 --name test20-dev -v /Users/henrynash/codewind-workspace/test20/:/project/userapp -v test20-deps:/project/deps -v /Users/henrynash/.appsody/appsody-controller:/appsody/appsody-controller -t --entrypoint /appsody/appsody-controller appsody/mystack:0.1 --mode=run]
[Container] Running: /bin/bash /project/userapp/hello.sh
[Container] Hello from Appsody!
```

For a tutorial, which describes in more detail how to create a new stack based on this starter, see [Creating a New Appsody Stack](https://developer.ibm.com/tutorials/create-appsody-stack/).

For any new stack, you should replace this README with one that explains your new stack in detail and any templates included, e.g.:

- How can a user start using this stack for application development, debugging etc.
- links to getting started guides etc.
