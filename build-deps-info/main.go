package main

import (
	"errors"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-plugins/build-deps-info/src/cmd/builddepsinfo"
)

func main() {
	plugins.PluginMain(components.App{
		Name:        "build-deps-info",
		Description: "Show the dependencies build and vcs details of a specific build name & build number from Artifactory.",
		Version:     "v0.1.0",
		Commands: []components.Command{
			{
				Name:        "print",
				Description: "Print the dependencies of a build",
				Aliases:     []string{"p"},
				Arguments:   getPrintArguments(),
				Flags:       getPrintFlags(),
				EnvVars:     []components.EnvVar{},
				Action: func(c *components.Context) error {
					if len(c.Arguments) != 2 {
						return errors.New("Wrong number of arguments. Expected: 2, " + "Received: " + strconv.Itoa(len(c.Arguments)))
					}
					rtDetails, err := config.GetDefaultArtifactoryConf()
					if err != nil {
						return err
					}
					servicesManager, err := utils.CreateServiceManager(rtDetails, false)
					if err != nil {
						return err
					}
					bdInfo := builddepsinfo.NewBuildDepsInfo().SetBuildName(c.Arguments[0]).SetBuildNumber(c.Arguments[1]).SetRepository(c.GetStringFlagValue("repo")).SetServicesManager(servicesManager)
					return bdInfo.Exec()
				},
			},
		},
	})
}

func getPrintArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "build-name",
			Description: "The name of the build you would like to print.",
		},
		{
			Name:        "build-number",
			Description: "The number of the build name you would like to print.",
		},
	}
}

func getPrintFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:         "repo",
			Description:  "In which repository in artifactory the dependencies is located",
			DefaultValue: "",
		},
	}
}
