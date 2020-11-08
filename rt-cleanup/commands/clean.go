package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	searchutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
)

//timeUnit: The unit of the time interval. year, month, day, hour or minute are allowed values. Default month.
//timeInterval: The time interval to look back before deleting an artifact. Default 1.
// Deletes all artifacts that have not been downloaded for the past n time units.

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
			Description:  "The time unit of the maximal time. year, month, day are allowed values. Default month.",
			DefaultValue: "month",
		},
		components.StringFlag{
			Name:         "maximalTime",
			Description:  "Artifacts that have not been downloaded for the past maximalTime will be deleted.",
			DefaultValue: "1",
		},
		components.StringFlag{
			Name:         "maximalSize",
			Description:  "Artifacts that are smaller than maximalSize MB will not be deleted. Default 0",
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
	conf.maximalSize = c.GetStringFlagValue("maximalSize") + "000000"
	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}
	return cleanArtifcats(conf, rtDetails)
}

func cleanArtifcats(c *cleanConfiguration, artifactoryDetails *config.ArtifactoryDetails) error {
	serviceManager, err := utils.CreateServiceManager(artifactoryDetails, false)
	if err != nil {
		return err
	}
	// resultStream, err := serviceManager.Aql(aqlQuery)
	//	defer resultStream.Close()

	aqlQuery := buildAQL(c)
	authConfig, err := artifactoryDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	conf := new(searchutils.CommonConfImpl)
	conf.SetArtifactoryDetails(authConfig)
	conf.DryRun = false
	resultReader, err := searchutils.ExecAqlSaveToFile(aqlQuery, conf)
	if err != nil {
		return err
	}
	defer resultReader.Close()
	_, err = serviceManager.DeleteFiles(resultReader)

	return err
}

func buildAQL(c *cleanConfiguration) (aqlQuery string) {
	// Finds all artfacts that hasn't been downloaded for the last time before maximalTime
	// (or never been downloaded) and theirs size is bigger than maximalSize
	aqlQuery = `items.find({` +
		`"type":"file",` +
		`"repo":%q,` +
		`"size":{"$gt":%q},` +
		`"$or": [` +
		`{` +
		`"stat.downloaded":{"$before":%q},` +
		`"stat.downloads":{"$eq":null}` +
		`}` +
		`]` +
		`})`

	aqlQuery = fmt.Sprintf(aqlQuery, c.repository, c.maximalSize, c.maximalTime)
	return
}

// func deleteArtifacts() (err error) {
// 	result, err := ioutil.ReadAll(resultStream)
// 	if err != nil {
// 		return err
// 	}
// 	parsedResult := new(aqlResult)
// 	if err = json.Unmarshal(result, parsedResult); err != nil {
// 		return errorutils.CheckError(err)
// 	}
// 	for _, file := range parsedResult.Results {
// 		log.Output(file.Path)
// 	}
// }

func parseTimeFlagsToTime(maximalTime, timeUnit string) (maimalDate time.Time, err error) {
	now := time.Now()
	timeValue, err := strconv.Atoi(maximalTime)
	if err != nil {
		return
	}
	switch timeUnit = strings.ToLower(strings.TrimSpace(timeUnit)); timeUnit {
	case "year":
		return now.AddDate(-timeValue, 0, 0), nil

	case "month":
		return now.AddDate(0, -timeValue, 0), nil

	case "day":
		return now.AddDate(0, 0, -timeValue), nil
	}
	return time.Now(), errors.New("Wrong timeUnit arguments. Expected: year, month or day. Received: " + timeUnit)

}
func parseTimeFlags(maximalTime, timeUnit string) (timeString string, err error) {
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

type aqlResult struct {
	Results []*results `json:"results,omitempty"`
}

type results struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}
