# release-bundle-generator

## About this plugin

This plugin is designed to simplify interaction with release bundles, by
generating them from other formats. Currently, it can generate release bundles
from Helm charts.

## Installation with JFrog CLI

Installing the latest version:

`$ jfrog plugin install release-bundle-generator`

Installing a specific version:

`$ jfrog plugin install release-bundle-generator@version`

Uninstalling a plugin

`$ jfrog plugin uninstall release-bundle-generator`

## Usage

To generate a release bundle from a chart, you can run:

``` shell
jfrog from-chart --chart-path=<chart path> --docker-repo=<Docker repo name> <bundle name> <bundle version>
```

This has the following parameters:
- The path (in Artifactory) of the Helm chart from which to generate the release
  bundle. All dependency Helm charts should be available in the same repository.
- The name of a Docker repository in Artifactory. All dependency Docker images
  should be available in this repository.
- The name and version that the new release bundle should have.
- All optional arguments can be checked by running `jfrog from-chart --help`

Note that if Helm or Docker dependencies are found in a remote repository, they
must be cached. Otherwise, they won't show up in the release bundle. After
generating a release bundle, the generator will output which dependencies were
and were not found; missing dependencies are not listed in the bundle.
