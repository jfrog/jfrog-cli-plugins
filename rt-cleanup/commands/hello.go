package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"strconv"
	"strings"
)

func GetHelloCommand() components.Command {
	return components.Command{
		Name:        "hello",
		Description: "Says Hello.",
		Aliases:     []string{"hi"},
		Arguments:   getHelloArguments(),
		Flags:       getHelloFlags(),
		EnvVars:     getHelloEnvVar(),
		Action: func(c *components.Context) error {
			return helloCmd(c)
		},
	}
}

func getHelloArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "addressee",
			Description: "The name of the person you would like to greet.",
		},
	}
}

func getHelloFlags() []components.Flag {
	return []components.Flag{
		components.BoolFlag{
			Name:         "shout",
			Description:  "Makes output uppercase.",
			DefaultValue: false,
		},
		components.StringFlag{
			Name:         "repeat",
			Description:  "Greets multiple times.",
			DefaultValue: "1",
		},
	}
}

func getHelloEnvVar() []components.EnvVar {
	return []components.EnvVar{
		{
			Name:        "HELLO_FROG_GREET_PREFIX",
			Default:     "A new greet from your plugin template: ",
			Description: "Adds a prefix to every greet.",
		},
	}
}

type helloConfiguration struct {
	addressee string
	shout     bool
	repeat    int
	prefix    string
}

func helloCmd(c *components.Context) error {
	if len(c.Arguments) != 1 {
		return errors.New("Wrong number of arguments. Expected: 1, " + "Received: " + strconv.Itoa(len(c.Arguments)))
	}
	var conf = new(helloConfiguration)
	conf.addressee = c.Arguments[0]
	conf.shout = c.GetBoolFlagValue("shout")

	repeat, err := strconv.Atoi(c.GetStringFlagValue("repeat"))
	if err != nil {
		return err
	}
	conf.repeat = repeat

	conf.prefix = os.Getenv("HELLO_FROG_GREET_PREFIX")
	if conf.prefix == "" {
		conf.prefix = "New greeting: "
	}

	log.Output(doGreet(conf))
	return nil
}

func doGreet(c *helloConfiguration) string {
	greet := c.prefix + "Hello " + c.addressee + "!\n"

	if c.shout {
		greet = strings.ToUpper(greet)
	}

	return strings.TrimSpace(strings.Repeat(greet, c.repeat))
}
