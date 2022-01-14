# build-deps-info

## About this plugin
The build-deps-info plugin prints the dependencies' details of a specific build, which has been previously published to Artifactory. For each dependency of the build, it shows the following information:

  1. Build name and build number of the original build creating this dependency.
  2. Git URL for the sources.

Note: If a specific dependency hasn't been published to Artifactory as an artifact of another build, the above details will not be available.

## Installation with JFrog CLI
Installing the latest version:

`$ jf plugin install build-deps-info`

Installing a specific version:

`$ jf plugin install build-deps-info@version`

Uninstalling a plugin

`$ jf plugin uninstall build-deps-info`

## Usage
### Commands
* show
    - Arguments:
        - build-name - The name of the build.
        - build-number - The number of the build.
    - Flags:
        - repo: Search in a specific repository **[Default: All Artifactory]**
        - server-id: Artifactory server ID configured using the config command **[Optional]**
    - Example:
    ```
  $ jf build-deps-info show my_build_name 1 --repo=maven-local
    +------------------------------------+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
    | MODULE ID                          | DEPENDENCY NAME                                   | BUILD         | VCS URL                                                                              |
    +------------------------------------+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
    | com.jfrog.cli:project:1.0-SNAPSHOT | com.jfrog.cli:MavenHelloWorldProject:1.0-SNAPSHOT | mvn-project/2 | https://github.com/jedib0t/go-pretty/commit/61333b7f82d34a6de2a8948538c227092431aee1 |
    +------------------------------------+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
  ```

## Release Notes
The release notes are available [here](RELEASE.md).
