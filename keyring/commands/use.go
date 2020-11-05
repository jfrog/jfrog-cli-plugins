package commands

import (
	"encoding/json"
	"errors"
	gofrogio "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/zalando/go-keyring"
	"io"
	"os/exec"
	"strings"
)

func GetUseCommand() components.Command {
	return components.Command{
		Name:        "use",
		Description: "Use the store Artifactory configuration for a JFrog CLI command.",
		Aliases:     []string{"u"},
		Arguments:   getUseArguments(),
		SkipFlagParsing: true,
		Action: func(c *components.Context) error {
			return useCmd(c)
		},
	}
}

func getUseArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "server ID",
			Description: "The stored server ID of the Artifactory configuration to use for this command.",
		},
	}
}

func useCmd(c *components.Context) error {
	if len(c.Arguments) == 0 {
		return errors.New("Wrong number of arguments. Expecting the server ID of the Artifactory configuration to be used for this command")
	}

	secret, err := keyring.Get(ServiceId, c.Arguments[0])
	if err != nil {
		return err
	}

	var conf = new(storeConfiguration)
	err = json.Unmarshal([]byte(secret), conf)
	if err != nil {
		return err
	}

	args, err := buildArgs(c.Arguments, conf)
	if err != nil {
		return err
	}

	cmd := new(Cmd)
	cmd.Args = args
	return gofrogio.RunCmd(cmd)
}

func buildArgs(args []string, conf *storeConfiguration) ([]string, error) {
	// Remove the first argument, which is the server ID.
	args = args[1:]
	// Add "rt" at the beginning.
	args = append([]string{"rt"}, args...)

	// Add the connection details flags.
	index := findFirstFlagIndex(args)
	leftSection:= append(args[0:index], []string{"--url", conf.Url, "--user", conf.User, "--password", conf.Password}...)
	args = append(leftSection, args[index:]...)

	return utils.ParseArgs(args)
}

// This function receives an array of command args, and searches for the first argument, which is
// actually a flag (that its name starts with --).
// If there are no flags, the last argument index plus 1 is returned.
func findFirstFlagIndex(args []string) int {
	var i int
	var arg string
	for i, arg = range args {
		if strings.HasPrefix(arg, "--") {
			return i
		}
	}
	return i+1
}

type Cmd struct {
	Args         []string
	StrWriter    io.WriteCloser
	ErrWriter    io.WriteCloser
}

func (c *Cmd) GetCmd() *exec.Cmd {
	return exec.Command("jfrog", c.Args...)
}

func (c *Cmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (c *Cmd) GetStdWriter() io.WriteCloser {
	return c.StrWriter
}

func (c *Cmd) GetErrWriter() io.WriteCloser {
	return c.ErrWriter
}

