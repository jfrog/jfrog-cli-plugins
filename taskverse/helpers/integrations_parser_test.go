package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIntegrationsParser(t *testing.T) {
	integrationsParser := NewIntegrationParser()
	options := &ParseIntegrationsOptions{
		PathToIntegrationsFile: "testdata/integrations.json",
	}
	err := integrationsParser.Parse(options)

	assert.Nil(t, err)

	expectedIntegrations := map[string]ProjectIntegration{
		"integration": {
			Id:                    1,
			MasterIntegrationId:   1,
			Name:                  "integration",
			MasterIntegrationType: "generic",
			ProjectId:             1,
			MasterIntegrationName: "generic",
			ProviderId:            0,
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
			CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	gotIntegrations := integrationsParser.GetIntegrations()
	assert.Equal(t, expectedIntegrations, gotIntegrations)

	expectedSimplifiedIntegrations := map[string]map[string]interface{}{
		"integration": {
			"id":          1,
			"name":        "integration",
			"masterName":  "generic",
			"displayName": "generic",
			"key":         "value",
		},
	}
	gotSimplifiedIntegrations := integrationsParser.GetSimplifiedIntegrations()
	assert.Equal(t, expectedSimplifiedIntegrations, gotSimplifiedIntegrations)

	expectedIntegrationEnvVars := map[string]string{
		"id":                    "1",
		"masterIntegrationId":   "1",
		"name":                  "integration",
		"masterIntegrationType": "generic",
		"providerId":            "0",
		"projectId":             "1",
		"environments":          "",
		"masterIntegrationName": "generic",
		"key":                   "value",
		"formJSONValues":        "[{\\\"label\\\":\\\"key\\\",\\\"value\\\":\\\"value\\\"}]",
		"formJSONValues_0":      "{\\\"label\\\":\\\"key\\\",\\\"value\\\":\\\"value\\\"}",
		"formJSONValues_len":    "1",
		"isInternal":            "false",
		"createdBy":             "1",
		"createdByUserName":     "user",
		"updatedBy":             "1",
		"updatedByUserName":     "user",
		"createdAt":             "2023-01-01T12:00:00Z",
		"updatedAt":             "2023-01-01T12:00:00Z",
	}
	gotIntegrationEnvVars := gotIntegrations["integration"].GetIntegrationAsEnvironmentVariable()
	assert.Equal(t, expectedIntegrationEnvVars, gotIntegrationEnvVars)
}
