package commands

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"strings"
	"taskverse/constants"
	"taskverse/helpers"
	"taskverse/helpers/runners"
)

func GetRunCommand() components.Command {
	return components.Command{
		Name:        "run",
		Description: "Run Pipeline Task.",
		Aliases:     []string{"r"},
		Arguments:   getRunArguments(),
		Flags:       getRunFlags(),
		EnvVars:     getRunEnvVar(),
		Action: func(c *components.Context) error {
			return runCmd(c)
		},
	}
}

func getRunArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "path_to_task",
			Description: "Path to the task folder",
		},
		{
			Name:        "arg",
			Description: "Arguments to the task in the format --arg=\"key=value\"",
		},
		{
			Name:        "env",
			Description: "Environment variables in the format --env=\"key=value\"",
		},
	}
}

func getRunFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:         "integrations",
			Description:  "Path to file containing integrations to be added to the environment",
			DefaultValue: "",
			Mandatory:    false,
		},
		components.BoolFlag{
			Name:         "on-step-complete",
			Description:  "Enable onStepComplete hook execution",
			DefaultValue: false,
		},
		components.StringFlag{
			Name:         "post-task-script",
			Description:  "Path to a script to be executed after the task is done",
			DefaultValue: "",
			Mandatory:    false,
		},
	}
}

func getRunEnvVar() []components.EnvVar {
	return []components.EnvVar{}
}

type runConfiguration struct {
	pathToTask             string
	pathToIntegrationsFile string
	args                   map[string]string
	envs                   map[string]string
	onStepComplete         bool
	pathToPostTaskScript   string
}

func runCmd(c *components.Context) error {
	conf, err := buildRunConfiguration(c)
	if err != nil {
		return err
	}

	// Initialize Dependencies
	integrationParser := helpers.NewIntegrationParser()
	dependenciesDownloader := helpers.NewDependenciesDownloader()
	runner := runners.NewDockerRunner()
	stepJsonAssembler := helpers.NewStepJsonAssembler(integrationParser)
	scriptAssembler := helpers.NewScriptAssembler(runner, integrationParser)

	doRun(conf, integrationParser, dependenciesDownloader, runner, stepJsonAssembler, scriptAssembler)
	return nil
}

func buildRunConfiguration(c *components.Context) (*runConfiguration, error) {
	if len(c.Arguments) < 1 {
		return nil, errors.New("invalid number of arguments")
	}

	var conf = &runConfiguration{
		pathToTask: "",
		args:       make(map[string]string),
		envs:       make(map[string]string),
	}

	conf.pathToTask = helpers.ResolvePathAndPanicIfNotFound(c.Arguments[0])

	err := validatePathToTask(conf)
	if err != nil {
		return nil, err
	}

	err = parseOptionalArguments(c, conf)
	if err != nil {
		return nil, err
	}

	integrationsFlag := c.GetStringFlagValue("integrations")
	if integrationsFlag != "" {
		conf.pathToIntegrationsFile = helpers.ResolvePathAndPanicIfNotFound(integrationsFlag)
	}

	postTaskScriptFlag := c.GetStringFlagValue("post-task-script")
	if postTaskScriptFlag != "" {
		conf.pathToPostTaskScript = helpers.ResolvePathAndPanicIfNotFound(postTaskScriptFlag)
	}

	onStepCompleteFlag := c.GetBoolFlagValue("on_step_complete")
	conf.onStepComplete = onStepCompleteFlag

	return conf, nil
}

func validatePathToTask(conf *runConfiguration) error {
	stat, err := os.Stat(conf.pathToTask)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("pathToTask argument should be a directory")
	}
	return nil
}

func parseOptionalArguments(c *components.Context, conf *runConfiguration) error {
	for i := 1; i < len(c.Arguments); i++ {
		argument := c.Arguments[i]
		log.Debug(fmt.Sprintf("Parsing argument: %s", argument))
		if strings.HasPrefix(argument, "--arg=") {
			log.Debug(fmt.Sprintf("Input argument found: %s", argument))
			key, value, err := parseKeyValueArgument(argument)
			if err != nil {
				return err
			}
			conf.args[key] = value
		} else if strings.HasPrefix(argument, "--env=") {
			log.Debug(fmt.Sprintf("Environment variable argument found: %s", argument))
			key, value, err := parseKeyValueArgument(argument)
			if err != nil {
				return err
			}
			conf.envs[key] = value
		} else {
			return fmt.Errorf("invalid argument: %s", argument)
		}
	}
	return nil
}

func parseKeyValueArgument(argument string) (key string, value string, err error) {
	input := argument[6:]
	if input == "" {
		return "", "", fmt.Errorf("invalid argument: %s", argument)
	}
	keyValuePair := strings.SplitN(input, "=", 2)
	if len(keyValuePair) != 2 {
		return "", "", fmt.Errorf("invalid argument: %s", argument)
	}
	parsedValue := keyValuePair[1]
	parsedValue = strings.Trim(parsedValue, ` "'`)
	return keyValuePair[0], parsedValue, nil
}

func doRun(c *runConfiguration,
	integrationParser helpers.IntegrationsParser,
	dependenciesDownloader helpers.DependenciesDownloader,
	runner runners.Runner,
	stepJsonAssembler helpers.StepJsonAssembler,
	scriptAssembler helpers.ScriptAssembler) {

	// Setup resources folder
	resourcesDir, err := coreutils.GetJfrogPluginsResourcesDir(constants.PLUGIN_NAME)

	// Parse integrations
	if c.pathToIntegrationsFile != "" {
		parseOptions := &helpers.ParseIntegrationsOptions{
			PathToIntegrationsFile: c.pathToIntegrationsFile,
		}
		err := integrationParser.Parse(parseOptions)
		coreutils.PanicOnError(err)
	}

	// Fetch and cache dependencies
	dependenciesDir := filepath.Join(resourcesDir, "dependencies")
	downloadOptions := &helpers.DependenciesDownloadOptions{
		TargetFolder: dependenciesDir,
		Output:       os.Stdout,
	}
	err = dependenciesDownloader.Download(downloadOptions)
	coreutils.PanicOnError(err)

	// Setup developer environment
	wd, err := os.Getwd()
	coreutils.PanicOnError(err)
	localDeveloperFolder := filepath.Join(wd, constants.TOOL_FOLDER)

	// Assemble step json
	stepJsonOpts := &helpers.StepJsonAssembleOptions{}
	stepJson, err := stepJsonAssembler.Assemble(stepJsonOpts)
	coreutils.PanicOnError(err)

	// Assemble script
	assembleOpts := &helpers.AssembleOptions{
		StepJson:                 stepJson,
		TaskArguments:            c.args,
		EnvironmentVariables:     c.envs,
		EnableOnStepCompleteHook: c.onStepComplete,
	}
	script, err := scriptAssembler.Assemble(assembleOpts)
	coreutils.PanicOnError(err)

	runOptions := &runners.RunnerOptions{
		PathToTask:                 c.pathToTask,
		PathToDependencies:         dependenciesDir,
		PathToLocalDeveloperFolder: localDeveloperFolder,
		PathToWorkingDirectory:     wd,
		PathToPostTaskScript:       c.pathToPostTaskScript,
		Script:                     script,
		StepJson:                   stepJson,
		Output:                     os.Stdout,
	}
	err = runner.Run(runOptions)
	if err != nil {
		log.Error(fmt.Errorf("task run failed: %w", err))
		return
	}
}
