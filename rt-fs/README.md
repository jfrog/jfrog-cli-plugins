# rt-fs

## About this plugin
This plugin executes file system commands in Artifactory. It designed to mimic the functionality of the Linux/Unix 'ls' and 'cat' commands.

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install rt-fs`

Installing a specific version:

`$ jfrog plugin install rt-fs@version`

Uninstalling a plugin

`$ jfrog plugin uninstall rt-fs`

## Usage
### Commands
* ls
    - Arguments:
        - path - Path in Artifactory.
    - Flags:
        - server-id: Artifactory server ID configured using the config command **[Optional]**
    - Example:
    ```
        $ jfrog rt-fs ls generic-local
        file_name1.zip   file_name2.zip   file_name3.zip
    ```
* cat
    - Arguments:
        - path - Path in Artifactory.
    - Flags:
        - server-id: Artifactory server ID configured using the config command **[Optional]**
    - Example:
    ```
        $ jfrog rt-fs cat generic-local/file_name1.txt
        Hello world
    ```

## Additional info
Files are displayed in white color.
Folders are displayed in blue color.

## Release Notes
The release notes are available [here](RELEASE.md).
