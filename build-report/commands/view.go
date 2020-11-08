package commands

import (
	"errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-plugins/build-report/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"strconv"
)

func GetViewCommand() components.Command {
	return components.Command{
		Name:        "view",
		Description: "Print build report of requested build",
		Aliases:     []string{"v"},
		Arguments:   getViewArguments(),
		Flags:       getViewFlags(),
		EnvVars:     getViewEnvVar(),
		Action: func(c *components.Context) error {
			return viewCmd(c)
		},
	}
}

const buildNameArgument = "build-name"
const buildNumberArgument = "build-number"
const diffFlag = "diff"

func getViewArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        buildNameArgument,
			Description: "Name of the build to print report for.",
		},
		{
			Name:        buildNumberArgument,
			Description: "Number of the build to print report for.",
		},
	}
}

func getViewFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        utils.ServerIdFlag,
			Description: "Artifactory server ID configured using the config command.",
		},
		components.StringFlag{
			Name: diffFlag,
			Description: "A build number to show diff with. " +
				"Renders the table to show difference in artifacts, dependencies and properties with the provided build number.",
		},
	}
}

func getViewEnvVar() []components.EnvVar {
	return []components.EnvVar{
		{
			Name:        coreutils.BuildName,
			Description: "Build name to be used by commands which expect a build name, unless sent as a command argument or option.",
		},
		{
			Name:        coreutils.BuildNumber,
			Description: "Build number to be used by commands which expect a build number, unless sent as a command argument or option.",
		},
	}
}

func viewCmd(c *components.Context) error {
	if !(len(c.Arguments) == 2 || len(c.Arguments) == 0) {
		return errors.New("wrong number of arguments. Expected 2 arguments, or 0 with build details passed as environment variables")
	}
	buildName, buildNumber, err := utils.GetBuildDetails(c)
	if err != nil {
		return err
	}

	buildNumberDiff := c.GetStringFlagValue(diffFlag)
	err = verifyOlderDiffBuildNumber(buildNumber, buildNumberDiff)
	if err != nil {
		return err
	}

	rtDetails, err := utils.GetRtDetails(c)
	if err != nil {
		return err
	}

	return doView(rtDetails, buildName, buildNumber, buildNumberDiff)
}

func verifyOlderDiffBuildNumber(buildNumber, buildNumberDiff string) error {
	if buildNumberDiff == "" {
		return nil
	}
	buildInt, err := strconv.Atoi(buildNumber)
	if err != nil {
		return err
	}
	buildDiffInt, err := strconv.Atoi(buildNumberDiff)
	if err != nil {
		return err
	}
	if buildDiffInt >= buildInt {
		return errors.New("build number to show diff with must be older than the report build number")
	}
	return nil
}

func doView(rtDetails *config.ArtifactoryDetails, buildName, buildNumber, buildNumberDiff string) error {
	publishedBuildInfo, err := utils.GetBuildInfo(rtDetails, buildName, buildNumber)
	if err != nil {
		return err
	}
	diff, err := utils.GetBuildDiff(rtDetails, buildName, buildNumber, buildNumberDiff)
	if err != nil {
		return err
	}

	printBuildDetailsTable(publishedBuildInfo)
	if diff != nil {
		printModulesDiffTable(diff)
	} else {
		printBuildModulesTable(publishedBuildInfo.BuildInfo.Modules)
	}
	return nil
}

func printBuildDetailsTable(publishedBuildInfo *buildinfo.PublishedBuildInfo) {
	t := table.NewWriter()
	t.SetTitle("Build Details")
	fillBuildDetailsTable(t, publishedBuildInfo.BuildInfo, publishedBuildInfo.Uri)
	utils.RenderWithDefaults(t)
}

func fillBuildDetailsTable(t table.Writer, buildInfo buildinfo.BuildInfo, buildUri string) {
	// Repeating Agents in the first header will be merged as one cell above their name/version.
	t.AppendHeader(table.Row{"Name", "Number", "Started", "Uri", "Artifactory Principal", "Agent", "Agent", "Build Agent", "Build Agent"}, table.RowConfig{AutoMerge: true})
	t.AppendHeader(table.Row{"", "", "", "", "", "Name", "Version", "Name", "Version"})
	t.AppendRow(table.Row{buildInfo.Name, buildInfo.Number, buildInfo.Started, buildUri, buildInfo.ArtifactoryPrincipal,
		buildInfo.Agent.Name, buildInfo.Agent.Version, buildInfo.BuildAgent.Name, buildInfo.BuildAgent.Version})
}

