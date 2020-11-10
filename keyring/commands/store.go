package commands

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/zalando/go-keyring"
)

const ServiceId = "keyring-jfrog-cli-plugin"

func GetStoreCommand() components.Command {
	return components.Command{
		Name:        "store",
		Description: "Store Artifactory configuration in the OS keyring.",
		Aliases:     []string{"s"},
		Flags:       getStoreFlags(),
		Action: func(c *components.Context) error {
			return storeCmd(c)
		},
	}
}

func getStoreFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID used for accessing the configuration.",
			Mandatory:   true,
		},
		components.StringFlag{
			Name:        "url",
			Description: "Artifactory URL.",
			Mandatory:   true,
		},
		components.StringFlag{
			Name:        "user",
			Description: "Artifactory username.",
			Mandatory:   true,
		},
		components.StringFlag{
			Name:        "password",
			Description: "Artifactory password.",
			Mandatory:   true,
		},
	}
}

type storeConfiguration struct {
	ServerId string
	Url      string
	User     string
	Password string
}

func storeCmd(c *components.Context) error {
	if len(c.Arguments) != 0 {
		return errors.New("This command expects no arguments")
	}
	var conf = storeConfiguration{
		ServerId: c.GetStringFlagValue("server-id"),
		Url:      c.GetStringFlagValue("url"),
		User:     c.GetStringFlagValue("user"),
		Password: c.GetStringFlagValue("password"),
	}
	return doStore(conf)
}

func doStore(conf storeConfiguration) error {
	secret, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	err = keyring.Set(ServiceId, conf.ServerId, string(secret))
	if err == nil {
		log.Info("Stored Artifactory configuration in the OS keering")
	}

	return err
}
