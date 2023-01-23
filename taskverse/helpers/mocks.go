package helpers

import (
	"github.com/stretchr/testify/mock"
	"taskverse/helpers/runners"
)

type IntegrationParserMock struct {
	mock.Mock
}

func (m *IntegrationParserMock) Parse(options *ParseIntegrationsOptions) error {
	args := m.Called(options)
	return args.Error(0)
}

func (m *IntegrationParserMock) GetIntegrations() map[string]ProjectIntegration {
	args := m.Called()
	return args.Get(0).(map[string]ProjectIntegration)
}

func (m *IntegrationParserMock) GetSimplifiedIntegrations() map[string]map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]map[string]interface{})
}

type DependenciesDownloaderMock struct {
	mock.Mock
}

func (m *DependenciesDownloaderMock) Download(options *DependenciesDownloadOptions) error {
	args := m.Called(options)
	return args.Error(0)
}

type RunnerMock struct {
	mock.Mock
}

func (m *RunnerMock) GetRuntimeConfiguration() *runners.RuntimeConfiguration {
	args := m.Called()
	return args.Get(0).(*runners.RuntimeConfiguration)
}

func (m *RunnerMock) Run(options *runners.RunnerOptions) error {
	args := m.Called(options)
	return args.Error(0)
}

type StepJsonAssemblerMock struct {
	mock.Mock
}

func (m *StepJsonAssemblerMock) Assemble(options *StepJsonAssembleOptions) ([]byte, error) {
	args := m.Called(options)
	content := args.Get(0).([]byte)
	err := args.Error(1)
	return content, err
}

type ScriptAssemblerMock struct {
	mock.Mock
}

func (m *ScriptAssemblerMock) Assemble(options *AssembleOptions) ([]byte, error) {
	args := m.Called(options)
	content := args.Get(0).([]byte)
	err := args.Error(1)
	return content, err
}
