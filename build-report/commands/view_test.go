package commands

import (
	"encoding/json"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-plugins/build-report/utils"
	"github.com/jfrog/jfrog-cli-plugins/build-report/utils/tests"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestBuildDetailsTableHasConstantLength(t *testing.T) {
	tests.LinesSameWidth = true
	buildInfo := entities.BuildInfo{
		Name:       "build-example",
		Number:     "5",
		Started:    "time",
		Principal:  "admin",
		Agent:      &entities.Agent{Name: "jfrog-cli-go", Version: "1.40.0"},
		BuildAgent: &entities.Agent{Name: "GENERIC"},
	}

	tw := &tests.TableWrapper{Table: &table.Table{}}
	fillBuildDetailsTable(tw, buildInfo)
	assert.True(t, tests.LinesSameWidth)
	tests.ClearWidth()
}

func TestBuildModulesTableHasConstantLength(t *testing.T) {
	tests.LinesSameWidth = true
	modules := []entities.Module{{
		Id: "ModuleId",
		Artifacts: []entities.Artifact{
			{
				Name: "art",
				Type: "json",
				Path: "/path/to/art",
				Checksum: entities.Checksum{
					Sha1: "abcd", Md5: "aaaa",
				},
			},
		},
		Dependencies: []entities.Dependency{
			{
				Id:   "dep",
				Type: "json",
				Checksum: entities.Checksum{
					Sha1: "abcd", Md5: "aaaa",
				},
			},
		},
	},
	}

	tw := &tests.TableWrapper{Table: &table.Table{}}
	fillBuildModulesTable(tw, modules)
	assert.True(t, tests.LinesSameWidth)
	tests.ClearWidth()
}

func TestModulesDiffTableHasConstantLength(t *testing.T) {
	tests.LinesSameWidth = true

	buildDiffJson, err := ioutil.ReadFile(filepath.Join("..", "testdata", "diff.json"))
	if err != nil {
		assert.NoError(t, err)
		return
	}

	var buildDiff utils.BuildDiff
	if err := json.Unmarshal(buildDiffJson, &buildDiff); err != nil {
		assert.NoError(t, err)
		return
	}

	tw := &tests.TableWrapper{Table: &table.Table{}}
	fillModulesDiffTable(tw, &buildDiff)
	assert.True(t, tests.LinesSameWidth)
	tests.ClearWidth()
}
