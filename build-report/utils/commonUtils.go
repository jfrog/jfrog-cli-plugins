package utils

import (
	"errors"
	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"os"
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
	buildName = os.Getenv(coreutils.BuildName)
	buildNumber = os.Getenv(coreutils.BuildNumber)
	if buildName == "" || buildNumber == "" {
		return "", "",
			errors.New("build name and build number are expected as command arguments or environment variables")
	}
	return
}

// Get build info from Artifactory.
func GetBuildInfo(rtDetails *config.ServerDetails, buildName, buildNumber string) (*entities.PublishedBuildInfo, bool, error) {
	servicesManager, err := utils.CreateServiceManager(rtDetails, -1, 0, false)
	if err != nil {
		return nil, false, err
	}
	buildInfoParams := services.BuildInfoParams{BuildName: buildName, BuildNumber: buildNumber}
	return servicesManager.GetBuildInfo(buildInfoParams)
}
