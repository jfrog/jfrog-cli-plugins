# file-spec-gen

## About this plugin
This plugin provides an easy way for generating file-specs json.

## Installation with JFrog CLI
Installing the latest version:

`$ jf plugin install file-spec-gen`

Installing a specific version:

`$ jf plugin install file-spec-gen@version`

Uninstalling a plugin

`$ jf plugin uninstall file-spec-gen`

## Usage
### Commands
* create
    - Arguments:
        No arguments
    - Flags:
        - --file: Output file path, if not provided the file-spec is outputted to the log. **[Optional]**
    - Examples:
    ```
  $ jf file-spec-gen create
  $ jf file-spec-gen create --file="/path/to/spec"
  ```

## Release Notes
The release notes are available [here](RELEASE.md).
