package commands

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jfrog/jfrog-cli-plugins/build-report/testUtils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildDetailsTableHasConstantLength(t *testing.T) {
	testUtils.LinesSameWidth = true
	buildInfo := buildinfo.BuildInfo{
		Name:                 "build-example",
		Number:               "5",
		Started:              "time",
		ArtifactoryPrincipal: "admin",
		Agent:                &buildinfo.Agent{Name: "jfrog-cli-go", Version: "1.40.0"},
		BuildAgent:           &buildinfo.Agent{Name: "GENERIC"},
	}

	tw := &testUtils.TableWrapper{Table: &table.Table{}}
	fillBuildDetailsTable(tw, buildInfo, "http://localhost:8082/artifactory/api/build/build-example/2")
	assert.True(t, testUtils.LinesSameWidth)
	testUtils.ClearWidth()
}

func TestBuildModulesTableHasConstantLength(t *testing.T) {
	testUtils.LinesSameWidth = true
	modules := []buildinfo.Module{{
		Id: "ModuleId",
		Artifacts: []buildinfo.Artifact{
			{
				Name: "art",
				Type: "json",
				Path: "/path/to/art",
				Checksum: &buildinfo.Checksum{
					Sha1: "abcd", Md5: "aaaa",
				},
			},
		},
		Dependencies: []buildinfo.Dependency{
			{
				Id:   "dep",
				Type: "json",
				Checksum: &buildinfo.Checksum{
					Sha1: "abcd", Md5: "aaaa",
				},
			},
		},
	},
	}

	tw := &testUtils.TableWrapper{Table: &table.Table{}}
	fillBuildModulesTable(tw, modules)
	assert.True(t, testUtils.LinesSameWidth)
	testUtils.ClearWidth()
}
