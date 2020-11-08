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
        - timeUnit: The time unit of the maximal time. year, month, day are allowed values. **[Default: month]**
        - maximalTime: Artifacts that have not been downloaded for the past maximalTime will be deleted. **[Default: 1]**
        - maximalSize: Artifacts that are smaller than maximalSize (bytes) will not be deleted. **[Default: 0]**
    - Examples:
    ```
  $ jfrog rt-cleanup clean example-repo-local --timeUnit=day --maximalTime=3 --maximalSize=1000000

    Will delate all files that haven't been downloaded in the past 3 days and have size bigger than 1000000 bytes
    from the example-repo-local repository.
    ```

### Environment variables
None.

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).

