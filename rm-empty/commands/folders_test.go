package commands

import (
	clientrtutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterEmptyFolders(t *testing.T) {
	// Write the test items into resultsWriter.
	resultsWriterClosed := false
	resultsWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
	assert.NoError(t, err)
	defer func() {
		if !resultsWriterClosed {
			assert.NoError(t, resultsWriter.Close())
		}
	}()
	for _, result := range getTestItems() {
		resultsWriter.Write(result)
	}

	// Close the writer now, so that we can read its content using a reader.
	assert.NoError(t, resultsWriter.Close())
	resultsWriterClosed = true

	// Read the test items into resultsReader.
	resultsReader := content.NewContentReader(resultsWriter.GetFilePath(), content.DefaultKey)
	defer func() {
		assert.NoError(t, resultsReader.Close())
	}()

	// Sort the test in resultsReader into sortedResultsReader.
	sortedResultsReader, err := content.SortContentReader(clientrtutils.ResultItem{}, resultsReader, true)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, sortedResultsReader.Close())
	}()

	// Create an empty writer, to store the empty folders.
	emptyFoldersWriterClosed := false
	emptyFoldersWriter, err := content.NewContentWriter(content.DefaultKey, true, false)
	assert.NoError(t, err)
	defer func() {
		if !emptyFoldersWriterClosed {
			assert.NoError(t, emptyFoldersWriter.Close())
		}
	}()

	// Run the filterEmptyFolders function, which writes all the empty folders in sortedResultsReader
	// into emptyFoldersWriter.
	filterEmptyFolders(sortedResultsReader, emptyFoldersWriter)

	// Close the writer now, so that we can read its content using a reader.
	assert.NoError(t, emptyFoldersWriter.Close())
	emptyFoldersWriterClosed = true

	// Read the items emptyFoldersWriter into emptyFoldersReader.
	emptyFoldersReader := content.NewContentReader(emptyFoldersWriter.GetFilePath(), content.DefaultKey)
	defer func() {
		assert.NoError(t, emptyFoldersReader.Close())
	}()

	// Sort the items in emptyFoldersReader into sortedEmptyFoldersReader.
	sortedEmptyFoldersReader, err := content.SortContentReader(clientrtutils.ResultItem{}, emptyFoldersReader, true)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, sortedEmptyFoldersReader.Close())
	}()

	// Assert that the content of sortedEmptyFoldersReader is identical to getTestEmptyFolders().
	assertEmptyFolders(t, sortedEmptyFoldersReader)
}

func assertEmptyFolders(t *testing.T, actualEmptyFoldersReader *content.ContentReader) {
	// Get expected and actual length.
	expectedEmptyFolders := getTestEmptyFolders()
	expectedEmptyFoldersLen := len(expectedEmptyFolders)
	actualEmptyFoldersLen, err := actualEmptyFoldersReader.Length()
	assert.NoError(t, err)

	// Assert expected and actual length.
	assert.Equal(t, expectedEmptyFoldersLen, actualEmptyFoldersLen, "Unexpected readers size. Expected %d but got %d", expectedEmptyFoldersLen, actualEmptyFoldersLen)

	// Assert the content of actualEmptyFoldersReader
	i := 0
	for item := new(clientrtutils.ResultItem); actualEmptyFoldersReader.NextRecord(item) == nil; item = new(clientrtutils.ResultItem) {
		assert.Equal(t, expectedEmptyFolders[i].Path, item.Path, "Unexpected empty folder path. Expected %s but got %s", expectedEmptyFolders[i].Path, item.Path)
		assert.Equal(t, "folder", item.Type, "Unexpected item type. Expected a folder but got %s", item.Type)
		i++
	}
	assert.NoError(t, actualEmptyFoldersReader.GetError())
}

func getTestItems() []clientrtutils.ResultItem {
	return []clientrtutils.ResultItem{
		clientrtutils.ResultItem{
			Path: "a/b",
			Type: "folder",
		},
		clientrtutils.ResultItem{
			Path: "a/b/c",
			Type: "folder",
		},
		clientrtutils.ResultItem{
			Path: "a/b/c/d",
			Type: "folder",
		},
		clientrtutils.ResultItem{
			Path: "a/b/c/a.zip",
			Type: "file",
		},
		clientrtutils.ResultItem{
			Path: "a/b/1",
			Type: "folder",
		},
	}
}

func getTestEmptyFolders() []clientrtutils.ResultItem {
	return []clientrtutils.ResultItem{
		clientrtutils.ResultItem{
			Path: "a/b/1",
			Type: "folder",
		},
		clientrtutils.ResultItem{
			Path: "a/b/c/d",
			Type: "folder",
		},
	}
}
