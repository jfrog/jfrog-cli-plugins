# rm-empty

## About this plugin
This plugin can be used to remove all empty folders under a specified path in Artifactory.

## Installation with JFrog CLI
Installing the latest version:

`$ jf plugin install rm-empty`

Installing a specific version:

`$ jf plugin install rm-empty@version`

Uninstalling a plugin

`$ jf plugin uninstall rm-empty`

## Usage
### Commands
* folders / f
    - Arguments:
        - path - A path in Artifactory, under which to remove all the empty folders.
    - Flags:
        - server-id: The JFrog instance ID configured using the ```jf c add``` command. If not provided, the default configured instance is used.
        - quiet: Skip the delete confirmation message
    - Examples:
    ```
    $ jf rm-empty folders repository/path/in/rt/

    $ jf rm-empty folders repository/path/in/rt/ --quiet

    $ jf rm-empty f repository/path/in/rt/

    $ jf rm-empty f repository/path/in/rt/ --server-id my-server-id
    ```

### Environment variables
None.

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).
