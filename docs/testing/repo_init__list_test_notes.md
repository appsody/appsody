# Appsody repo/list/init
- repo URLs:
    - https://raw.githubusercontent.com/appsody/stacks/master/index.yaml
    - https://github.com/appsody/stacks/releases/latest/download/experimental-index.yaml
    - https://github.com/appsody/stacks/releases/latest/download/incubator-index.yaml

- Here are the tests to run:
    - `appsody init <repo-name>/stack`
    - `appsody init stack`
    - `appsody list`
    - `appsody list <repo-name>`
    - `appsody add <repo-name>`
    - `appsody repo set-default <repo-name>`

- Test appsody list with old yaml and new repository.yaml
- Test appsody init with old yaml and new repository.yaml
- Test appsody repo add with old repository.yaml and new repository.yaml
- Test appsody completion
- Test appsody repo set-default
- Run through run, test, debug, build, deploy, stop, version
