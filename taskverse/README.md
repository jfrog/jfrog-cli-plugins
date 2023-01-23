# taskverse

## About this plugin
This plugin allows developers to run Pipelines Tasks locally in their own machines.
It helps with the development process by considerably reducing the feedback loop.

## Installation with JFrog CLI

Installing the latest version:

`$ jf plugin install taskverse`

Installing a specific version:

`$ jf plugin install taskverse@version`

Uninstalling a plugin

`$ jf plugin uninstall taskverse`

## Usage

### Commands

* run
    - Arguments:
        - path_to_task - Path to the folder containing the task.yml file **[Required]**
        - arg - Arguments to the task in the format `--arg="key=value"`. Can be specified multiple times **[Optional]**
        - env - Environment variables in the format `--env="key=value"`. Can be specified multiple times **[Optional]**
    - Flags:
        - integrations - Path to file containing integrations to be added to the environment
        - on-step-complete - Enable onStepComplete hook execution **[Default: false]**
        - post-task-script - Path to a script to be executed after the task is done
    - Example:
```shell
→ jf taskverse run /home/user/jfrog-pipelines-task-hello-world --arg="name=user" --integrations ./integrations.json
13:54:40 [🔵Info] Parsing project integrations
13:54:40 [🔵Info] All build plane dependencies are available
13:54:40 [🔵Info] Docker dependency is available
13:54:40 [🔵Info] Assembling stepJson file
13:54:40 [🔵Info] Assembling steplet script
13:54:40 [🔵Info] Creating required folders and files at /Users/eliom/workspace/pip/src/taskverse/devenv/.taskverse
13:54:40 [🔵Info] Checking docker network
13:54:40 [🔵Info] Docker network taskverse is available
13:54:40 [🔵Info] Checking dind container
13:54:41 [🔵Info] Dind container taskverse-dind is running
13:54:41 [🔵Info] Checking for stale containers from previous runs
Error: No such container: taskverse-run
13:54:41 [🔵Info] Booting task container
16: Pulling from jfrog/pipelines-u20node
Digest: sha256:b8005e9273433f19aaa1b8cd36d301b43413bf300d1b44c43c6b3ffcd3c506b0
Status: Image is up to date for releases-docker.jfrog.io/jfrog/pipelines-u20node:16
Using JFrog CLI jf version 2.29.2

TASK => execution
pipe task --name /task --id task --arg name=user
node16 dist/index.js main
[Info] Hello user!
```

## Dependencies

-  Docker 20.x or newer

## Limitations

- This plugin does not work on Windows environments

## How does it work?

The plugin uses a Docker container to emulate the step environment and provide 
all the tools available to a step in a real build node. It has a lightweight
implementation of the build node agent that assembles a steplet script containing 
all required environmental configurations required by a step, including required 
files and folders, environment variables and utility functions. This script also 
contains the required commands to execute the task.

To emulate the step environment the plugin uses the Ubuntu 20 node 16 build image: 
releases-docker.jfrog.io/jfrog/pipelines-u20node:16. It also uses the reqKick image
to fetch some additional dependencies, including the utility functions script,
the JFrog cli and the pipe tool. The docker client dependency is downloaded from
https://download.docker.com.

A dind container is used to fully isolate the step environment from the host environment.
By doing that we allow tasks to use docker at full capacity without risking disrupting
the developer host machine.

The following docker components are created when using this plugin:
- `taskverse` docker network
- `taskverse-dind` dind docker container
- `taskverse-run` steplet docker container
- `taskverse-reqKick` temporary docker container used to fetch build plane dependencies

### Important Directories

The dependencies downloaded by the plugin are cached under `$JFROG_CLI_HOME/plugins/taskverse/resources/dependencies`

When running the plugin, the current working directory (the location where the 
plugin was executed) is mounted to the container as the current step working directory. 
Any changes made by the task to the current step working directory will be reflected
in the current working directory. Inside the container the current working directory
can be referenced via `/workdir`. This is specially useful when passing inputs or
setting environment variables used by the task that need to point to files or folders
in the current workspace.

Additional changes made by the task, including changes to task outputs, step
workspace, step cache, run state, pipelines state and so on are available at 
`.taskverse`. Developers can inspect the content generated there to make sure 
their tasks are working as expected.


