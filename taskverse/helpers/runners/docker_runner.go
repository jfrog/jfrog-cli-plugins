package runners

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"taskverse/constants"
	"time"
)

// TODO: Support docker image parameter
// TODO: Support windows
const DEFAULT_IMAGE = "releases-docker.jfrog.io/jfrog/pipelines-u20node:16"
const NETWORK_NAME = constants.PLUGIN_NAME
const CONTAINER_NAME = constants.PLUGIN_NAME + "-run"
const DIND_CONTAINER_NAME = constants.PLUGIN_NAME + "-dind"
const CONTAINER_WORK_DIR = "/workdir"
const CONTAINER_TASK_DIR = "/task"
const CONTAINER_DEPENDENCIES_DIR = "/dependencies"
const CONTAINER_POST_TASK_SCRIPT_PATH = "/post-task/script.sh"

type DockerRunner struct {
	dockerCommand        []string
	runtimeConfiguration *RuntimeConfiguration
}

func NewDockerRunner() *DockerRunner {
	return &DockerRunner{}
}

func (r *DockerRunner) GetRuntimeConfiguration() *RuntimeConfiguration {
	r.runtimeConfiguration = &RuntimeConfiguration{
		ContainerName:         CONTAINER_NAME,
		PathToTask:            CONTAINER_TASK_DIR,
		PathToDependencies:    CONTAINER_DEPENDENCIES_DIR,
		PathToDeveloperFolder: filepath.Join(CONTAINER_WORK_DIR, constants.TOOL_FOLDER),
		PathToStepJsonFile:    filepath.Join(CONTAINER_WORK_DIR, constants.TOOL_FOLDER, "step", "stepJson.json"),
		PathToStepletScript:   filepath.Join(CONTAINER_WORK_DIR, constants.TOOL_FOLDER, "step", "script.sh"),
		PathToPostTaskScript:  CONTAINER_POST_TASK_SCRIPT_PATH,
		Os:                    "Ubuntu_20.04",
		OsFamily:              "linux",
		ScriptExtension:       "sh",
		Architecture:          "x86_64",
	}
	return r.runtimeConfiguration
}

func (r *DockerRunner) Run(options *RunnerOptions) error {
	log.Info("Creating required folders and files at", options.PathToLocalDeveloperFolder)
	err := r.createFolders(options)
	coreutils.PanicOnError(err)

	err = r.writeScriptToFolder(options.Script, options)
	coreutils.PanicOnError(err)

	err = r.writeStepJsonToFolder(options.StepJson, options)
	coreutils.PanicOnError(err)

	log.Info("Checking docker network")
	dockerNetworkAvailable, err := r.checkDockerNetworkExists()
	coreutils.PanicOnError(err)
	if !dockerNetworkAvailable {
		log.Info(fmt.Sprintf("Creating docker network %s", NETWORK_NAME))
		err = r.createDockerNetwork(options)
		coreutils.PanicOnError(err)
	} else {
		log.Info(fmt.Sprintf("Docker network %s is available", NETWORK_NAME))
	}

	log.Info("Checking dind container")
	dindAvailable, err := r.checkDindContainerExists()
	coreutils.PanicOnError(err)
	if !dindAvailable {
		log.Info(fmt.Sprintf("Creating dind container %s", DIND_CONTAINER_NAME))
		err = r.bootDindContainer(options)
		coreutils.PanicOnError(err)
	} else {
		log.Info(fmt.Sprintf("Dind container %s is running", DIND_CONTAINER_NAME))
	}

	log.Info("Checking for stale containers from previous runs")
	err = r.removeStaleContainer(options)
	coreutils.PanicOnError(err)

	log.Info("Booting task container")
	r.createDockerRunArguments(options)
	log.Debug(fmt.Sprintf("docker %s", strings.Join(r.dockerCommand, " ")))
	return r.runDockerCommandAndStreamOutput(options)
}

