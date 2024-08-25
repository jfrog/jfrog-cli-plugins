package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	searchutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

func GetCleanCommand() components.Command {
	return components.Command{
		Name:        "clean",
		Description: "Deletes all artifacts that have not been downloaded for the past n time units.",
		Aliases:     []string{"c"},
		Arguments:   getCleanArguments(),
		Flags:       getCleanFlags(),
		EnvVars:     getCleanEnvVar(),
		Action:      cleanCmd,
	}
}

func getCleanArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "repository",
			Description: "A repository to clean",
		},
	}
}

func getCleanFlags() []components.Flag {
	return []components.Flag{
		components.NewStringFlag("server-id", "Artifactory server ID configured using the config command."),
		components.NewStringFlag("time-unit", "The time unit of the no-dl time. year, month and day are the allowed values.", components.WithStrDefaultValue("month")),
		components.NewStringFlag("no-dl", "Artifacts that have not been downloaded or modified for at least no-dl will be deleted.", components.WithStrDefaultValue("1")),
	}
}

func getCleanEnvVar() []components.EnvVar {
	return []components.EnvVar{
		{},
	}
}

type cleanConfiguration struct {
	repository       string
	noDownloadedTime string
}

func cleanCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	var conf = new(cleanConfiguration)
	conf.repository = c.Arguments[0]
	noDownloadedTime, err := parseTimeFlags(c.GetStringFlagValue("no-dl"), c.GetStringFlagValue("time-unit"))
	if err != nil {
		return err
	}
	conf.noDownloadedTime = noDownloadedTime
	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}
	return cleanArtifacts(conf, rtDetails)
}

func cleanArtifacts(config *cleanConfiguration, artifactoryDetails *config.ServerDetails) error {
	// Search for artifacts to delete using AQL
	aqlQuery := buildAQL(config)
	authConfig, err := artifactoryDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	rtConf, err := searchutils.NewCommonConfImpl(authConfig)
	if err != nil {
		return err
	}
	resultReader, err := searchutils.ExecAqlSaveToFile(aqlQuery, rtConf)
	if err != nil {
		return err
	}
	defer resultReader.Close()

	// Delete the artifacts we found
	serviceManager, err := utils.CreateServiceManager(artifactoryDetails, -1, 0, false)
	if err != nil {
		return err
	}
	_, err = serviceManager.DeleteFiles(resultReader)
	return err
}

func buildAQL(c *cleanConfiguration) (aqlQuery string) {
	// Finds all artfacts that hasn't been downloaded or modified for at least noDownloadedTime
	aqlQuery = `items.find({` +
		`"type":"file",` +
		`"repo":%[1]q,` +
		`"$or":[` +
		`{"$and":[` +
		`{"modified":{"$before":%[2]q}},` +
		`{"stat.downloaded":{"$before":%[2]q}},` +
		`{"stat.downloads":{"$gt":"0"}}` +
		`]},` +
		`{"$and":[` +
		`{"modified":{"$before":%[2]q}},` +
		`{"stat.downloads":{"$eq":null}}` +
		`]}` +
		`]` +
		`})`

	return fmt.Sprintf(aqlQuery, c.repository, c.noDownloadedTime)
}

// given the 2 inputs: timeUnit and time returns a string represents this time interval.
// For example: 1, month => 1mo
func parseTimeFlags(noDownloadedTime, timeUnit string) (timeString string, err error) {
	// Validate no-dl
	timeValue, err := strconv.Atoi(noDownloadedTime)
	if err != nil {
		return
	}
	timeString = strconv.Itoa(timeValue)

	switch timeUnit = strings.ToLower(strings.TrimSpace(timeUnit)); timeUnit {
	case "year":
		return timeString + "y", nil

	case "month":
		return timeString + "mo", nil

	case "day":
		return timeString + "d", nil
	}
	return "", errors.New("wrong timeUnit arguments. Expected: year, month or day. Received: " + timeUnit)

}

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ServerDetails, error) {
	serverId := c.GetStringFlagValue("server-id")
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
