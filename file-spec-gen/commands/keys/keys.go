package keys

import (
	"github.com/c-bata/go-prompt"
	coreutils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
)

const (
	// File-spec commands.
	SpecCommand = "specCommand"
	Search      = "search"
	Download    = "download"
	Upload      = "upload"
	Move        = "move"
	Copy        = "copy"
	Delete      = "delete"
	SetProps    = "setProps"

	// General keys.
	Pattern          = "pattern"
	Aql              = "aql"
	SpecType         = "specType"
	Target           = "target"
	Props            = "props"
	ExcludeProps     = "excludeProps"
	Recursive        = "recursive"
	Exclusions       = "exclusions"
	ArchiveEntries   = "archiveEntries"
	Build            = "build"
	Bundle           = "bundle"
	SortBy           = "sortBy"
	SortOrder        = "sortOrder"
	Asc              = "asc"
	Desc             = "desc"
	Limit            = "limit"
	Offset           = "offset"
	Flat             = "flat"
	ValidateSymlinks = "validateSymlinks"
	Regexp           = "regexp"
)

var OptionalSuggestsMap = map[string]prompt.Suggest{
	coreutils.SaveAndExit: {Text: coreutils.SaveAndExit},
	Props:                 {Text: Props},
	ExcludeProps:          {Text: ExcludeProps},
	Target:                {Text: Target},
	Recursive:             {Text: Recursive},
	Exclusions:            {Text: Exclusions},
	ArchiveEntries:        {Text: ArchiveEntries},
	Build:                 {Text: Build},
	Bundle:                {Text: Bundle},
	SortBy:                {Text: SortBy},
	SortOrder:             {Text: SortOrder},
	Limit:                 {Text: Limit},
	Offset:                {Text: Offset},
	Flat:                  {Text: Flat},
	ValidateSymlinks:      {Text: ValidateSymlinks},
	Regexp:                {Text: Regexp},
}

var SearchPatternOptionalKeys = []string{
	Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder, Limit, Offset,
}

var SearchAqlOptionalKeys = []string{
	Props, ExcludeProps, Recursive, ArchiveEntries, SortBy, SortOrder, Limit, Offset,
}

var SearchBuildBundleOptionalKeys = []string{
	Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder,
}

var DownloadPatternOptionalKeys = []string{
	Target, Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder, Limit, Offset, Flat,
}

var DownloadBuildBundleOptionalKeys = []string{
	Target, Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder, Flat,
}

var DownloadAqlOptionalKeys = []string{
	Target, Props, ExcludeProps, Recursive, ArchiveEntries, SortBy, SortOrder, Limit, Offset, Flat,
}

var UploadOptionalKeys = []string{
	Target, Props, Recursive, Exclusions, Flat,
}

var MoveCopyPatternOptionalKeys = []string{
	Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder, Limit, Offset, Flat, ValidateSymlinks,
}

var MoveCopyAqlOptionalKeys = []string{
	Props, ExcludeProps, Recursive, ArchiveEntries, SortBy, SortOrder, Limit, Offset, Flat, ValidateSymlinks,
}

var MoveCopyBuildBundleOptionalKeys = []string{
	Props, ExcludeProps, Recursive, Exclusions, ArchiveEntries, SortBy, SortOrder, Flat, ValidateSymlinks,
}
