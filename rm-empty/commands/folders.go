package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	clientrtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"

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
		Action:      foldersCmd,
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
		components.NewStringFlag("server-id", "Artifactory server ID configured using the config command"),
		components.NewBoolFlag("quiet", "Skip the delete confirmation message"),
	}
}

func getEnvVars() []components.EnvVar {
	return []components.EnvVar{
		{},
	}
}

func foldersCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}

	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}

	path := c.Arguments[0]
	return deleteEmptyFolders(rtDetails, path, c.GetBoolFlagValue("quiet"))
}

// Deletes all the empty folders under the specified path in Artifactory.
func deleteEmptyFolders(rtDetails *config.ServerDetails, path string, quiet bool) (err error) {
	// Create a search command, to find all the files and folders under the specified path.
	spec := spec.NewBuilder().Pattern(path).IncludeDirs(true).Recursive(true).BuildSpec()
	cmd := generic.NewSearchCommand()
	cmd.SetServerDetails(rtDetails).SetSpec(spec).SetRetries(3)

	log.Info("Searching for all items under", path)

	// Run the search and receive a reader with the results.
	var reader *content.ContentReader
	if reader, err = cmd.Search(); err != nil {
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
	if sortedFilesReader, err = content.SortContentReader(clientrtutils.ResultItem{}, reader, true); err != nil {
		return
	}

	// Create a writer, that will be used to store the paths of the empty folders found.
	var emptyFoldersWriter *content.ContentWriter
	emptyFoldersWriter, err = content.NewContentWriter(content.DefaultKey, true, false)
	emptyFoldersReaderClosed := false
	defer func() {
		if !emptyFoldersReaderClosed {
			e := emptyFoldersWriter.Close()
			if err == nil {
				err = e
			}
		}
	}()

	// Find all empty folders by scanning the sortedFilesReader, and write them into the emptyFoldersWriter.
	var totalFound int
	totalFound, err = filterEmptyFolders(sortedFilesReader, emptyFoldersWriter)
	if err != nil {
		return
	}

	logEmptyFoldersFound(totalFound)

	// The writer needs to be closed before it cam be read from.
	err = emptyFoldersWriter.Close()
	if err != nil {
		return
	}
	emptyFoldersReaderClosed = true

	// Create a reader from the writer, to read all the empty folders found.
	emptyFoldersReader := content.NewContentReader(emptyFoldersWriter.GetFilePath(), content.DefaultKey)
	defer func() {
		e := emptyFoldersReader.Close()
		if err == nil {
			err = e
		}
	}()

	var length int
	length, err = emptyFoldersReader.Length()
	if err != nil {
		return
	}
	if length == 0 {
		return
	}

	// Delete the folders in the reader.
	return deleteItem(emptyFoldersReader, rtDetails, quiet)
}

func logEmptyFoldersFound(totalFound int) {
	switch totalFound {
	case 0:
		log.Info("Found no empty folders.")
	case 1:
		log.Info("Found 1 empty folder.")
	default:
		log.Info("Found", totalFound, "empty folders.")
	}
}

// Find all empty folders by scanning the sortedFilesReader, and write them into the emptyFoldersWriter.
func filterEmptyFolders(sortedFilesReader *content.ContentReader, emptyFoldersWriter *content.ContentWriter) (totalFound int, err error) {
	var prevFolder *clientrtutils.ResultItem
	for item := new(clientrtutils.ResultItem); sortedFilesReader.NextRecord(item) == nil; item = new(clientrtutils.ResultItem) {
		if prevFolder != nil && !strings.HasPrefix(item.Path, prevFolder.Path) {
			emptyFoldersWriter.Write(prevFolder)
			totalFound++
		}
		if item.Type == "folder" && !isRepo(item.Path) {
			prevFolder = item
		} else {
			prevFolder = nil
		}
	}
	if prevFolder != nil {
		emptyFoldersWriter.Write(prevFolder)
		totalFound++
	}
	return totalFound, sortedFilesReader.GetError()
}

// Returns true if the provided path leads to the root of a repository.
func isRepo(path string) bool {
	slashCount := strings.Count(path, "/")
	if strings.HasSuffix(path, "/") {
		return slashCount == 1
	}
	return slashCount == 0
}

// Deletes the paths sent in the provided reader from the provided Artifactory server.
func deleteItem(reader *content.ContentReader, rtDetails *config.ServerDetails, quiet bool) (err error) {
	// Create a delete command
	cmd := generic.NewDeleteCommand()
	cmd.SetServerDetails(rtDetails)

	// Get confirmation from the users before deleted the paths.
	allowDelete := quiet
	if !quiet {
		allowDelete, err = utils.ConfirmDelete(reader)
	}
	if err != nil || !allowDelete {
		return
	}

	// Delete the paths from Artifactory.
	_, _, err = cmd.DeleteFiles(reader)
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
