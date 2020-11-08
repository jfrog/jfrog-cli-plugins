package commands

import (
	"errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"os"
)

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ArtifactoryDetails, error) {
	serverId := c.GetStringFlagValue(serverIdFlag)
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

// Returns the build details that were provided by arguments or environment variables.
func getBuildDetails(c *components.Context) (buildName, buildNumber string, err error) {
	if len(c.Arguments) == 2 {
		return c.Arguments[0], c.Arguments[1], nil
	}
	buildName, buildNumber = utils.GetBuildNameAndNumber("", "")
	if buildName == "" || buildNumber == "" {
		return "", "",
			errors.New("build name and build number are expected as command arguments or environment variables")
	}
	return
}

// Get build info from Artifactory.
func getBuildInfo(rtDetails *config.ArtifactoryDetails, buildName, buildNumber string) (*buildinfo.PublishedBuildInfo, error) {
	servicesManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return nil, err
	}
	buildInfoParams := services.BuildInfoParams{BuildName: buildName, BuildNumber: buildNumber}
	return servicesManager.GetBuildInfo(buildInfoParams)
}

func renderWithDefaults(t table.Writer) {
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Title.Align = text.AlignCenter
	t.Render()
}
