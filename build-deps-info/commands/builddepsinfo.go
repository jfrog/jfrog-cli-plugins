package builddepsinfo

import (
	"errors"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jfrog/jfrog-cli-plugins/build-deps-info/commands/utils"
	servicesutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// Artifactory has a max number of character for a single request,
// therefore we limit the maximum number of sha1 for a single AQL request.
const (
	sha1BatchSize = 125
)

type BuildDepsInfo struct {
	buildName       string
	buildNumber     string
	repository      string
	servicesManager artifactory.ArtifactoryServicesManager
}

func NewBuildDepsInfo() *BuildDepsInfo {
	return &BuildDepsInfo{}
}

func (p *BuildDepsInfo) SetBuildName(buildName string) *BuildDepsInfo {
	p.buildName = buildName
	return p
}

func (p *BuildDepsInfo) SetBuildNumber(buildNumber string) *BuildDepsInfo {
	p.buildNumber = buildNumber
	return p
}

func (p *BuildDepsInfo) SetRepository(repository string) *BuildDepsInfo {
	p.repository = repository
	return p
}

func (p *BuildDepsInfo) SetServicesManager(servicesManager artifactory.ArtifactoryServicesManager) *BuildDepsInfo {
	p.servicesManager = servicesManager
	return p
}

func (p *BuildDepsInfo) Exec() error {
	biParams := services.NewBuildInfoParams()
	biParams.BuildName, biParams.BuildNumber = p.buildName, p.buildNumber
	buildinfo, found, err := p.servicesManager.GetBuildInfo(biParams)
	if err != nil || !found {
		return err
	}
	if buildinfo.BuildInfo.Name == "" || buildinfo.BuildInfo.Number == "" {
		return errors.New("Build '" + p.buildName + "/" + p.buildNumber + "' not found")
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Module Id", "Dependency name", "BUILD", "VCS URL"})
	sha1ToBuildProps, err := getDependenciesDetails(buildinfo.BuildInfo.Modules, p.repository, p.servicesManager)
	for _, module := range buildinfo.BuildInfo.Modules {
		for _, dep := range module.Dependencies {
			depPropsInfo := sha1ToBuildProps[dep.Sha1]
			t.AppendRow(table.Row{module.Id, utils.Optional(dep.Id), utils.Optional(depPropsInfo.Build), utils.OptionalVcsUrl(&depPropsInfo.Vcs)})
		}
		t.Render()
	}
	return nil
}

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(serverId string) (*config.ArtifactoryDetails, error) {
	details, err := commands.GetConfig(serverId, false)
	if err != nil {
		return nil, err
	}
	if details.Url == "" {
		return nil, errors.New("no server-id was found, or the server-id has no url")
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

type DependencyProps struct {
	Build string
	buildinfo.Vcs
}

func getDependenciesDetails(bim []buildinfo.Module, repo string, sm artifactory.ArtifactoryServicesManager) (result map[string]*DependencyProps, err error) {
	result = make(map[string]*DependencyProps)
	// List of dependencies sha1.
	sha1Set := utils.NewStringSet()
	for _, module := range bim {
		for _, dep := range module.Dependencies {
			result[dep.Sha1] = &DependencyProps{}
			sha1Set.Add(dep.Sha1)
		}
	}
	reader, err := getArtifactsPropsBySha1(repo, sha1Set, sm)
	if err != nil || reader == nil {
		return
	}
	defer utils.Cleanup(reader.Close, &err)
	// Update the dependencies build.
	for currentResult := new(serviceutils.ResultItem); reader.NextRecord(currentResult) == nil; currentResult = new(serviceutils.ResultItem) {
		var buildName, buildNumber, vcsUrl, vcsRevision string
		for _, prop := range currentResult.Properties {
			switch prop.Key {
			case "build.name":
				buildName = prop.Value + "/"
			case "build.number":
				buildNumber = prop.Value
			case "vcs.url":
				vcsUrl = prop.Value
			case "vcs.revision":
				vcsRevision = prop.Value
			}
		}
		item := result[currentResult.Actual_Sha1]
		item.Build = buildName + buildNumber
		item.Vcs = buildinfo.Vcs{Url: vcsUrl, Revision: vcsRevision}
	}
	return
}

// Search for artifacts properties by sha1.
// AQL requests have a size limit, therefore, we split the requests into small groups.
func getArtifactsPropsBySha1(repository string, sha1s *utils.StringSet, sm artifactory.ArtifactoryServicesManager) (readerResults *content.ContentReader, err error) {
	if sha1s.IsEmpty() {
		return
	}
	sha1Batches := utils.GroupItems(sha1s.ToSlice(), sha1BatchSize)
	searchResults := make([]*content.ContentReader, len(sha1Batches))
	producerConsumer := parallel.NewBounedRunner(3, false)
	errorsQueue := clientutils.NewErrorsQueue(1)
	handlerFunc := createGetArtifactsPropsBySha1Func(repository, sm, searchResults)
	go func() {
		defer producerConsumer.Done()
		for i, sha1Bach := range sha1Batches {
			producerConsumer.AddTaskWithError(handlerFunc(sha1Bach, i), errorsQueue.AddError)
		}
	}()
	producerConsumer.Run()
	if err := errorsQueue.GetError(); err != nil {
		return nil, err
	}
	var totalReaders []*content.ContentReader
	for _, reader := range searchResults {
		if reader == nil {
			continue
		}
		totalReaders = append(totalReaders, reader)
		defer utils.Cleanup(reader.Close, &err)
	}
	readerResults, err = content.MergeReaders(totalReaders, content.DefaultKey)
	return
}

// Creates a function that fetches dependency data from Artifactory.
func createGetArtifactsPropsBySha1Func(repo string, sm artifactory.ArtifactoryServicesManager, searchResult []*content.ContentReader) func(sha1s []string, index int) parallel.TaskFunc {
	return func(sha1s []string, index int) parallel.TaskFunc {
		return func(threadId int) error {
			start := time.Now()
			aql := utils.CreateSearchBySha1AndRepoAqlQuery(repo, sha1s)
			params := services.NewSearchParams()
			params.Aql = servicesutils.Aql{ItemsFind: aql}
			reader, err := sm.SearchFiles(params)
			if err != nil {
				return errorutils.CheckError(err)
			}
			t := time.Now()
			elapsed := t.Sub(start)
			log.Debug(clientutils.GetLogMsgPrefix(threadId, false), "Finished searching artifacts properties by sha1 in", repo, ". Took ", elapsed.Seconds(), " seconds to complete the operation.\n")
			searchResult[index] = reader
			return nil
		}
	}
}
