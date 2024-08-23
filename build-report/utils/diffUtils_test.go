package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/build-info-go/entities"
	"github.com/stretchr/testify/assert"
)

func TestBuildDiffStruct(t *testing.T) {
	buildDiffJson, err := os.ReadFile(filepath.Join("..", "testdata", "diff.json"))
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
		Artifact: entities.Artifact{
			Name: "one.more",
			Type: "more",
			Checksum: entities.Checksum{
				Sha1: "2fed359ef19c218d052b6ad0f8ac701a5a929030",
				Md5:  "7a4ceb07c7af56fbc520f335534714cd",
			}}}
	assert.Equal(t, buildDiff.Artifacts.New[0], expectedNewArt)

	expectedRemovedDep := DependencyDiff{Module: "buildreport", DiffId: "spec"}
	assert.Equal(t, buildDiff.Dependencies.Removed[0], expectedRemovedDep)
}
