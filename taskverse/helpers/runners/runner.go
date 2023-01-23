package runners

import "io"

type RunnerOptions struct {
	PathToTask                 string
	PathToDependencies         string
	PathToLocalDeveloperFolder string
	PathToWorkingDirectory     string
	PathToPostTaskScript       string
	Script                     []byte
	StepJson                   []byte
	Output                     io.Writer
}

type RuntimeConfiguration struct {
	ContainerName         string
	PathToTask            string
	PathToDependencies    string
	PathToDeveloperFolder string
	PathToStepJsonFile    string
	PathToStepletScript   string
	PathToPostTaskScript  string
	Os                    string
	OsFamily              string
	ScriptExtension       string
	Architecture          string
}

type Runner interface {
	GetRuntimeConfiguration() *RuntimeConfiguration
	Run(options *RunnerOptions) error
}