func (r *DockerRunner) createFolders(options *RunnerOptions) error {
	// Remove content if exists
	err := os.RemoveAll(options.PathToLocalDeveloperFolder)
	if err != nil {
		return err
	}

	// Create empty folder
	err = os.MkdirAll(options.PathToLocalDeveloperFolder, os.ModePerm)
	if err != nil {
		return err
	}

	// Create task folder
	err = os.MkdirAll(filepath.Join(options.PathToLocalDeveloperFolder, "task"), os.ModePerm)
	if err != nil {
		return err
	}

	// Create step folder
	err = os.MkdirAll(filepath.Join(options.PathToLocalDeveloperFolder, "step"), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (r *DockerRunner) writeScriptToFolder(scriptContent []byte, options *RunnerOptions) error {
	pathToScript := filepath.Join(options.PathToLocalDeveloperFolder, "step", "script.sh")
	return os.WriteFile(pathToScript, scriptContent, os.ModePerm)
}

func (r *DockerRunner) writeStepJsonToFolder(stepJsonContent []byte, options *RunnerOptions) error {
	pathToStepJson := filepath.Join(options.PathToLocalDeveloperFolder, "step", "stepJson.json")
	return os.WriteFile(pathToStepJson, stepJsonContent, os.ModePerm)
}

func (r *DockerRunner) checkDockerNetworkExists() (bool, error) {
	getNetworkIdCommand := []string{"network", "ls", "-f", "name=" + constants.PLUGIN_NAME, "-q"}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(getNetworkIdCommand, " ")))
	cmd := exec.Command("docker", getNetworkIdCommand...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	log.Debug(fmt.Sprintf("Output: %s", output))
	return strings.TrimSpace(string(output)) != "", nil
}

func (r *DockerRunner) createDockerNetwork(options *RunnerOptions) error {
	createNetworkCommand := []string{"network", "create", NETWORK_NAME}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(createNetworkCommand, " ")))
	cmd := exec.Command("docker", createNetworkCommand...)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (r *DockerRunner) checkDindContainerExists() (bool, error) {
	getDindCommand := []string{"ps", "-f", "name=" + DIND_CONTAINER_NAME, "-q"}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(getDindCommand, " ")))
	cmd := exec.Command("docker", getDindCommand...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	log.Debug(fmt.Sprintf("Output: %s", output))
	return strings.TrimSpace(string(output)) != "", nil
}

func (r *DockerRunner) bootDindContainer(options *RunnerOptions) error {
	// Remove stale container from previous run
	cmd := exec.Command("docker", "rm", "-f", DIND_CONTAINER_NAME)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	err := cmd.Run()
	if err != nil {
		return err
	}

	runDindCommand := []string{
		"run",
		"--privileged",
		"--name", DIND_CONTAINER_NAME,
		"--network", NETWORK_NAME,
		"-d",
		"-e", "DOCKER_TLS_CERTDIR=",
		"--pull", "always",
		"docker:dind",
	}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(runDindCommand, " ")))
	cmd = exec.Command("docker", runDindCommand...)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Wait 15 seconds for dind to start
	log.Debug("Waiting 15 seconds for dind to start")
	time.Sleep(15 * time.Second)
	return nil
}

func (r *DockerRunner) createDockerRunArguments(options *RunnerOptions) error {
	r.dockerCommand = []string{
		"run",
		"-v", fmt.Sprintf("%s:%s", options.PathToTask, CONTAINER_TASK_DIR),
		"-v", fmt.Sprintf("%s:%s", options.PathToDependencies, CONTAINER_DEPENDENCIES_DIR),
		"-v", fmt.Sprintf("%s:%s", options.PathToWorkingDirectory, CONTAINER_WORK_DIR),
		"-w", fmt.Sprintf("%s", CONTAINER_WORK_DIR),
		"--init",
		"--name", CONTAINER_NAME,
		"--network", NETWORK_NAME,
		"-e", fmt.Sprintf("DOCKER_HOST=%s", DIND_CONTAINER_NAME),
		"--rm",
		"--pull", "always",
	}

	if options.PathToPostTaskScript != "" {
		r.dockerCommand = append(r.dockerCommand,
			"-v", fmt.Sprintf("%s:%s", options.PathToPostTaskScript, CONTAINER_POST_TASK_SCRIPT_PATH))
	}

	r.dockerCommand = append(r.dockerCommand,
		DEFAULT_IMAGE,
		"bash", "-c", fmt.Sprintf("%s/%s/script.sh", r.runtimeConfiguration.PathToDeveloperFolder, "step"),
	)

	return nil
}

func (r *DockerRunner) removeStaleContainer(options *RunnerOptions) error {
	cmd := exec.Command("docker", "rm", "-f", CONTAINER_NAME)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (r *DockerRunner) runDockerCommandAndStreamOutput(options *RunnerOptions) error {
	cmd := exec.Command("docker", r.dockerCommand...)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}
