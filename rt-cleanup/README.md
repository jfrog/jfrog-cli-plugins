# rt-cleanup

## About this plugin
This plugin is a simple Artifactory cleanup plugin.
It can be used to delete all artifacts that have not been downloaded for the past n time units (both can bu configured)
from a given repository.

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install rt-cleanup`

Installing a specific version:

`$ jfrog plugin install rt-cleanup@version`

Uninstalling a plugin

`$ jfrog plugin uninstall rt-cleanup`

## Usage
### Commands
* clean 
    - Arguments:
        - repository - The name of the repository you would like to clean.
    - Flags:
        - server-id: The Artifactory server ID configured using the config command.
        - time-unit: The time unit of the no-dl time. year, month and day are the allowed values. **[Default: month]**
        - no-dl: Artifacts that have not been downloaded for at least no-dl will be deleted.. **[Default: 1]**
    - Examples:
    ```
    $ jfrog rt-cleanup clean example-repo-local --time-unit=day --no-dl=3

    ```

### Environment variables
None.

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).