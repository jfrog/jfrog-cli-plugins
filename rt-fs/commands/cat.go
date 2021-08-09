package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"

	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

const (
	downloadThreads      = 1
	downloadMinSplitSize = 5120
	downloadSplitCount   = 3
)

func GetCatCommand() components.Command {
	return components.Command{
		Name:        "cat",
		Description: "Run cat.",
		Arguments:   getCommonArguments(),
		Flags:       getCommonFlags(),
		Action: func(c *components.Context) error {
			return catCmd(c)
		},
	}
}

func catCmd(c *components.Context) error {
	conf, err := createCommonConfiguration(c)
	if err != nil {
		return err
	}

	if strings.HasSuffix(conf.path, "/") || !strings.ContainsAny(conf.path, "/") {
		return errors.New("cat: " + conf.path + " : Path must be in a form of `<repo>/<name>` or `<repo>/<dir>/<name>`.")
	}

	return doCat(conf)
}

func doCat(c *commonConfiguration) error {
	// Create a temporary file for the download results
	target, err := ioutil.TempFile("", "rt-fs-cat")
	if err != nil {
		return err
	}
	if err = target.Close(); err != nil {
		return err
	}
	defer os.Remove(target.Name())

	// Download file
	downloadCmd := generic.NewDownloadCommand().SetConfiguration(createDownloadConfiguration()).SetBuildConfiguration(new(utils.BuildConfiguration))
	downloadCmd.SetServerDetails(c.details).SetSpec(createDownloadSpec(c, target.Name()))
	if err := commands.Exec(downloadCmd); err != nil {
		return err
	}

	// If no file downloaded, return "no such file" error
	if downloadCmd.Result().SuccessCount() != 1 {
		return errors.New("cat: " + c.path + ": No such file.")
	}

	// Print results
	bytes, err := ioutil.ReadFile(target.Name())
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))

	return nil
}

func createDownloadConfiguration() *utils.DownloadConfiguration {
	return &utils.DownloadConfiguration{
		Threads:      downloadThreads,
		MinSplitSize: downloadMinSplitSize,
		SplitCount:   downloadSplitCount,
	}
}

func createDownloadSpec(c *commonConfiguration, target string) *spec.SpecFiles {
	return &spec.SpecFiles{
		Files: []spec.File{
			{
				Aql:    createAql(c.path),
				Target: target,
				Flat:   "true",
			},
		},
	}
}

func createAql(downloadPath string) clientutils.Aql {
	splitRes := strings.Split(downloadPath, "/")
	repo := splitRes[0]
	name := splitRes[len(splitRes)-1]

	var path string
	if len(splitRes) > 2 {
		path = downloadPath[len(repo)+1 : len(downloadPath)-len(name)-1]
	} else {
		path = "."
	}

	return clientutils.Aql{
		ItemsFind: fmt.Sprintf(`{"repo":"%s","path":"%s","name":"%s"}`, repo, path, name),
	}
}
