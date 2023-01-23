package helpers

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"taskverse/constants"
)

type DependenciesDownloadOptions struct {
	TargetFolder string
	Output       io.Writer
}

type DependenciesDownloader interface {
	Download(options *DependenciesDownloadOptions) error
}

// TODO: Make it configurable in run command
const REQKICK_DOCKER_IMAGE = "releases-docker.jfrog.io/jfrog/pipelines-reqkick:" + constants.BUILD_PLANE_VERSION
const CONTAINER_NAME = constants.PLUGIN_NAME + "-reqKick"

type DependenciesDownloaderImpl struct {
	targetDir              string
	utilityFunctionsCached bool
	jfrogCliCached         bool
	pipeToolCached         bool
	dockerToolCached       bool
}

func NewDependenciesDownloader() DependenciesDownloader {
	return &DependenciesDownloaderImpl{}
}

func (d *DependenciesDownloaderImpl) Download(options *DependenciesDownloadOptions) error {
	err := d.createTargetDir(options)
	coreutils.PanicOnError(err)

	err = d.checkDependenciesAlreadyAvailable(options)
	coreutils.PanicOnError(err)

	if d.utilityFunctionsCached && d.jfrogCliCached && d.pipeToolCached {
		log.Info("All build plane dependencies are available")
	} else {
		log.Info("Downloading missing build plane dependencies")

		log.Info("Pulling reqKick docker image to fetch dependencies")
		err = d.pullReqKickImage(options)
		coreutils.PanicOnError(err)

		log.Info("Checking for stale reqKick containers from previous runs")
		err = d.removeReqKickContainer(options)
		coreutils.PanicOnError(err)

		defer func() {
			log.Info("Removing reqKick container")
			d.removeReqKickContainer(options)
		}()

		log.Info("Starting reqKick container")
		err = d.startReqKickContainer(options)
		coreutils.PanicOnError(err)

		if !d.utilityFunctionsCached {
			log.Info("Fetching utility functions script")
			err = d.copyUtilityFunctionsScript(options)
			coreutils.PanicOnError(err)
		}

		if !d.jfrogCliCached {
			log.Info("Fetching JFrog CLI")
			err = d.copyJFrogCLI(options)
			coreutils.PanicOnError(err)
		}

		if !d.pipeToolCached {
			log.Info("Fetching Pipe tool")
			err = d.copyPipeTool(options)
			coreutils.PanicOnError(err)
		}
	}

	if d.dockerToolCached {
		log.Info("Docker dependency is available")
	} else {
		log.Info("Downloading Docker client")

		err = d.downloadDocker()
		coreutils.PanicOnError(err)
	}

	return nil
}

func (d *DependenciesDownloaderImpl) createTargetDir(options *DependenciesDownloadOptions) error {
	d.targetDir = options.TargetFolder

	// Create empty folder
	err := os.MkdirAll(d.targetDir, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (d *DependenciesDownloaderImpl) checkDependenciesAlreadyAvailable(options *DependenciesDownloadOptions) error {
	pathToUtilityFunctionsScript := filepath.Join(d.targetDir, "header.sh")
	d.utilityFunctionsCached = fileutils.IsPathExists(pathToUtilityFunctionsScript, true)

	pathToJFrogCLI := filepath.Join(d.targetDir, "jfrog2", "jfrog")
	d.jfrogCliCached = fileutils.IsPathExists(pathToJFrogCLI, true)

	pathToPipeTool := filepath.Join(d.targetDir, "pipe", "pipe")
	d.pipeToolCached = fileutils.IsPathExists(pathToPipeTool, true)

	pathToDockerTool := filepath.Join(d.targetDir, "docker", "docker")
	d.dockerToolCached = fileutils.IsPathExists(pathToDockerTool, true)

	return nil
}

func (d *DependenciesDownloaderImpl) pullReqKickImage(options *DependenciesDownloadOptions) error {
	cmd := exec.Command("docker", "pull", REQKICK_DOCKER_IMAGE)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (d *DependenciesDownloaderImpl) removeReqKickContainer(options *DependenciesDownloadOptions) error {
	cmd := exec.Command("docker", "rm", "-f", CONTAINER_NAME)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (d *DependenciesDownloaderImpl) startReqKickContainer(options *DependenciesDownloadOptions) error {
	dockerRunParameters := []string{"run",
		"-d",
		"--name", CONTAINER_NAME,
		"--entrypoint",
		"tail",
		REQKICK_DOCKER_IMAGE,
		"-f", "/dev/null"}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(dockerRunParameters, " ")))
	cmd := exec.Command("docker", dockerRunParameters...)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (d *DependenciesDownloaderImpl) copyUtilityFunctionsScript(options *DependenciesDownloadOptions) error {
	source := "/jfrog-init/reqKick/node_modules/pipelines-core/execTemplates/steps/jfrog/v1.0/Bash/header.sh"
	target := filepath.Join(d.targetDir, "header.sh")
	return d.copyContentFromReqKickContainer(source, target, options)
}

func (d *DependenciesDownloaderImpl) copyJFrogCLI(options *DependenciesDownloadOptions) error {
	source := "/jfrog-init/jfrog"
	target := filepath.Join(d.targetDir, "jfrog")
	err := d.copyContentFromReqKickContainer(source, target, options)
	if err != nil {
		return err
	}

	source = "/jfrog-init/jfrog2"
	target = filepath.Join(d.targetDir, "jfrog2")
	return d.copyContentFromReqKickContainer(source, target, options)
}

func (d *DependenciesDownloaderImpl) copyPipeTool(options *DependenciesDownloadOptions) error {
	source := "/jfrog-init/pipe"
	target := filepath.Join(d.targetDir, "pipe")
	return d.copyContentFromReqKickContainer(source, target, options)
}

func (d *DependenciesDownloaderImpl) copyContentFromReqKickContainer(source string, target string, options *DependenciesDownloadOptions) error {
	// Remove target if exists
	err := os.RemoveAll(target)
	if err != nil {
		return err
	}

	dockerCopyParameters := []string{"cp",
		fmt.Sprintf("%s:%s", CONTAINER_NAME, source),
		target,
	}
	log.Debug(fmt.Sprintf("docker %s", strings.Join(dockerCopyParameters, " ")))
	cmd := exec.Command("docker", dockerCopyParameters...)
	cmd.Stdout = options.Output
	cmd.Stderr = options.Output
	return cmd.Run()
}

func (d *DependenciesDownloaderImpl) downloadDocker() error {
	// Create target folder
	err := os.MkdirAll(filepath.Join(d.targetDir, "docker"), os.ModePerm)
	if err != nil {
		return err
	}

	fileUrl := "https://download.docker.com/linux/static/stable/x86_64/docker-20.10.18.tgz"
	targetFile := filepath.Join(d.targetDir, "docker", "docker.tgz")
	err = DownloadFile(fileUrl, targetFile)
	if err != nil {
		return err
	}

	reader, err := os.Open(targetFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	extractTargetFolder := filepath.Join(d.targetDir, "docker")
	err = ExtractTarGz(reader, extractTargetFolder)
	if err != nil {
		return err
	}

	// Move docker content one level up
	err = os.Rename(filepath.Join(extractTargetFolder, "docker"), filepath.Join(d.targetDir, "docker_tmp"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(d.targetDir, "docker"))
	if err != nil {
		return err
	}
	return os.Rename(filepath.Join(d.targetDir, "docker_tmp"), filepath.Join(d.targetDir, "docker"))
}
