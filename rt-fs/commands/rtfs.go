package commands

import (
	"fmt"
	"strconv"

	"github.com/buger/goterm"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"

	commandsutils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
)

func GetLsCommand() components.Command {
	return components.Command{
		Name:        "ls",
		Description: "Run ls.",
		Aliases:     []string{"ls, list"},
		Arguments:   getLsArguments(),
		Flags:       getLsFlags(),
		Action: func(c *components.Context) error {
			return lsCmd(c)
		},
	}
}

func getLsArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "path",
			Description: "[Mandatory] Path in Artifactory.",
		},
	}
}

func getLsFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID configured using the config command.",
		},
	}
}

type lsConfiguration struct {
	details *config.ArtifactoryDetails
	path    string
}

func lsCmd(c *components.Context) error {
	if err := checkInputs(c); err != nil {
		return err
	}

	confDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}

	// Increase log level to avoid search command logs
	increaseLogLevel()

	conf := &lsConfiguration{
		details: confDetails,
		path:    c.Arguments[0],
	}

	return doLs(conf)
}

func doLs(c *lsConfiguration) error {
	// Execute search command
	searchCmd := generic.NewSearchCommand()
	spec := spec.NewBuilder().Pattern(c.path).IncludeDirs(true).BuildSpec()
	searchCmd.SetRtDetails(c.details).SetSpec(spec)
	if err := commands.Exec(searchCmd); err != nil {
		return err
	}

	// Get structured search results and the path with the max length
	searchResults, maxPathLength, err := processSearchResults(c.path, searchCmd.Result())
	if err != nil {
		return err
	}

	// Print results
	printLsResults(searchResults, maxPathLength)

	return nil
}

func printLsResults(searchResults []utils.SearchResult, maxPathLength int) {
	maxResultsInLine := goterm.Width() / maxPathLength
	if maxResultsInLine == 0 {
		maxResultsInLine = 1
	}

	pattern := "%-" + strconv.Itoa(maxPathLength) + "s"
	var color int
	for i, res := range searchResults {
		if i > 0 && i%maxResultsInLine == 0 {
			fmt.Println()
		}
		if res.Type == "folder" {
			color = goterm.BLUE
		} else {
			color = goterm.WHITE
		}
		output := fmt.Sprintf(pattern, res.Path)
		fmt.Print(goterm.Color(output, color))
	}
	fmt.Println()
}

// Gets the search results and builds an array of SearchResults.
// Return also the path with the maximum size.
func processSearchResults(pattern string, searchResults *commandsutils.Result) ([]utils.SearchResult, int, error) {
	reader := searchResults.Reader()
	defer reader.Close()
	if err := checkSearchResults(reader, pattern); err != nil {
		return nil, 0, err
	}

	allResults := []utils.SearchResult{}
	maxPathLength := 0
	result := new(utils.SearchResult)
	for i := 0; reader.NextRecord(result) == nil; i++ {
		result.Path = trimFoldersFromPath(pattern, result.Path)
		pathLength := len(result.Path)
		if pathLength > 0 {
			if pathLength > maxPathLength {
				maxPathLength = pathLength
			}
			allResults = append(allResults, *result)
		}
		result = new(utils.SearchResult)
	}
	return allResults, maxPathLength + 1, reader.GetError()
}
