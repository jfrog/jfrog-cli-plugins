# hello-frog

## About this plugin
Build-deps-info plugin print dependencies details of a specific build name & build number in Artifactory which includes,
dependency's build-name/build-number and a link to the vcs(source code).

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install build-deps-info`

Installing a specific version:

`$ jfrog plugin install build-deps-info@version`

Uninstalling a plugin

`$ jfrog plugin uninstall build-deps-info`

## Usage
### Commands
* print
    - arguments:
        - build-name - The name of the build.
        - build-number - The number of the build.
    - flags:
        - repo: Search the dependencies' build in a specific folder **[Default: All Artifactory]**
    - example:
    ```
  $ jfrog build-deps-info print  my_build_name 1 --repo=maven-local
    +---+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
    | # | DEPENDENCY NAME                                   | BUILD         | VCS URL                                                                              |
    +---+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
    |   | com.jfrog.cli:MavenHelloWorldProject:1.0-SNAPSHOT | mvn-project/2 | https://github.com/jedib0t/go-pretty/commit/61333b7f82d34a6de2a8948538c227092431aee1 |
    +---+---------------------------------------------------+---------------+--------------------------------------------------------------------------------------+
  ```

### Environment variables
None.

## Additional info
None.

## Release Notes
The release notes are available [here](RELEASE.md).
