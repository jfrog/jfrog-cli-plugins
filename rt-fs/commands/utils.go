package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/common/commands"

	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/log"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
)

func getCommonArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "path",
			Description: "[Mandatory] Path in Artifactory.",
		},
	}
}

func getCommonFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID configured using the config command.",
		},
	}
}

type commonConfiguration struct {
	details *config.ServerDetails
	path    string
}

func createCommonConfiguration(c *components.Context) (*commonConfiguration, error) {
	if err := checkInputs(c); err != nil {
		return nil, err
	}

	confDetails, err := getRtDetails(c)
	if err != nil {
		return nil, err
	}

	// Increase log level to avoid search command logs
	increaseLogLevel()

	conf := &commonConfiguration{
		details: confDetails,
		path:    c.Arguments[0],
	}
	return conf, nil
}

func checkInputs(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	if strings.Contains(c.Arguments[0], "*") {
		return errors.New("Wildcards are not supported in paths.")
	}
	return nil
}

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ServerDetails, error) {
	details, err := commands.GetConfig(c.GetStringFlagValue("server-id"), false)
	if err != nil {
		return nil, err
	}
	if details.ArtifactoryUrl == "" {
		return nil, errors.New("no server-id was found, or the server-id has no Artifactory url.")
	}
	details.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(details.ArtifactoryUrl)
	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	if err != nil {
		return nil, err
	}
	return details, nil
}

// If the results are a single directory and the pattern does not end with "/",
// we should run a second search with "/" in the end of the pattern.
func shouldRunSecondSearch(path string, reader *content.ContentReader) (bool, error) {
	// If the path is a repository or the path is already ends with "/", a second search is not needed.
	if !strings.Contains(path, "/") || strings.HasSuffix(path, "/") {
		return false, nil
	}

	// Check the length of the results. If the length != 1, don't search again.
	length, err := reader.Length()
	if err != nil {
		reader.Close()
		return false, err
	}
	if length != 1 {
		return false, nil
	}

	// Check the type of the single result. If it's not a folder - don't search again.
	result := new(utils.SearchResult)
	if err := reader.NextRecord(result); err != nil {
		reader.Close()
		return false, err
	}
	if result.Type != "folder" {
		// Reset the reader to allow reading again the file.
		reader.Reset()
		return false, nil
	}
	return true, nil
}

// Set the log level to ERROR to avoid the following outputs:
// [Info] Searching artifacts...
// [Info] Found 1 artifact.
func increaseLogLevel() {
	if log.GetCliLogLevel() == clientlog.INFO {
		clientlog.SetLogger(clientlog.NewLogger(clientlog.ERROR, nil))
	}
}

// Trim the pattern from path.
func trimFoldersFromPath(pattern, path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return path
	}
	return path[lastSlash+1:]
}
