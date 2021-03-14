package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"

	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	servicesutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Get builds diff from Artifactory.
func GetBuildDiff(rtDetails *config.ServerDetails, buildName, buildNumber, buildNumberDiff string) (*BuildDiff, error) {
	if buildNumberDiff == "" {
		return nil, nil
	}

	artAuth, err := rtDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	httpClientsDetails := artAuth.CreateHttpClientDetails()
	client, err := jfroghttpclient.JfrogClientBuilder().SetServiceDetails(&artAuth).Build()
	if err != nil {
		return nil, err
	}

	restApi := path.Join("api", "build", buildName, buildNumber)
	params := map[string]string{"diff": buildNumberDiff}
	requestFullUrl, err := servicesutils.BuildArtifactoryUrl(artAuth.GetUrl(), restApi, params)
	log.Debug("Getting build-info diff from: ", requestFullUrl)

	resp, body, _, err := client.SendGet(requestFullUrl, true, &httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	var buildDiff BuildDiff
	if err := json.Unmarshal(body, &buildDiff); err != nil {
		return nil, err
	}
	return &buildDiff, nil
}

// Struct to hold the build diff response from Artifactory.
type BuildDiff struct {
	Artifacts    ArtifactsChanges    `json:"artifacts,omitempty"`
	Dependencies DependenciesChanges `json:"dependencies,omitempty"`
	Properties   PropertiesChanges   `json:"properties,omitempty"`
}

type ArtifactsChanges struct {
	Updated   []ArtifactDiff `json:"updated,omitempty"`
	Unchanged []ArtifactDiff `json:"unchanged,omitempty"`
	Removed   []ArtifactDiff `json:"removed,omitempty"`
	New       []ArtifactDiff `json:"new,omitempty"`
}

type DependenciesChanges struct {
	Updated   []DependencyDiff `json:"updated,omitempty"`
	Unchanged []DependencyDiff `json:"unchanged,omitempty"`
	Removed   []DependencyDiff `json:"removed,omitempty"`
	New       []DependencyDiff `json:"new,omitempty"`
}

type FileDiff interface {
	GetModuleId() string
	GetIdOrName() string
	GetDiffIdOrName() string
	GetType() string
	GetArtOrDep() string
	GetSha1() string
	GetMd5() string
}

type ArtifactDiff struct {
	Module   string `json:"module,omitempty"`
	DiffName string `json:"diffName,omitempty"`
	buildinfo.Artifact
}

func (a ArtifactDiff) GetModuleId() string {
	return a.Module
}

func (a ArtifactDiff) GetIdOrName() string {
	return a.Name
}

func (a ArtifactDiff) GetDiffIdOrName() string {
	return a.DiffName
}

func (a ArtifactDiff) GetType() string {
	return a.Type
}

func (a ArtifactDiff) GetArtOrDep() string {
	return "Artifact"
}

func (a ArtifactDiff) GetSha1() string {
	if a.Checksum == nil {
		return ""
	}
	return a.Sha1
}

func (a ArtifactDiff) GetMd5() string {
	if a.Checksum == nil {
		return ""
	}
	return a.Md5
}

type DependencyDiff struct {
	Module string `json:"module,omitempty"`
	DiffId string `json:"diffId,omitempty"`
	buildinfo.Dependency
}

func (d DependencyDiff) GetModuleId() string {
	return d.Module
}

func (d DependencyDiff) GetIdOrName() string {
	return d.Id
}

func (d DependencyDiff) GetDiffIdOrName() string {
	return d.DiffId
}

func (d DependencyDiff) GetType() string {
	return d.Type
}

func (d DependencyDiff) GetArtOrDep() string {
	return "Dependency"
}

func (d DependencyDiff) GetSha1() string {
	if d.Checksum == nil {
		return ""
	}
	return d.Sha1
}

func (d DependencyDiff) GetMd5() string {
	if d.Checksum == nil {
		return ""
	}
	return d.Md5
}

type PropertiesChanges struct {
	Updated   []PropertyDiff `json:"updated,omitempty"`
	Unchanged []PropertyDiff `json:"unchanged,omitempty"`
	Removed   []PropertyDiff `json:"removed,omitempty"`
	New       []PropertyDiff `json:"new,omitempty"`
}

type PropertyDiff struct {
	Key                  string `json:"key,omitempty"`
	Value                string `json:"value,omitempty"`
	DiffValue            string `json:"diffValue,omitempty"`
	CompoundKeyValue     string `json:"compoundKeyValue,omitempty"`
	CompoundDiffKeyValue string `json:"compoundDiffKeyValue,omitempty"`
}

type Change int

const (
	Updated Change = iota
	Unchanged
	Removed
	New
)

func (c Change) String() string {
	switch c {
	case Updated:
		return "Updated"
	case Unchanged:
		return "Unchanged"
	case Removed:
		return "Removed"
	case New:
		return "New"
	}
	return ""
}
