# rb-gen

## About this plugin

This plugin is designed to simplify interaction with release bundles, by
generating them from other formats. Currently, it can generate release bundles
from Helm v2 charts.

## Helm v3 support

Currently this plugin does not support Helm v3 with api v2 templates.

## Installation with JFrog CLI

Installing the latest version:

`$ jfrog plugin install rb-gen`

Installing a specific version:

`$ jfrog plugin install rb-gen@version`

Uninstalling a plugin

`$ jfrog plugin uninstall rb-gen`

## Usage

### Commands

- from-chart
  - Arguments:
    - bundle-name: The name of the new release bundle
    - bundle-version: The version of the new release bundle
  - Flags:
    - chart-path: The path (in Artifactory) of the Helm chart from which to
      generate the release bundle. All dependency Helm charts should be
      available in the same repository.
    - docker-repo: The name of a Docker repository in Artifactory. All
      dependency Docker images should be available in this repository.
  - Example:
    ``` shell
    jfrog rb-gen from-chart --user=<User> --apikey=<ApiKey> --url=<Artifactory Url> --dist-url=<Distribution Url> --chart-path=<chart path> --docker-repo=<Docker repo name> <bundle name> <bundle version>
    ```
  - Example assuming two virtual repos "helm" + "docker" containing all docker and helm artifacts.
    ``` shell 
    jfrog rb-gen from-chart --user=User --apikey=ApiKey --url=https://artifactory.example.com/artifactory --dist-url=https://artifactory.example.com/distribution --chart-path=helm/my-helm-chart-1.0.0.tgz --docker-repo=docker release-bundle-name 1.0.0
    ```
Note that if Helm or Docker dependencies are found in a remote repository, they
must be cached. Otherwise, they won't show up in the release bundle. After
generating a release bundle, the generator will output which dependencies were
and were not found; missing dependencies are not listed in the bundle.

## Release Notes
The release notes are available [here](RELEASE.md).