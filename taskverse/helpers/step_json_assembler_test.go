package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAssemble(t *testing.T) {
	integrationParser := new(IntegrationParserMock)
	integrationParser.On("GetIntegrations").Return(make(map[string]ProjectIntegration))
	integrationParser.On("GetSimplifiedIntegrations").Return(make(map[string]map[string]interface{}))

	expectedStepJson := `{"step":{"id":1,"name":"mock_step","runId":1,"pipelineId":1,"pipelineStepId":1,"type":"Bash","execution":{"onExecute":["echo task"]},"configuration":{"affinityGroup":"affinity_group","inputSteps":[],"inputResources":[],"outputResources":[],"integrations":[],"environmentVariables":[],"nodePool":"node_pool","timeoutSeconds":300,"runtime":{"type":"image","image":{"imageName":"releases-docker.jfrog.io/jfrog/pipelines-u20node","imageTag":16}},"isOnDemand":false,"instanceSize":null,"nodeId":1,"nodeName":"node"}},"resources":{},"integrations":{},"affinityGroupSteps":{},"inputStepIds":[]}`

	assembler := NewStepJsonAssembler(integrationParser)
	options := &StepJsonAssembleOptions{}
	gotStepJson, err := assembler.Assemble(options)

	integrationParser.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, expectedStepJson, string(gotStepJson))
}

func TestAssembleWithIntegration(t *testing.T) {
	integration := ProjectIntegration{
		Id:                    1,
		MasterIntegrationId:   1,
		Name:                  "integration",
		MasterIntegrationType: "generic",
		ProjectId:             1,
		MasterIntegrationName: "generic",
		ProviderId:            1,
		Environments:          nil,
		IsInternal:            false,
		CreatedByUserName:     "user",
		UpdatedByUserName:     "user",
		FormJSONValues: []FormJSONValues{
			{
				Label: "key",
				Value: "value",
			},
		},
		CreatedBy: 1,
		UpdatedBy: 1,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	integrationParser := new(IntegrationParserMock)
	integrationParser.On("GetIntegrations").Return(map[string]ProjectIntegration{"integration": integration})
	simplifiedResponse := make(map[string]map[string]interface{})
	simplifiedResponse[integration.Name] = make(map[string]interface{})
	simplifiedResponse[integration.Name]["id"] = integration.Id
	simplifiedResponse[integration.Name]["name"] = integration.Name
	simplifiedResponse[integration.Name]["masterName"] = integration.MasterIntegrationName
	simplifiedResponse[integration.Name]["displayName"] = integration.MasterIntegrationName
	for _, value := range integration.FormJSONValues {
		simplifiedResponse[integration.Name][value.Label] = value.Value
	}
	integrationParser.On("GetSimplifiedIntegrations").Return(simplifiedResponse)

	expectedStepJson := `{"step":{"id":1,"name":"mock_step","runId":1,"pipelineId":1,"pipelineStepId":1,"type":"Bash","execution":{"onExecute":["echo task"]},"configuration":{"affinityGroup":"affinity_group","inputSteps":[],"inputResources":[],"outputResources":[],"integrations":[{"name":"integration"}],"environmentVariables":[],"nodePool":"node_pool","timeoutSeconds":300,"runtime":{"type":"image","image":{"imageName":"releases-docker.jfrog.io/jfrog/pipelines-u20node","imageTag":16}},"isOnDemand":false,"instanceSize":null,"nodeId":1,"nodeName":"node"}},"resources":{},"integrations":{"integration":{"displayName":"generic","id":1,"key":"value","masterName":"generic","name":"integration"}},"affinityGroupSteps":{},"inputStepIds":[]}`

	assembler := NewStepJsonAssembler(integrationParser)
	options := &StepJsonAssembleOptions{}
	gotStepJson, err := assembler.Assemble(options)

	integrationParser.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, expectedStepJson, string(gotStepJson))
}
