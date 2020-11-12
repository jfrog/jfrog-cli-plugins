package create

import (
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
	ioutil.WriteFile(filepath.Join(tempDirPath, "filespec-test"), []byte("This is test file content."), 0644)
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
	expectedResult := "{\"files\": [{\"limit\":\"3\",\"pattern\":\"testPattern\"},{\"pattern\":\"testAql\",\"props\":\"a=b;c=d\",\"recursive\":\"true\"}]}"
	resultMap1 := map[string]interface{}{"pattern": "testPattern", "limit": "3"}
	resultMap2 := map[string]interface{}{"pattern": "testAql", "props": "a=b;c=d", "recursive": "true"}

	// Create Result-maps slice.
	var resultSlice []specslicetype
	resultSlice = append(resultSlice, resultMap1, resultMap2)

	// Test produced file-spec json.
	actualResult, err := buildFileSpecJson(resultSlice)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, string(actualResult))
}
