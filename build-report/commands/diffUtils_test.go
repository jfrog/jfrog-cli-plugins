package commands

import (
	"encoding/json"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jfrog/jfrog-cli-plugins/build-report/testUtils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestBuildDiffStruct(t *testing.T) {
	buildDiffJson, err := ioutil.ReadFile(filepath.Join("testdata", "diff.json"))
	if err != nil {
		assert.NoError(t, err)
		return
	}

	var buildDiff BuildDiff
	if err := json.Unmarshal(buildDiffJson, &buildDiff); err != nil {
		assert.NoError(t, err)
		return
	}
	assert.Len(t, buildDiff.Dependencies.Unchanged, 2)
	assert.Len(t, buildDiff.Artifacts.Updated, 1)
	assert.Len(t, buildDiff.Properties.Removed, 0)
	expectedNewArt := ArtifactDiff{Module: "buildreport",
		Artifact: buildinfo.Artifact{
			Name: "one.more",
			Type: "more",
			Checksum: &buildinfo.Checksum{
				Sha1: "2fed359ef19c218d052b6ad0f8ac701a5a929030",
				Md5:  "7a4ceb07c7af56fbc520f335534714cd",
			}}}
	assert.Equal(t, buildDiff.Artifacts.New[0], expectedNewArt)

	expectedRemovedDep := DependencyDiff{Module: "buildreport", DiffId: "spec"}
	assert.Equal(t, buildDiff.Dependencies.Removed[0], expectedRemovedDep)
}

func TestModulesDiffTableHasConstantLength(t *testing.T) {
	testUtils.LinesSameWidth = true

	buildDiffJson, err := ioutil.ReadFile(filepath.Join("testdata", "diff.json"))
	if err != nil {
		assert.NoError(t, err)
		return
	}

	var buildDiff BuildDiff
	if err := json.Unmarshal(buildDiffJson, &buildDiff); err != nil {
		assert.NoError(t, err)
		return
	}

	tw := &testUtils.TableWrapper{Table: &table.Table{}}
	fillModulesDiffTable(tw, &buildDiff)
	assert.True(t, testUtils.LinesSameWidth)
	testUtils.ClearWidth()
}