func printBuildModulesTable(modules []buildinfo.Module) {
	t := table.NewWriter()
	t.SetTitle("Modules")

	fillBuildModulesTable(t, modules)

	// Merges the elements on the "Module" and "Art/Dep" columns
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
	})
	utils.RenderWithDefaults(t)
}

func fillBuildModulesTable(t table.Writer, modules []buildinfo.Module) {
	t.AppendHeader(table.Row{"Module", "Art/Dep", "Name/ID", "Type", "Sha1", "Md5"})
	for _, mod := range modules {
		for _, art := range mod.Artifacts {
			t.AppendRow(table.Row{mod.Id, "Artifact", art.Name, art.Type, art.Sha1, art.Md5})
		}
		for _, dep := range mod.Dependencies {
			t.AppendRow(table.Row{mod.Id, "Dependency", dep.Id, dep.Type, dep.Sha1, dep.Md5})
		}
	}
}

var modulesDiffHeader = table.Row{"Module", "Art/Dep", "Name/ID", "Diff Name/Id", "Type", "Sha1", "Md5", "Change"}

// Prints a table showing the the builds modules diff.
func printModulesDiffTable(diff *utils.BuildDiff) {
	t := table.NewWriter()
	t.SetTitle("Modules")

	fillModulesDiffTable(t, diff)

	// Merges the elements on the "Module" and "Art/Dep" columns
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
	})

	// Sorts the table to show modules and their artifacts/dependencies joint.
	t.SortBy([]table.SortBy{{Name: "Module", Mode: table.Asc},
		{Name: "Art/Dep", Mode: table.Asc}})

	// Colors each line according to the change of the file.
	t.SetRowPainter(func(row table.Row) text.Colors {
		switch row[len(modulesDiffHeader)-1] {
		case utils.New.String():
			return text.Colors{text.FgGreen}
		case utils.Unchanged.String():
			return text.Colors{}
		case utils.Updated.String():
			return text.Colors{text.FgBlue}
		case utils.Removed.String():
			return text.Colors{text.FgRed}
		}
		return nil
	})
	utils.RenderWithDefaults(t)
}

func fillModulesDiffTable(t table.Writer, diff *utils.BuildDiff) {
	t.AppendHeader(modulesDiffHeader)
	addArtifactsChanges(t, diff.Artifacts)
	addDependenciesChanges(t, diff.Dependencies)
}

func addArtifactsChanges(t table.Writer, artifacts utils.ArtifactsChanges) {
	addArtifactDiffRowsByChange(t, artifacts.New, utils.New)
	addArtifactDiffRowsByChange(t, artifacts.Unchanged, utils.Unchanged)
	addArtifactDiffRowsByChange(t, artifacts.Updated, utils.Updated)
	addArtifactDiffRowsByChange(t, artifacts.Removed, utils.Removed)
}

func addArtifactDiffRowsByChange(t table.Writer, artifacts []utils.ArtifactDiff, change utils.Change) {
	switch change {
	case utils.Removed:
		for _, a := range artifacts {
			addRemovedFileRow(t, a)
		}
	default:
		for _, a := range artifacts {
			addFileRow(t, a, change)
		}
	}
}

func addDependenciesChanges(t table.Writer, dependencies utils.DependenciesChanges) {
	addDependencyDiffRowsByChange(t, dependencies.New, utils.New)
	addDependencyDiffRowsByChange(t, dependencies.Unchanged, utils.Unchanged)
	addDependencyDiffRowsByChange(t, dependencies.Updated, utils.Updated)
	addDependencyDiffRowsByChange(t, dependencies.Removed, utils.Removed)
}

func addDependencyDiffRowsByChange(t table.Writer, dependencies []utils.DependencyDiff, change utils.Change) {
	switch change {
	case utils.Removed:
		for _, d := range dependencies {
			addRemovedFileRow(t, d)
		}
	default:
		for _, d := range dependencies {
			addFileRow(t, d, change)
		}
	}
}

func addFileRow(t table.Writer, f utils.FileDiff, change utils.Change) {
	t.AppendRow(table.Row{f.GetModuleId(), f.GetArtOrDep(), f.GetIdOrName(), f.GetDiffIdOrName(), f.GetType(), f.GetSha1(), f.GetMd5(), change.String()})
}

func addRemovedFileRow(t table.Writer, f utils.FileDiff) {
	t.AppendRow(table.Row{f.GetModuleId(), f.GetArtOrDep(), f.GetIdOrName(), f.GetDiffIdOrName(), "", "", "", utils.Removed.String()})
}
