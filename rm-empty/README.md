# rm-empty

## About this plugin

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install rm-empty

Installing a specific version:

`$ jfrog plugin install rm-empty@version`

Uninstalling a plugin

`$ jfrog plugin uninstall rm-empty

## Usage
### Commands
* folders 
    - Arguments:
        - path - A path in Artifactory, under which to remove all the empty folders.
    - Flags:
        - server-id: The Artifactory server ID configured using the config command.
    - Examples:
    ```
    $ jfrog rm-empty folders repository/path/in/rt/
  
    $ jfrog rm-empty folders repository/path/in/rt/ --quiet=true

    ```

### Environment variables
None.

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).
