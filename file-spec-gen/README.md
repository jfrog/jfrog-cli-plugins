# file-spec-gen

## About this plugin
This plugin provides an easy way for generating file-specs json.

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin install file-spec-gen`

Installing a specific version:

`$ jfrog plugin install file-spec-gen@version`

Uninstalling a plugin

`$ jfrog plugin uninstall file-spec-gen`

## Usage
### Commands
* create
    - Arguments:
        No arguments
    - Flags:
        - --file: Output file path, if not provided the file-spec is outputted to the log. **[Optional]**
    - Examples:
    ```
  $ jfrog file-spec-gen create
  $ jfrog file-spec-gen create --file="/path/to/spec"
  ```

## Release Notes
The release notes are available [here](RELEASE.md).
