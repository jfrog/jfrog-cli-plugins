package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"strings"
	"taskverse/constants"
	"taskverse/helpers/runners"
	"taskverse/helpers/templates"
	"text/template"
)

type AssembleOptions struct {
	TaskArguments            map[string]string
	EnvironmentVariables     map[string]string
	StepJson                 []byte
	stepJsonMap              map[string]interface{}
	EnableOnStepCompleteHook bool
}

func (o *AssembleOptions) GetValueFromStepJson(key string) string {
	if o.StepJson == nil {
		panic("StepJson not found")
	}

	if o.stepJsonMap == nil {
		o.stepJsonMap = make(map[string]interface{})
		err := json.Unmarshal(o.StepJson, &o.stepJsonMap)
		coreutils.PanicOnError(err)
	}

	value, err := o.getValueFromMapRecursively(o.stepJsonMap, key)
	if err != nil {
		panic(fmt.Errorf("stepJson file not found for key %s", key))
	}
	return value
}

func (o *AssembleOptions) getValueFromMapRecursively(data map[string]interface{}, key string) (string, error) {
	keyParts := strings.SplitN(key, ".", 2)
	newData := data[keyParts[0]]
	if newData == nil {
		return "", errors.New("value not found")
	}
	if len(keyParts) == 1 {
		return fmt.Sprintf("%v", newData), nil
	}
	return o.getValueFromMapRecursively(newData.(map[string]interface{}), keyParts[1])
}

type ScriptAssembler interface {
	Assemble(options *AssembleOptions) ([]byte, error)
}

type ScriptAssemblerImpl struct {
	runner               runners.Runner
	runtimeConfiguration *runners.RuntimeConfiguration
	integrationsParser   IntegrationsParser
}

func NewScriptAssembler(runner runners.Runner, integrationsParser IntegrationsParser) ScriptAssembler {
	return &ScriptAssemblerImpl{
		runner:             runner,
		integrationsParser: integrationsParser,
	}
}

func (s *ScriptAssemblerImpl) Assemble(options *AssembleOptions) ([]byte, error) {
	log.Info("Assembling steplet script")

	s.runtimeConfiguration = s.runner.GetRuntimeConfiguration()

	//TODO Add Windows support
	scriptTemplate, err := template.ParseFS(templates.TemplateFiles, "resources/script.sh")
	if err != nil {
		return nil, err
	}

	data := assembleScriptTemplateParams{
		PathToDependencies:       s.runtimeConfiguration.PathToDependencies,
		PathToTask:               s.runtimeConfiguration.PathToTask,
		EnvironmentVariables:     s.assembleEnvironmentVariables(options),
		EnableOnStepCompleteHook: options.EnableOnStepCompleteHook,
	}

	// Create task input arguments
	argumentsBuffer := bytes.Buffer{}
	for key, value := range options.TaskArguments {
		argumentsBuffer.WriteString(fmt.Sprintf("--arg %s=%s ", key, value))
	}
	data.TaskParameters = data.TaskParameters + " " + argumentsBuffer.String()

	// Add user defined environment variables
	for key, value := range options.EnvironmentVariables {
		data.EnvironmentVariables[key] = value
	}

	// Execute template
	var scriptBuffer bytes.Buffer
	scriptTemplate.Execute(&scriptBuffer, data)
	return scriptBuffer.Bytes(), nil
}

