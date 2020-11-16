package builddepsinfo

import (
	"testing"

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/stretchr/testify/assert"
)

type ServiceManagerMock struct {
	artifactory.EmptyArtifactoryServicesManager
}

func (smm *ServiceManagerMock) SearchFiles(params services.SearchParams) (*content.ContentReader, error) {
	cw, err := content.NewContentWriter(content.DefaultKey, true, false)
	if err != nil {
		return nil, err
	}
	defer cw.Close()
	item := utils.ResultItem{
		Actual_Sha1: "456",
		Properties: []utils.Property{
			{Key: "build.name", Value: "Build-Name"},
			{Key: "build.number", Value: "Build-Number"},
			{Key: "vcs.url", Value: "www.vcs.com"},
			{Key: "vcs.revision", Value: "248"},
		},
	}
	cw.Write(item)
	return content.NewContentReader(cw.GetFilePath(), cw.GetArrayKey()), nil
}

func TestGetDependenciesDetails(t *testing.T) {
	modules := []buildinfo.Module{{Id: "my-plugin:", Artifacts: []buildinfo.Artifact{
		{Name: "Artifact-name", Type: "Type", Checksum: &buildinfo.Checksum{Sha1: "123"}},
	}, Dependencies: []buildinfo.Dependency{
		{Id: "Dependency", Type: "File", Checksum: &buildinfo.Checksum{Sha1: "456"}},
	}}}
	smMock := new(ServiceManagerMock)

	sha1ToBuildProps, err := getDependenciesDetails(modules, "repository", smMock)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []map[string]*DependencyProps{sha1ToBuildProps}, []map[string]*DependencyProps{GetFirstSearchResultSortedByAsc()})
}

func GetFirstSearchResultSortedByAsc() map[string]*DependencyProps {
	return map[string]*DependencyProps{
		"456": {Build: "Build-Name/Build-Number", Vcs: buildinfo.Vcs{Url: "www.vcs.com", Revision: "248"}},
	}
}
