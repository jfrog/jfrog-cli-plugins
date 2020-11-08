package commands

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var modulesDiffHeader = table.Row{"Module", "Art/Dep", "Name/ID", "Diff Name/Id", "Type", "Sha1", "Md5", "Change"}

// Prints a table showing the the builds modules diff.
func printModulesDiffTable(diff *BuildDiff) {
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
		case New.String():
			return text.Colors{text.FgGreen}
		case Unchanged.String():
			return text.Colors{}
		case Updated.String():
			return text.Colors{text.FgBlue}
		case Removed.String():
			return text.Colors{text.FgRed}
		}
		return nil
	})
	renderWithDefaults(t)
}

func fillModulesDiffTable(t table.Writer, diff *BuildDiff) {
	t.AppendHeader(modulesDiffHeader)
	addArtifactsChanges(t, diff.Artifacts)
	addDependenciesChanges(t, diff.Dependencies)
}

func addArtifactsChanges(t table.Writer, artifacts ArtifactsChanges) {
	addArtifactDiffRowsByChange(t, artifacts.New, New)
	addArtifactDiffRowsByChange(t, artifacts.Unchanged, Unchanged)
	addArtifactDiffRowsByChange(t, artifacts.Updated, Updated)
	addArtifactDiffRowsByChange(t, artifacts.Removed, Removed)
}

func addArtifactDiffRowsByChange(t table.Writer, artifacts []ArtifactDiff, change Change) {
	switch change {
	case Removed:
		for _, a := range artifacts {
			addRemovedFileRow(t, a)
		}
	default:
		for _, a := range artifacts {
			addFileRow(t, a, change)
		}
	}
}

func addDependenciesChanges(t table.Writer, dependencies DependenciesChanges) {
	addDependencyDiffRowsByChange(t, dependencies.New, New)
	addDependencyDiffRowsByChange(t, dependencies.Unchanged, Unchanged)
	addDependencyDiffRowsByChange(t, dependencies.Updated, Updated)
	addDependencyDiffRowsByChange(t, dependencies.Removed, Removed)
}

func addDependencyDiffRowsByChange(t table.Writer, dependencies []DependencyDiff, change Change) {
	switch change {
	case Removed:
		for _, d := range dependencies {
			addRemovedFileRow(t, d)
		}
	default:
		for _, d := range dependencies {
			addFileRow(t, d, change)
		}
	}
}

func addFileRow(t table.Writer, f FileDiff, change Change) {
	t.AppendRow(table.Row{f.GetModuleId(), f.GetArtOrDep(), f.GetIdOrName(), f.GetDiffIdOrName(), f.GetType(), f.GetSha1(), f.GetMd5(), change.String()})
}

func addRemovedFileRow(t table.Writer, f FileDiff) {
	t.AppendRow(table.Row{f.GetModuleId(), f.GetArtOrDep(), f.GetIdOrName(), f.GetDiffIdOrName(), "", "", "", Removed.String()})
}
