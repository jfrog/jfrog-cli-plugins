package main

import (
	"errors"
	"strconv"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	builddepsinfo "github.com/jfrog/jfrog-cli-plugins/build-deps-info/commands"
)

func main() {
	plugins.PluginMain(components.CreateApp("build-deps-info", "v1.2.4", "Print the dependencies build and a link to vcs of a specific build name & build number in Artifactory.",
		[]components.Command{{
			Name:        "show",
			Description: "Show the details of the build dependencies",
			Aliases:     []string{"s"},
			Arguments:   getShowArguments(),
			Flags:       getShowFlags(),
			EnvVars:     []components.EnvVar{},
			Action: func(c *components.Context) error {
				if len(c.Arguments) != 2 {
					return errors.New("wrong number of arguments. Expected: 2, " + "Received: " + strconv.Itoa(len(c.Arguments)))
				}
				rtDetails, err := commands.GetConfig(c.GetStringFlagValue("server-id"), true)
				if err != nil {
					return err
				}
				servicesManager, err := utils.CreateServiceManager(rtDetails, -1, 0, false)
				if err != nil {
					return err
				}
				bdInfo := builddepsinfo.NewBuildDepsInfo().SetBuildName(c.Arguments[0]).SetBuildNumber(c.Arguments[1]).SetRepository(c.GetStringFlagValue("repo")).SetServicesManager(servicesManager)
				return bdInfo.Exec()
			},
		},
		},
	))
}

func getShowArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "build-name",
			Description: "The name of the build you would like to show.",
		},
		{
			Name:        "build-number",
			Description: "The number of the build name you would like to show.",
		},
	}
}

func getShowFlags() []components.Flag {
	return []components.Flag{
		components.NewStringFlag("repo", "In which repository in artifactory the dependencies is located"),
		components.NewStringFlag("server-id", "Artifactory server ID configured using the config command. If not specified, the default configured Artifactory server is used."),
	}
}
