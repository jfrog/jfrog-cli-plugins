package helpers

import (
	"github.com/stretchr/testify/assert"
	"os"
	"taskverse/helpers/runners"
	"testing"
)

func TestScriptAssembler(t *testing.T) {
	integrationParser := new(IntegrationParserMock)
	integrationParser.On("GetIntegrations").Return(make(map[string]ProjectIntegration))

	runtimeConfiguration := &runners.RuntimeConfiguration{
		ContainerName:         "",
		PathToTask:            "",
		PathToDependencies:    "",
		PathToDeveloperFolder: "",
		PathToStepJsonFile:    "",
		PathToStepletScript:   "",
		PathToPostTaskScript:  "",
		Os:                    "",
		OsFamily:              "",
		ScriptExtension:       "",
		Architecture:          "",
	}
	runner := new(RunnerMock)
	runner.On("GetRuntimeConfiguration").Return(runtimeConfiguration)

	stepJson, err := os.ReadFile("testdata/stepJson.json")
	assert.Nil(t, err)

	scriptAsembler := NewScriptAssembler(runner, integrationParser)
	options := &AssembleOptions{
		TaskArguments:            nil,
		EnvironmentVariables:     nil,
		StepJson:                 stepJson,
		EnableOnStepCompleteHook: false,
	}
	script, err := scriptAsembler.Assemble(options)

	integrationParser.AssertExpectations(t)
	runner.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, script)
}
