package commands

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestValidateSpecPath(t *testing.T) {
	// Create temp workdir.
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer fileutils.RemoveTempDir(tempDirPath)

	// Expect error when path is a directory.
	err = validateSpecPath(tempDirPath + string(filepath.Separator))
	assert.Error(t, err)

	// Expect no error.
	err = validateSpecPath(filepath.Join(tempDirPath, "filespec-test"))
	assert.NoError(t, err)

	// Expect error when file already exists.
	err = ioutil.WriteFile(filepath.Join(tempDirPath, "filespec-test"), []byte("This is test file content."), 0644)
	assert.NoError(t, err)
	err = validateSpecPath(filepath.Join(tempDirPath, "filespec-test"))
	assert.Error(t, err)
}

func TestHandleResult(t *testing.T) {
	contentToWrite := []byte("This is the result content.")

	// Create temp workdir.
	tempDirPath, err := fileutils.CreateTempDir()
	assert.NoError(t, err)
	defer fileutils.RemoveTempDir(tempDirPath)
	filePath := filepath.Join(tempDirPath, "filespec-test")

	// Test output to file.
	err = handleResult(contentToWrite, filePath)
	assert.NoError(t, err)
	actualContent, err := fileutils.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, string(contentToWrite), string(actualContent))
}

func TestBuildFileSpecJson(t *testing.T) {
	// Create expected result.
	expectedFile1 := &testFile{Pattern: "testPattern", Limit: "3"}
	expectedFile2 := &testFile{Pattern: "testAql", Props: "a=b;c=d", Recursive: "true"}
	expectedResult := testFileSpec{Files: []*testFile{expectedFile1, expectedFile2}}

	// Create test data.
	resultMap1 := map[string]interface{}{"pattern": "testPattern", "limit": "3"}
	resultMap2 := map[string]interface{}{"pattern": "testAql", "props": "a=b;c=d", "recursive": "true"}
	var resultSlice []specslicetype
	resultSlice = append(resultSlice, resultMap1, resultMap2)

	// Test produced file-spec json.
	actualResult, err := buildFileSpecJson(resultSlice)
	assert.NoError(t, err)
	var actualResultStruct testFileSpec
	err = json.Unmarshal(actualResult, &actualResultStruct)
	assert.NoError(t, err)
	assertEqualTestSpecs(t, &expectedResult, &actualResultStruct)
}

func assertEqualTestSpecs(t *testing.T, expected, actual *testFileSpec) {
	assert.Equal(t, len(expected.Files), len(actual.Files))
	for i, expectedFile := range expected.Files {
		actualFile := actual.Files[i]
		assert.Equal(t, expectedFile.Pattern, actualFile.Pattern, getAssertMessage("pattern", expectedFile.Pattern, actualFile.Pattern, i))
		assert.Equal(t, expectedFile.Limit, actualFile.Limit, getAssertMessage("limit", expectedFile.Limit, actualFile.Limit, i))
		assert.Equal(t, expectedFile.Props, actualFile.Props, getAssertMessage("props", expectedFile.Props, actualFile.Props, i))
		assert.Equal(t, expectedFile.Recursive, actualFile.Recursive, getAssertMessage("recursive", expectedFile.Recursive, actualFile.Recursive, i))
	}
}

func getAssertMessage(fieldName, expected, actual string, fileSpecNumber int) string {
	return fmt.Sprintf("The expected %s value of file-spec #%d is: %s, but the actual is: %s", fieldName, fileSpecNumber, expected, actual)
}

type testFileSpec struct {
	Files []*testFile `json:"files,omitempty"`
}

type testFile struct {
	Pattern   string `json:"pattern,omitempty"`
	Limit     string `json:"limit,omitempty"`
	Props     string `json:"props,omitempty"`
	Recursive string `json:"recursive,omitempty"`
}