func (s *ScriptAssemblerImpl) assembleEnvironmentVariables(options *AssembleOptions) map[string]string {
	envVars := map[string]string{
		"JFROG_DEV_TOOL_FOLDER":           s.runtimeConfiguration.PathToDeveloperFolder,
		"JFROG_DEV_POST_TASK_SCRIPT_PATH": s.runtimeConfiguration.PathToPostTaskScript,
		"JFROG_TASK_WORKSPACE_DIR":        "/task_workspace",
		"script_extension":                s.runtimeConfiguration.ScriptExtension,
		"JFROG_SCRIPT_EXTENSION":          s.runtimeConfiguration.ScriptExtension,
		"JFROG_DEFAULT_TASK_REPO":         "pipelines-tasks-virtual",
		"step_json_path":                  s.runtimeConfiguration.PathToStepJsonFile,
		"JFROG_STEP_JSON_PATH":            s.runtimeConfiguration.PathToStepJsonFile,
		"step_workspace_dir":              filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step"),
		"JFROG_STEP_WORKSPACE_DIR":        filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step"),
		"operating_system":                s.runtimeConfiguration.Os,
		"JFROG_OPERATING_SYSTEM":          s.runtimeConfiguration.Os,
		"JFROG_OPERATING_SYSTEM_FAMILY":   s.runtimeConfiguration.OsFamily,
		"architecture":                    s.runtimeConfiguration.Architecture,
		"JFROG_ARCHITECTURE":              s.runtimeConfiguration.Architecture,
		"steplet_script_path":             s.runtimeConfiguration.PathToStepletScript,
		"reqexec_bin_path":                "",
		"run_dir":                         filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "run"),
		"step_dependency_state_dir":       filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "dependency_state"),
		"step_output_dir":                 filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "output"),
		"steplet_workspace_dir":           filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "steplet"),
		"JFROG_STEPLET_WORKSPACE_DIR":     filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "steplet"),
		"JFROG_POST_HOOKS_DIR":            filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "steplet", "post_hooks"),
		"pipeline_workspace_dir":          filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "pipeline"),
		"JFROG_PIPELINE_WORKSPACE_DIR":    filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "pipeline"),
		"JFROG_PIPELINE_VARIABLES_FILE":   filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "pipeline", "pipeline.env"),
		"shared_workspace":                filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "shared"),
		"JFROG_SHARED_WORKSPACE_DIR":      filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "shared"),
		"step_tmp_dir":                    filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "tmp"),
		"JFROG_STEP_TMP_DIR":              filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "tmp"),
		"step_workspace_tmp_dir":          filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "step", "tmp"),
		"reqexec_dir":                     "",
		"pipelines_api_url":               "http://localhost:8082/pipelines/api",
		"JFROG_PIPELINES_API_URL":         "http://localhost:8082/pipelines/api",
		"builder_api_token":               "1234",
		"JFROG_BUILDER_API_TOKEN":         "1234",
		"no_verify_ssl":                   "false",
		"step_docker_container_name":      s.runtimeConfiguration.ContainerName,
		"custom_certs_dir":                "/certs",
		"custom_certs_enabled":            "false",
		"JFROG_CLI_USER_AGENT":            constants.PLUGIN_NAME,
		"CI":                              "true",
		"JFROG_PIPE_CLI_PATH":             filepath.Join(s.runtimeConfiguration.PathToDependencies, "pipe", "pipe"),
		"jfrog_cli_path":                  filepath.Join(s.runtimeConfiguration.PathToDependencies, "jfrog2", "jfrog"),
		"step_id":                         options.GetValueFromStepJson("step.id"),
		"steplet_id":                      "1",
		"step_name":                       options.GetValueFromStepJson("step.name"),
		"step_namespace":                  "jfrog",
		"IS_K8S_BUILD_PLANE":              "false",
		"step_type":                       options.GetValueFromStepJson("step.type"),
		"step_url":                        "http://localhost:8082/ui/pipelines/myPipelines/project/pipeline/1/step?branch=main",
		"JFROG_CLI_BUILD_URL":             "http://localhost:8082/ui/pipelines/myPipelines/project/pipeline/1/step?branch=main",
		"JFROG_PIPELINES_DEBUG":           "false",
		"step_timeout_seconds":            "600",
		"step_runtime":                    options.GetValueFromStepJson("step.configuration.runtime.type"),
		"step_image_name":                 options.GetValueFromStepJson("step.configuration.runtime.image.imageName"),
		"step_image_tag":                  options.GetValueFromStepJson("step.configuration.runtime.image.imageTag"),
		"step_affinity_group":             options.GetValueFromStepJson("step.configuration.affinityGroup"),
		"step_node_pool_name":             options.GetValueFromStepJson("step.configuration.nodePool"),
		"step_node_id":                    options.GetValueFromStepJson("step.configuration.nodeId"),
		"step_node_name":                  options.GetValueFromStepJson("step.configuration.nodeName"),
		"step_triggered_by_identity_name": "user",
		"step_build_plane_version":        constants.BUILD_PLANE_VERSION,
		"step_platform":                   s.runtimeConfiguration.Os,
		"pipeline_name":                   "pipeline",
		"run_id":                          options.GetValueFromStepJson("step.runId"),
		"JFROG_RUN_ID":                    options.GetValueFromStepJson("step.runId"),
		"run_number":                      options.GetValueFromStepJson("step.runId"),
		"JFROG_RUN_NUMBER":                options.GetValueFromStepJson("step.runId"),
		"steplet_number":                  "1",
		"JFROG_STEPLET_NUMBER":            "1",
		"steplet_run_state_dir":           filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "run", "state"),
		"JFROG_STEPLET_RUN_STATE_DIR":     filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "run", "state"),
		"JFROG_RUN_VARIABLES_FILE":        filepath.Join(s.runtimeConfiguration.PathToDeveloperFolder, "run", "state", "run.env"),
		"project_name":                    "project",
		"JFROG_PROJECT_NAME":              "project",
		"project_id":                      "1",
		"JFROG_PROJECT_ID":                "1",
	}

	if os.Getenv("JFROG_CLI_LOG_LEVEL") == "DEBUG" {
		envVars["JFROG_PIPELINES_DEBUG"] = "true"
	}

	integrations := s.integrationsParser.GetIntegrations()
	for _, integration := range integrations {
		namePrefix := fmt.Sprintf("int_%s", ConvertFirstRuneToLowerCase(integration.Name))
		integrationVariables := integration.GetIntegrationAsEnvironmentVariable()
		for k, v := range integrationVariables {
			envVars[fmt.Sprintf("%s_%s", namePrefix, k)] = v
		}
	}

	return envVars
}

type assembleScriptTemplateParams struct {
	PathToDependencies       string
	PathToTask               string
	TaskParameters           string
	EnvironmentVariables     map[string]string
	EnableOnStepCompleteHook bool
}
