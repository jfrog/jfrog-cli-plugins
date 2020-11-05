package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/zalando/go-keyring"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
)

func GetDeleteCommand() components.Command {
	return components.Command{
		Name:        "delete",
		Description: "Delete Artifactory configuration from the system keyring.",
		Aliases:     []string{"del"},
		Arguments:   getDeleteArguments(),
		Action: func(c *components.Context) error {
			return deleteCmd(c)
		},
	}
}

func getDeleteArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "server ID",
			Description: "Stored server ID of the Artifactory configuration to delete from the system keyring.",
		},
	}
}

func deleteCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Expecting only one argument, which is the server ID of the Artifactory configuration to delete from the system keyring")
	}
	serverId := c.Arguments[0]
	if !coreutils.AskYesNo("Are you sure you want to delete the server with ID "+serverId+" from the system keering?", false) {
		return nil
	}
	return doDelete(serverId)
}

func doDelete(serverId string) error {
	err := keyring.Delete(ServiceId, serverId)
	if err == nil {
		log.Info("Deleted Artifactory configuration with ID", serverId, "from the system keering")
	}
	return err
}

