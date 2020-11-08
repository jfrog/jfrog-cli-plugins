package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
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
		Action: func(c *components.Context) error {
			return cleanCmd(c)
		},
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
		components.StringFlag{
			Name:         "timeUnit",
			Description:  "The time unit of the maximal time. year, month, day are allowed values.",
			DefaultValue: "month",
		},
		components.StringFlag{
			Name:         "maximalTime",
			Description:  "Artifacts that have not been downloaded for the past maximalTime will be deleted.",
			DefaultValue: "1",
		},
		components.StringFlag{
			Name:         "maximalSize",
			Description:  "Artifacts that are smaller than maximalSize (bytes) will not be deleted.",
			DefaultValue: "0",
		},
	}
}

func getCleanEnvVar() []components.EnvVar {
	return []components.EnvVar{
		{},
	}
}

type cleanConfiguration struct {
	repository  string
	maximalTime string
	maximalSize string
}

func cleanCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	var conf = new(cleanConfiguration)
	conf.repository = c.Arguments[0]
	maximalTime, err := parseTimeFlags(c.GetStringFlagValue("maximalTime"), c.GetStringFlagValue("timeUnit"))
	if err != nil {
		return err
	}
	conf.maximalTime = maximalTime
	conf.maximalSize = c.GetStringFlagValue("maximalSize")
	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}
	return cleanArtifcats(conf, rtDetails)
}

func cleanArtifcats(config *cleanConfiguration, artifactoryDetails *config.ArtifactoryDetails) error {
	// Search for artifacts to delete using AQL
	aqlQuery := buildAQL(config)
	authConfig, err := artifactoryDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	rtConf := new(searchutils.CommonConfImpl)
	rtConf.SetArtifactoryDetails(authConfig)
	resultReader, err := searchutils.ExecAqlSaveToFile(aqlQuery, rtConf)
	if err != nil {
		return err
	}
	defer resultReader.Close()

	// Delete the artifacts we found
	serviceManager, err := utils.CreateServiceManager(artifactoryDetails, false)
	if err != nil {
		return err
	}
	_, err = serviceManager.DeleteFiles(resultReader)
	return err
}

func buildAQL(c *cleanConfiguration) (aqlQuery string) {
	// Finds all artfacts that hasn't been downloaded for the last time before maximalTime
	// (or never been downloaded) and theirs size is bigger than maximalSize
	aqlQuery = `items.find({` +
		`"type":"file",` +
		`"repo":%q,` +
		`"size":{"$gte":%q},` +
		`"$or": [` +
		`{` +
		`"stat.downloaded":{"$before":%q},` +
		`"stat.downloads":{"$eq":null}` +
		`}` +
		`]` +
		`})`

	return fmt.Sprintf(aqlQuery, c.repository, c.maximalSize, c.maximalTime)
}

// given the 2 inputs: timeUnit and time returns a string represents this time interval.
// For example: 1, month => 1mo
func parseTimeFlags(maximalTime, timeUnit string) (timeString string, err error) {
	// Validate maximalTime
	timeValue, err := strconv.Atoi(maximalTime)
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
	return "", errors.New("Wrong timeUnit arguments. Expected: year, month or day. Received: " + timeUnit)

}

/// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ArtifactoryDetails, error) {
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
