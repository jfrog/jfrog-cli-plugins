package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/buger/goterm"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
)

// The minimal space between ls results in the screen
const minSpace = 1

func GetLsCommand() components.Command {
	return components.Command{
		Name:        "ls",
		Description: "Run ls.",
		Aliases:     []string{"ls, list"},
		Arguments:   getCommonArguments(),
		Flags:       getCommonFlags(),
		Action: func(c *components.Context) error {
			return lsCmd(c)
		},
	}
}

func lsCmd(c *components.Context) error {
	conf, err := createCommonConfiguration(c)
	if err != nil {
		return err
	}

	return doLs(conf)
}

func doLs(c *commonConfiguration) error {
	// Execute search command
	reader, err := doSearch(c)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Get structured search results and the path with the max length
	searchResults, maxPathLength, err := processSearchResults(c.path, reader)
	if err != nil {
		return err
	}

	// Print results
	printLsResults(searchResults, maxPathLength)

	return nil
}

func doSearch(c *commonConfiguration) (*content.ContentReader, error) {
	// Run the first search
	searchCmd := generic.NewSearchCommand()
	searchSpec := spec.NewBuilder().Pattern(c.path).IncludeDirs(true).BuildSpec()
	searchCmd.SetServerDetails(c.details).SetSpec(searchSpec)
	if err := commands.Exec(searchCmd); err != nil {
		return nil, err
	}

	// Check the search results
	reader := searchCmd.Result().Reader()
	if err := checkSearchResults(reader, c.path); err != nil {
		reader.Close()
		return nil, err
	}

	// Check if a second search is needed
	runSecondSearch, err := shouldRunSecondSearch(c.path, reader)
	if !runSecondSearch || err != nil {
		return reader, err
	}

	// Close the first search reader
	if err := reader.Close(); err != nil {
		return nil, err
	}

	// Run search again with "/" in the end of the pattern
	searchSpec = spec.NewBuilder().Pattern(c.path + "/").IncludeDirs(true).BuildSpec()
	searchCmd.SetSpec(searchSpec)
	err = commands.Exec(searchCmd)
	return searchCmd.Result().Reader(), err
}

func printLsResults(searchResults []utils.SearchResult, maxPathLength int) {
	maxPathLength += minSpace
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
func processSearchResults(pattern string, reader *content.ContentReader) ([]utils.SearchResult, int, error) {
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
	return allResults, maxPathLength, reader.GetError()
}

// Check validity of search results.
func checkSearchResults(reader *content.ContentReader, pattern string) error {
	if length, err := reader.Length(); length == 0 {
		if err == nil {
			err = errors.New("ls: cannot access '" + pattern + "': No such file or directory")
		}
		return err
	}
	return nil
}
