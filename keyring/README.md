# keyring

## About this plugin
This plugin allows the following:
* Storing the connection details of multiple Artifactory instances in the OS keyring.
* Using the stored connection details for executing any Artifactory JFrog CLI command.
* Deleting the stored connection details from the OS keyring.

Supports Linux, Windows and MacOS.

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install keyring`

Installing a specific version:

`$ jfrog plugin install keyring@version`

Uninstalling a plugin

`$ jfrog plugin uninstall keyring`

## Usage
### Commands
* Store the Artifactory connection details is the OS keyring.
    - Name: store
    - Alias: s
    - Arguments:
        No arguments
    - Flags:
        - --server-id: Artifactory server ID used for accessing the configuration. **[Mandatory]**
        - --url: Artifactory URL. **[Mandatory]**
        - --user: Artifactory username. **[Mandatory]**
        - --password: Artifactory password. **[Mandatory]**
        
    - Example:
    ```
  $ jfrog keyring store --server-id Artifactory-1 --url https://my-rt.com/artifactory --user my-user --password my-password
  ```
* Use the stored Artifactory connection details for executing any Artifactory JFrog CLI command.
    - Name: use
    - Alias: u
    - Arguments:
        Stored server ID
    - Flags:
        No flags
        
    - Examples:
    ```
  $ jfrog keyring use Artifactory-1 search "generic-local/path/"
  $ jfrog keyring use Artifactory-1 download "generic-local/path/" --flat --recursive
  ```
* Delete Artifactory connection details from the OS keyring.
    - Name: delete
    - Alias: del
    - Arguments:
        Stored server ID
    - Flags:
        No flags
        
    - Example:
    ```
  $ jfrog keyring del Artifactory-1
  ```

## Release Notes
The release notes are available [here](RELEASE.md).
