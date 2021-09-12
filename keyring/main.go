package main

import (
	"github.com/jfrog/jfrog-cli-core/v2/plugins"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-plugins/keyring/commands"
)

func main() {
	plugins.PluginMain(getApp())
}

func getApp() components.App {
	app := components.App{}
	app.Name = "keyring"
	app.Description = "Store Artifactory configuration in the OS keyring and use them when running JFrog CLI commands."
	app.Version = "v1.0.1"
	app.Commands = getCommands()
	return app
}

func getCommands() []components.Command {
	return []components.Command{
		commands.GetStoreCommand(),
		commands.GetUseCommand(),
		commands.GetDeleteCommand(),
	}
}
