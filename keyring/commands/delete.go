package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/zalando/go-keyring"
)

func GetDeleteCommand() components.Command {
	return components.Command{
		Name:        "delete",
		Description: "Delete Artifactory configuration from the OS keyring.",
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
			Description: "The server ID stored in the OS keyring. The configuration for this server ID will be deleted from the OS keyring.",
		},
	}
}

func deleteCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Expecting only one argument, which is the server ID of the Artifactory configuration to delete from the OS keyring")
	}
	serverId := c.Arguments[0]
	if !coreutils.AskYesNo("Are you sure you want to delete the server with ID "+serverId+" from the OS keering?", false) {
		return nil
	}
	return doDelete(serverId)
}

func doDelete(serverId string) error {
	err := keyring.Delete(ServiceId, serverId)
	if err == nil {
		log.Info("Deleted Artifactory configuration with ID", serverId, "from the OS keering")
	}
	return err
}
