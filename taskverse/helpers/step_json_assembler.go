package helpers

import (
	"encoding/json"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type StepJson struct {
	Step      Step `json:"step"`
	Resources struct {
	} `json:"resources"`
	Integrations       map[string]map[string]interface{} `json:"integrations"`
	AffinityGroupSteps struct {
	} `json:"affinityGroupSteps"`
	InputStepIds []string `json:"inputStepIds"`
}

type Step struct {
	Id             int           `json:"id"`
	Name           string        `json:"name"`
	RunId          int           `json:"runId"`
	PipelineId     int           `json:"pipelineId"`
	PipelineStepId int           `json:"pipelineStepId"`
	Type           string        `json:"type"`
	Execution      Execution     `json:"execution"`
	Configuration  Configuration `json:"configuration"`
}

type Execution struct {
	OnExecute []string `json:"onExecute"`
}

type Configuration struct {
	AffinityGroup        string                `json:"affinityGroup"`
	InputSteps           []NamedReference      `json:"inputSteps"`
	InputResources       []NamedReference      `json:"inputResources"`
	OutputResources      []NamedReference      `json:"outputResources"`
	Integrations         []NamedReference      `json:"integrations"`
	EnvironmentVariables []EnvironmentVariable `json:"environmentVariables"`
	NodePool             string                `json:"nodePool"`
	TimeoutSeconds       int                   `json:"timeoutSeconds"`
	Runtime              Runtime               `json:"runtime"`
	IsOnDemand           bool                  `json:"isOnDemand"`
	InstanceSize         interface{}           `json:"instanceSize"`
	NodeId               int                   `json:"nodeId"`
	NodeName             string                `json:"nodeName"`
}

type Runtime struct {
	Type  string `json:"type"`
	Image Image  `json:"image"`
}

type Image struct {
	ImageName string `json:"imageName"`
	ImageTag  int    `json:"imageTag"`
}

type NamedReference struct {
	Name string `json:"name"`
}

type EnvironmentVariable struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	IsReadOnly bool   `json:"isReadOnly"`
	Level      string `json:"level"`
}

type StepJsonAssembleOptions struct {
}

type StepJsonAssembler interface {
	Assemble(options *StepJsonAssembleOptions) ([]byte, error)
}

type StepJsonAssemblerImpl struct {
	integrationsParser IntegrationsParser
}

func NewStepJsonAssembler(integrationsParser IntegrationsParser) StepJsonAssembler {
	return &StepJsonAssemblerImpl{
		integrationsParser: integrationsParser,
	}
}

func (s *StepJsonAssemblerImpl) Assemble(options *StepJsonAssembleOptions) ([]byte, error) {
	log.Info("Assembling stepJson file")

	// Prepare integrations collections
	integrations := s.integrationsParser.GetIntegrations()
	simplifiedIntegrations := s.integrationsParser.GetSimplifiedIntegrations()
	namedIntegrations := []NamedReference{}
	for key, _ := range integrations {
		namedIntegrations = append(namedIntegrations, NamedReference{
			Name: key,
		})
	}

	stepJson := StepJson{
		Step: Step{
			Id:             1,
			Name:           "mock_step",
			RunId:          1,
			PipelineId:     1,
			PipelineStepId: 1,
			// TODO: Add windows support
			Type: "Bash",
			Execution: Execution{
				OnExecute: []string{"echo task"},
			},
			Configuration: Configuration{
				AffinityGroup:        "affinity_group",
				InputSteps:           []NamedReference{},
				InputResources:       []NamedReference{},
				OutputResources:      []NamedReference{},
				Integrations:         namedIntegrations,
				EnvironmentVariables: []EnvironmentVariable{},
				NodePool:             "node_pool",
				TimeoutSeconds:       300,
				Runtime: Runtime{
					Type: "image",
					Image: Image{
						ImageName: "releases-docker.jfrog.io/jfrog/pipelines-u20node",
						ImageTag:  16,
					},
				},
				IsOnDemand:   false,
				InstanceSize: nil,
				NodeId:       1,
				NodeName:     "node",
			},
		},
		Resources:          struct{}{},
		Integrations:       simplifiedIntegrations,
		AffinityGroupSteps: struct{}{},
		InputStepIds:       []string{},
	}

	return json.Marshal(stepJson)
}
