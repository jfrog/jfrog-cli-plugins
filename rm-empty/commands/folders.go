package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	clientrtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

const ServerIdFlag = "server-id"

func GetCleanCommand() components.Command {
	return components.Command{
		Name:        "folders",
		Description: "Remove all empty folders under the specified path in Artifactory",
		Aliases:     []string{"f"},
		Arguments:   getArguments(),
		Flags:       getFlags(),
		EnvVars:     getEnvVars(),
		Action: func(c *components.Context) error {
			return foldersCmd(c)
		},
	}
}

func getArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "path",
			Description: "Path in Artifactory. The path should start with a repository",
		},
	}
}

func getFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID configured using the config command.",
		},
	}
}

func getEnvVars() []components.EnvVar {
	return []components.EnvVar{
		{},
	}
}

func foldersCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}

	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}

	path := c.Arguments[0]
	return deleteEmptyFolders(rtDetails, path)
}

// Deletes all the empty folders under the specified path in Artifactory.
func deleteEmptyFolders(rtDetails *config.ServerDetails, path string) (err error) {
	// Create a search command, to find all the files and folders under the specified path.
	spec := spec.NewBuilder().Pattern(path).IncludeDirs(true).Recursive(true).BuildSpec()
	cmd := generic.NewSearchCommand()
	cmd.SetServerDetails(rtDetails).SetSpec(spec).SetRetries(3)

	// Run the search and receive a reader with the results.
	var reader *content.ContentReader
	reader, err = cmd.Search()
	if err != nil {
		return
	}
	defer func() {
		e := reader.Close()
		if err == nil {
			err = e
		}
	}()

	// Sort the results in the reader, so that the empty folders can be found by reading the results
	// record by record.
	var sortedFilesReader *content.ContentReader
	sortedFilesReader, err = content.SortContentReader(clientrtutils.ResultItem{}, reader, true)
	if err != nil {
		return
	}

	// Create a writer, that will be used to store the paths of the empty folders found.
	var emptyFoldersWriter *content.ContentWriter
	emptyFoldersWriter, err = content.NewContentWriter(content.DefaultKey, true, false)
	defer func() {
		e := emptyFoldersWriter.Close()
		if err == nil {
			err = e
		}
	}()

	// Find all empty folders by scanning the sortedFilesReader, and write them into the emptyFoldersWriter.
	filterEmptyFolders(sortedFilesReader, emptyFoldersWriter)

	// Create a reader from the writer, to read all the empty folders found.
	emptyFoldersReader := content.NewContentReader(emptyFoldersWriter.GetFilePath(), content.DefaultKey)
	defer func() {
		e := emptyFoldersReader.Close()
		if err == nil {
			err = e
		}
	}()

	// Delete the folders in the reader.
	deleteItem(emptyFoldersReader, rtDetails)
	return
}

// Find all empty folders by scanning the sortedFilesReader, and write them into the emptyFoldersWriter.
func filterEmptyFolders(sortedFile *content.ContentReader, emptyFoldersWriter *content.ContentWriter) {
	var prevItem *clientrtutils.ResultItem
	var lastItem *clientrtutils.ResultItem

	for item := new(clientrtutils.ResultItem); sortedFile.NextRecord(item) == nil; item = new(clientrtutils.ResultItem) {
		if prevItem != nil && !strings.HasPrefix(item.Path, prevItem.Path) {
			emptyFoldersWriter.Write(prevItem)
		}
		if item.Type == "folder" {
			prevItem = item
		} else {
			prevItem = nil
		}
		lastItem = item
	}
	if lastItem != nil && lastItem.Type == "folder" {
		emptyFoldersWriter.Write(prevItem)
	}
}

// Deletes the paths sent in the provided reader from the provided Artifactory server.
func deleteItem(reader *content.ContentReader, rtDetails *config.ServerDetails) (success, failure int, err error) {
	// Create a delete command
	cmd := generic.NewDeleteCommand()
	cmd.SetServerDetails(rtDetails)

	// Get confirmation from the users before deleted the paths.
	var allowDelete bool
	allowDelete, err = utils.ConfirmDelete(reader)
	if err != nil || !allowDelete {
		return
	}

	// Delete the paths from Artifactory.
	success, failure, err = cmd.DeleteFiles(reader)
	return
}

// Returns the Artifactory Details of the provided server-id, or the default one.
func getRtDetails(c *components.Context) (*config.ServerDetails, error) {
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