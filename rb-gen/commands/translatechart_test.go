package commands

import (
	"k8s.io/helm/pkg/chartutil"
	"testing"
)

func TestHelmToFilespec1(t *testing.T) {
	expected := "{\"files\":[{\"pattern\":\"testhelmrepo/acs-engine-autoscaler-2.2.2.tgz\"},{\"pattern\":\"testhelmrepo/*/acs-engine-autoscaler-2.2.2.tgz\"}]}"
	chrt, err := chartutil.Load("testdata/acs-engine-autoscaler-2.2.2.tgz")
	if err != nil {
		t.Fatalf("Error loading test chart: %s\n", err)
	}
	spec, _, err := createFilespec(chrt, "testhelmrepo", "testdockerrepo", "")
	if err != nil {
		t.Fatalf("Error generating filespec: %s\n", err)
	}
	if spec != expected {
		t.Fatalf("Generated spec is incorrect. Expected:\n%s\nGot:\n%s\n", expected, spec)
	}
}

func TestHelmToFilespec2(t *testing.T) {
	expected := "{\"files\":[{\"pattern\":\"testdockerrepo/alpine/3.10/\"},{\"pattern\":\"testdockerrepo/*/alpine/3.10/\"},{\"pattern\":\"testdockerrepo/bitnami/postgresql/9.6.17-debian-10-r21/\"},{\"pattern\":\"testdockerrepo/*/bitnami/postgresql/9.6.17-debian-10-r21/\"},{\"pattern\":\"testdockerrepo/jfrog/artifactory-jcr/7.4.1/\"},{\"pattern\":\"testdockerrepo/*/jfrog/artifactory-jcr/7.4.1/\"},{\"pattern\":\"testdockerrepo/jfrog/nginx-artifactory-pro/7.4.1/\"},{\"pattern\":\"testdockerrepo/*/jfrog/nginx-artifactory-pro/7.4.1/\"},{\"pattern\":\"testhelmrepo/artifactory-9.4.0.tgz\"},{\"pattern\":\"testhelmrepo/*/artifactory-9.4.0.tgz\"},{\"pattern\":\"testhelmrepo/artifactory-jcr-2.2.0.tgz\"},{\"pattern\":\"testhelmrepo/*/artifactory-jcr-2.2.0.tgz\"},{\"pattern\":\"testhelmrepo/postgresql-8.7.3.tgz\"},{\"pattern\":\"testhelmrepo/*/postgresql-8.7.3.tgz\"}]}"
	chrt, err := chartutil.Load("testdata/artifactory-jcr-2.2.0.tgz")
	if err != nil {
		t.Fatalf("Error loading test chart: %s\n", err)
	}
	spec, _, err := createFilespec(chrt, "testhelmrepo", "testdockerrepo", "")
	if err != nil {
		t.Fatalf("Error generating filespec: %s\n", err)
	}
	if spec != expected {
		t.Fatalf("Generated spec is incorrect. Expected:\n%s\nGot:\n%s\n", expected, spec)
	}
}

func TestHelmToFilespecWithoutValues(t *testing.T) {
	expected := "{\"files\":[{\"pattern\":\"testdockerrepo/main-image/v0.0.1/\"},{\"pattern\":\"testdockerrepo/*/main-image/v0.0.1/\"},{\"pattern\":\"testhelmrepo/values-test-0.1.0.tgz\"},{\"pattern\":\"testhelmrepo/*/values-test-0.1.0.tgz\"}]}"
	chrt, err := chartutil.Load("testdata/values-test-0.1.0.tgz")
	if err != nil {
		t.Fatalf("Error loading test chart: %s\n", err)
	}
	spec, _, err := createFilespec(chrt, "testhelmrepo", "testdockerrepo", "")
	if err != nil {
		t.Fatalf("Error generating filespec: %s\n", err)
	}
	if spec != expected {
		t.Fatalf("Generated spec is incorrect. Expected:\n%s\nGot:\n%s\n", expected, spec)
	}
}

func TestHelmToFilespecWithValues(t *testing.T) {
	expected := "{\"files\":[{\"pattern\":\"testdockerrepo/extra-container/v0.0.1/\"},{\"pattern\":\"testdockerrepo/*/extra-container/v0.0.1/\"},{\"pattern\":\"testdockerrepo/main-image/v0.0.1/\"},{\"pattern\":\"testdockerrepo/*/main-image/v0.0.1/\"},{\"pattern\":\"testhelmrepo/values-test-0.1.0.tgz\"},{\"pattern\":\"testhelmrepo/*/values-test-0.1.0.tgz\"}]}"
	chrt, err := chartutil.Load("testdata/values-test-0.1.0.tgz")
	if err != nil {
		t.Fatalf("Error loading test chart: %s\n", err)
	}
	spec, _, err := createFilespec(chrt, "testhelmrepo", "testdockerrepo", "testdata/values-test.yaml")
	if err != nil {
		t.Fatalf("Error generating filespec: %s\n", err)
	}
	if spec != expected {
		t.Fatalf("Generated spec is incorrect. Expected:\n%s\nGot:\n%s\n", expected, spec)
	}
}
