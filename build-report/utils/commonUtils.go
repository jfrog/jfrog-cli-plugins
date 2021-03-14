package utils

import (
	"errors"

	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

const ServerIdFlag = "server-id"

// Returns the Artifactory Details of the provided server-id, or the default one.
func GetRtDetails(c *components.Context) (*config.ServerDetails, error) {
	serverId := c.GetStringFlagValue(ServerIdFlag)
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
func GetBuildDetails(c *components.Context) (buildName, buildNumber string, err error) {
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
func GetBuildInfo(rtDetails *config.ServerDetails, buildName, buildNumber string) (*buildinfo.PublishedBuildInfo, bool, error) {
	servicesManager, err := utils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return nil, false, err
	}
	buildInfoParams := services.BuildInfoParams{BuildName: buildName, BuildNumber: buildNumber}
	return servicesManager.GetBuildInfo(buildInfoParams)
}
