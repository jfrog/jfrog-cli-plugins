package commands

import (
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-plugins/file-spec-gen/commands/create"
)

func GetFileSpecGenCommand() components.Command {
	return components.Command{
		Name:        "create",
		Description: "Generates a file-spec json.",
		Aliases:     []string{"cr"},
		Flags:       getCreateFlags(),
		Action: func(c *components.Context) error {
			return create.Run(c)
		},
	}
}

func getCreateFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "file",
			Description: "Output generated file-spec to file.",
		},
	}
}
