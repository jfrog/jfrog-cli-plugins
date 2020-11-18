package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	commandsutils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/stretchr/testify/assert"
)

var trimPrefixProvider = []struct {
	pattern  string
	path     string
	expected string
}{
	{"abc", "abc", "abc"},
	{"abc", "abc/def", "def"},
	{"abc/", "abc/def", "def"},
	{"abc/def", "abc/def", "def"},
	{"abc/def/", "abc/def", "def"},
}

func TestTrimPrefixFromPath(t *testing.T) {
	for _, triple := range trimPrefixProvider {
		t.Run(fmt.Sprintf("%v", triple), func(t *testing.T) {
			assert.Equal(t, triple.expected, trimFoldersFromPath(triple.pattern, triple.path))
		})
	}
}

var processSearchResultsProvider = []struct {
	sampleName    string
	searchPattern string
	expectedPaths []string
	expectedTypes []string
	maxPathLength int
}{
	{"sample1.json", "libs-release-local", []string{"org", ".nupkg"}, []string{"folder", "file"}, 6},
	{"sample1.json", "libs-release-local/", []string{"org", ".nupkg"}, []string{"folder", "file"}, 6},
	{"sample2.json", "libs-release-local/org/jfrog/example/gradle/", []string{"services", "shared", "api", "webservice", "gradle-example-ci-server"}, []string{"folder", "folder", "folder", "folder", "folder", "folder"}, 24},
	{"sample3.json", "libs-release-local/org", []string{"org"}, []string{"folder", "file"}, 3},
}

func TestProcessSearchResults(t *testing.T) {
	searchResults := &commandsutils.Result{}
	for _, sample := range processSearchResultsProvider {
		t.Run(fmt.Sprintf("%v", sample), func(t *testing.T) {
			filePath := prepareSample(t, sample.sampleName)
			contentReader := content.NewContentReader(filePath, "results")
			searchResults.SetReader(contentReader)
			assert.NoError(t, checkSearchResults(contentReader, sample.searchPattern))
			actualResults, actualMaxPath, err := processSearchResults(sample.searchPattern, searchResults.Reader())
			assert.NoError(t, err)
			assert.Equal(t, sample.maxPathLength, actualMaxPath)
			for i := range sample.expectedPaths {
				assert.Equal(t, actualResults[i].Path, sample.expectedPaths[i])
				assert.Equal(t, actualResults[i].Type, sample.expectedTypes[i])
			}
		})
	}
}

func TestCheckSearchResultsFail(t *testing.T) {
	searchResults := &commandsutils.Result{}
	contentReader := content.NewContentReader("", "results")
	searchResults.SetReader(contentReader)
	err := checkSearchResults(searchResults.Reader(), "dummyPattern")
	assert.EqualError(t, err, "ls: cannot access 'dummyPattern': No such file or directory")
}

var shouldRunSecondSearchProvider = []struct {
	sampleName string
	path       string
	expected   bool
}{
	{"sample1.json", "libs-release-local", false},
	{"sample1.json", "libs-release-local/", false},
	{"sample2.json", "libs-release-local/org/jfrog/example/gradle/", false},
	{"sample2.json", "libs-release-local/org/", false},
	{"sample3.json", "libs-release-local/org", true},
}

func TestShouldRunSecondSearch(t *testing.T) {
	searchResults := &commandsutils.Result{}
	for _, sample := range shouldRunSecondSearchProvider {
		t.Run(fmt.Sprintf("%v", sample), func(t *testing.T) {
			filePath := prepareSample(t, sample.sampleName)
			contentReader := content.NewContentReader(filePath, "results")
			searchResults.SetReader(contentReader)
			actual, err := shouldRunSecondSearch(sample.path, contentReader)
			assert.NoError(t, err)
			assert.Equal(t, sample.expected, actual)
		})
	}
}

// Copy a sample from commands/testdata to a temp file and return the target path
func prepareSample(t *testing.T, fileName string) string {
	// Read sample from commands/testdata/
	dir, err := os.Getwd()
	assert.NoError(t, err)
	source := filepath.Join(dir, "testdata", fileName)
	data, err := ioutil.ReadFile(source)
	assert.NoError(t, err)

	// Write sample to a temp file
	dest, err := ioutil.TempFile("", fileName)
	assert.NoError(t, err)
	err = ioutil.WriteFile(dest.Name(), data, 0644)
	assert.NoError(t, err)
	return dest.Name()
}
