package helpers

import (
	"encoding/json"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"os"
	"time"
)

type ProjectIntegration struct {
	Id                    int              `json:"id"`
	MasterIntegrationId   int              `json:"masterIntegrationId"`
	Name                  string           `json:"name"`
	MasterIntegrationType string           `json:"masterIntegrationType"`
	ProjectId             int              `json:"projectId"`
	MasterIntegrationName string           `json:"masterIntegrationName"`
	ProviderId            int              `json:"providerId"`
	Environments          interface{}      `json:"environments"`
	IsInternal            bool             `json:"isInternal"`
	CreatedByUserName     string           `json:"createdByUserName"`
	UpdatedByUserName     string           `json:"updatedByUserName"`
	FormJSONValues        []FormJSONValues `json:"formJSONValues"`
	CreatedBy             int              `json:"createdBy"`
	UpdatedBy             int              `json:"updatedBy"`
	CreatedAt             time.Time        `json:"createdAt"`
	UpdatedAt             time.Time        `json:"updatedAt"`
}

func (p ProjectIntegration) GetIntegrationAsEnvironmentVariable() map[string]string {
	envVars := GetStructAsEnvironmentVariables(p)
	for _, formJsonValue := range p.FormJSONValues {
		valueKey := ConvertFirstRuneToLowerCase(formJsonValue.Label)
		envVars[valueKey] = formJsonValue.Value
	}
	return envVars
}

type FormJSONValues struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ParseIntegrationsOptions struct {
	PathToIntegrationsFile string
}

type IntegrationsParser interface {
	Parse(*ParseIntegrationsOptions) error
	GetIntegrations() map[string]ProjectIntegration
	GetSimplifiedIntegrations() map[string]map[string]interface{}
}

type IntegrationsParserImpl struct {
	integrations []ProjectIntegration
}

func NewIntegrationParser() IntegrationsParser {
	return &IntegrationsParserImpl{
		integrations: []ProjectIntegration{},
	}
}

func (i *IntegrationsParserImpl) Parse(options *ParseIntegrationsOptions) error {
	log.Info("Parsing project integrations")
	jsonFile, err := os.Open(options.PathToIntegrationsFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	fileContent, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(fileContent, &i.integrations)
}

func (i *IntegrationsParserImpl) GetIntegrations() map[string]ProjectIntegration {
	integrationsMap := make(map[string]ProjectIntegration)
	for _, integration := range i.integrations {
		integrationsMap[integration.Name] = integration
	}
	return integrationsMap
}

func (i *IntegrationsParserImpl) GetSimplifiedIntegrations() map[string]map[string]interface{} {
	integrationsMap := make(map[string]map[string]interface{})
	for _, integration := range i.integrations {
		integrationsMap[integration.Name] = make(map[string]interface{})
		integrationsMap[integration.Name]["id"] = integration.Id
		integrationsMap[integration.Name]["name"] = integration.Name
		integrationsMap[integration.Name]["masterName"] = integration.MasterIntegrationName
		integrationsMap[integration.Name]["displayName"] = integration.MasterIntegrationName
		for _, value := range integration.FormJSONValues {
			integrationsMap[integration.Name][value.Label] = value.Value
		}
	}
	return integrationsMap
}
