package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"taskverse/constants"
	"taskverse/helpers"
	"taskverse/helpers/runners"
	"testing"
)

func TestInvalidatePathToTask(t *testing.T) {
	conf := &runConfiguration{
		pathToTask: "/invalid_path_to_task",
	}
	err := validatePathToTask(conf)
	assert.NotNil(t, err)
}

func TestValidatePathToTask(t *testing.T) {
	path, err := os.MkdirTemp(os.TempDir(), "taskverse_test_")
	defer os.RemoveAll(path)
	assert.Nil(t, err)
	conf := &runConfiguration{
		pathToTask: path,
	}
	err = validatePathToTask(conf)
	assert.Nil(t, err)
}

func TestParseOptionalArguments(t *testing.T) {
	context := &components.Context{
		Arguments: []string{
			"pathToTask",
			"--arg=key=value",
			"--env=key=value",
		},
	}
	conf := &runConfiguration{
		args: make(map[string]string),
		envs: make(map[string]string),
	}
	err := parseOptionalArguments(context, conf)
	assert.Nil(t, err)
}

func TestParseKeyValueArgument(t *testing.T) {
	tests := []struct {
		argument      string
		expectedKey   string
		expectedValue string
		expectedError error
	}{
		{"--arg=key=value", "key", "value", nil},
		{"--arg=key=value_with_=_value", "key", "value_with_=_value", nil},
		{"--arg=key=\"value_with_quotes\"", "key", "value_with_quotes", nil},
		{"--arg=key='value_with_single_quotes'", "key", "value_with_single_quotes", nil},
		{"--arg=key=value with spaces", "key", "value with spaces", nil},
		{"key=value", "", "", errors.New("invalid argument: key=value")},
		{"--arg=", "", "", errors.New("invalid argument: --arg=")},
	}
	for _, test := range tests {
		gotKey, gotValue, gotError := parseKeyValueArgument(test.argument)
		assert.Equal(t, test.expectedKey, gotKey)
		assert.Equal(t, test.expectedValue, gotValue)
		assert.Equal(t, test.expectedError, gotError)
	}
}

func TestInvalidArguments(t *testing.T) {
	context := &components.Context{
		Arguments: []string{},
	}
	gotConf, err := buildRunConfiguration(context)
	assert.Nil(t, gotConf)
	assert.Equal(t, err, errors.New("invalid number of arguments"))
}

func TestBuildRunConfiguration(t *testing.T) {
	// Create temporary folder
	path, err := os.MkdirTemp(os.TempDir(), "taskverse_test_")
	assert.Nil(t, err)
	defer os.RemoveAll(path)
	os.Chdir(path)

	// Setup expectations
	currentFolder, err := os.Getwd()
	assert.Nil(t, err)
	taskFolder := currentFolder
	expectedConf := &runConfiguration{
		pathToTask:             taskFolder,
		pathToIntegrationsFile: "",
		args:                   make(map[string]string),
		envs:                   make(map[string]string),
		onStepComplete:         false,
		pathToPostTaskScript:   "",
	}

	context := &components.Context{
		Arguments: []string{
			taskFolder,
		},
	}

	gotConf, err := buildRunConfiguration(context)

	assert.Nil(t, err)
	assert.NotNil(t, gotConf)
	assert.Equal(t, gotConf, expectedConf)
}

func TestDoRun(t *testing.T) {
	// Create temporary folder
	path, err := os.MkdirTemp(os.TempDir(), "taskverse_test_")
	assert.Nil(t, err)
	defer os.RemoveAll(path)
	os.Chdir(path)

	// Setup expected folders
	currentFolder, err := os.Getwd()
	assert.Nil(t, err)
	expectedLocalDeveloperFolder := filepath.Join(currentFolder, ".taskverse")
	expectedWorkingDirectory := currentFolder

	// Setup run configuration
	pathToTask := "/task"
	pathToIntegrationsFile := "integrations.json"
	output := os.Stdout
	runConf := &runConfiguration{
		pathToTask:             pathToTask,
		pathToIntegrationsFile: pathToIntegrationsFile,
	}

	// Initialize helper mocks and expected method calls
	integrationParser := new(helpers.IntegrationParserMock)
	expectedParseIntegrationOptions := &helpers.ParseIntegrationsOptions{
		PathToIntegrationsFile: pathToIntegrationsFile,
	}
	integrationParser.On("Parse", expectedParseIntegrationOptions).Return(nil)

	dependenciesDownloader := new(helpers.DependenciesDownloaderMock)
	resourcesDir, err := coreutils.GetJfrogPluginsResourcesDir(constants.PLUGIN_NAME)
	if err != nil {
		t.Error(err)
	}
	dependenciesDir := filepath.Join(resourcesDir, "dependencies")
	expectedDownloadOptions := &helpers.DependenciesDownloadOptions{
		TargetFolder: dependenciesDir,
		Output:       output,
	}
	dependenciesDownloader.On("Download", expectedDownloadOptions).Return(nil)

	runner := new(helpers.RunnerMock)
	expectedRunOptions := &runners.RunnerOptions{
		PathToTask:                 pathToTask,
		PathToDependencies:         dependenciesDir,
		PathToLocalDeveloperFolder: expectedLocalDeveloperFolder,
		PathToWorkingDirectory:     expectedWorkingDirectory,
		PathToPostTaskScript:       "",
		Script:                     []byte{},
		StepJson:                   []byte{},
		Output:                     output,
	}
	runner.On("Run", expectedRunOptions).Return(nil)

	stepJsonAssembler := new(helpers.StepJsonAssemblerMock)
	expectedStepJsonAssembleOptions := &helpers.StepJsonAssembleOptions{}
	stepJsonAssembler.On("Assemble", expectedStepJsonAssembleOptions).Return([]byte{}, nil)

	scriptAssembler := new(helpers.ScriptAssemblerMock)
	expectedAssembleOpts := &helpers.AssembleOptions{
		StepJson:                 []byte{},
		TaskArguments:            runConf.args,
		EnvironmentVariables:     runConf.envs,
		EnableOnStepCompleteHook: false,
	}
	scriptAssembler.On("Assemble", expectedAssembleOpts).Return([]byte{}, nil)

	// Execute run command
	doRun(runConf, integrationParser, dependenciesDownloader, runner, stepJsonAssembler, scriptAssembler)

	// Assert mock calls
	integrationParser.AssertExpectations(t)
	dependenciesDownloader.AssertExpectations(t)
	scriptAssembler.AssertExpectations(t)
	stepJsonAssembler.AssertExpectations(t)
	runner.AssertExpectations(t)
}
