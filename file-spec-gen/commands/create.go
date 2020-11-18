package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	corecommandutils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"strings"
	"sync"
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

type specslicetype map[string]interface{}

var once sync.Once
var fileSpecsCommand string

func GetFileSpecGenCommand() components.Command {
	return components.Command{
		Name:        "create",
		Description: "Generates a file-spec json.",
		Aliases:     []string{"cr"},
		Flags:       getCreateFlags(),
		Action: func(c *components.Context) error {
			return Run(c)
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

func Run(c *components.Context) error {
	if len(c.Arguments) != 0 {
		return errors.New("the command expects no arguments")
	}

	// Check if should output to file.
	file := c.GetStringFlagValue("file")
	if file != "" {
		// Validate output file path.
		err := validateSpecPath(file)
		if err != nil {
			return err
		}
	}

	// Run the command.
	output, err := runCreate()
	if err != nil {
		return err
	}

	// Handle result.
	return handleResult(output, file)
}

func handleResult(output []byte, file string) error {
	outputJson := clientutils.IndentJson(output)
	if file == "" {
		log.Output(outputJson)
		return nil
	}

	if err := ioutil.WriteFile(file, []byte(outputJson), 0644); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("file-spec successfully created at %s", file))
	return nil
}

func validateSpecPath(templatePath string) error {
	exists, err := fileutils.IsDirExists(templatePath, false)
	if err != nil {
		return err
	}
	if exists || strings.HasSuffix(templatePath, string(os.PathSeparator)) {
		return errors.New("path cannot be a directory, please enter a path in which the new file-spec file will be created")
	}
	exists, err = fileutils.IsFileExists(templatePath, false)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("file already exists, please enter a path in which the new file-spec will be created")
	}
	return nil
}

func runCreate() ([]byte, error) {
	// Perform the questionnaire.
	resultSlice, err := doQuestionnaire()
	if err != nil {
		return nil, err
	}
	// Create the file spec json.
	return buildFileSpecJson(resultSlice)
}

func doQuestionnaire() ([]specslicetype, error) {
	var resultSlice []specslicetype

	// Ask for first spec.
	fileSpecQuestionnaire := &corecommandutils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{SpecCommand},
		QuestionsMap:           questionMap,
	}
	err := fileSpecQuestionnaire.Perform()
	if err != nil {
		return nil, err
	}

	// Add more specs if required.
	addAnotherSpec := askForAnotherFileSpec()
	for addAnotherSpec {
		// Get answerMap and append to result with separating ',' at the end.
		resultSlice = append(resultSlice, fileSpecQuestionnaire.AnswersMap)

		// Create a new questionnaire, use the same command as previous file-spec.
		// Meaning - if previous file-spec is used for search, each additional file-spec should be
		// for searching as well.
		fileSpecQuestionnaire = &corecommandutils.InteractiveQuestionnaire{
			MandatoryQuestionsKeys: []string{},
			QuestionsMap:           questionMap,
		}
		_, err := specCommandCallback(fileSpecQuestionnaire, fileSpecsCommand)
		if err != nil {
			return nil, err
		}

		// Perform.
		err = fileSpecQuestionnaire.Perform()
		if err != nil {
			return nil, err
		}

		// Check if should add another file-spec.
		addAnotherSpec = askForAnotherFileSpec()
	}

	// Append map to result.
	return append(resultSlice, fileSpecQuestionnaire.AnswersMap), nil
}

func askForAnotherFileSpec() bool {
	return coreutils.AskYesNo("Do you want to add another file-spec?", false)
}

func buildFileSpecJson(resultSlice []specslicetype) ([]byte, error) {
	fileSpecPattern := "{\"files\": %s}"

	sliceJson, err := json.Marshal(resultSlice)
	if err != nil {
		return nil, err
	}

	sliceMapStr := string(sliceJson)
	finalStr := fmt.Sprintf(fileSpecPattern, sliceMapStr)

	return []byte(finalStr), nil
}

// Add required questions based on the spec purpose.
func specCommandCallback(iq *corecommandutils.InteractiveQuestionnaire, specCommand string) (string, error) {
	// Each SpecType has its own mandatory and optional configuration keys.
	// We set the questionnaire's keys according to the selected value.
	switch specCommand {
	case Search, Download, Delete, SetProps:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, SpecType)
	case Upload:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, Pattern)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getUploadOptionalConf()...)
	case Move, Copy:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, SpecType, Target)
	default:
		return "", errors.New(fmt.Sprintf("unsupported %s was configured", SpecCommand))
	}

	// Update fileSpecsCommand only on the first questionnaire.
	once.Do(func() {
		fileSpecsCommand = specCommand
	})
	return "", nil
}

// Add required questions for Pattern/ Aql file-spec according to file-spec command.
func specTypeCallback(iq *corecommandutils.InteractiveQuestionnaire, specType string) (string, error) {
	// Each SpecType has its own optional configuration keys.
	// We set the questionnaire's optionalKeys suggests according to the selected value.
	if _, ok := iq.AnswersMap[SpecCommand]; !ok {
		if fileSpecsCommand == "" {
			// If fileSpecsCommand is empty, this is the first questionnaire run and the AnswerMap must contain
			// a spec-command.
			return "", errors.New(fmt.Sprintf("%s is missing in configuration map", SpecCommand))
		}
		// This is not the first run, populate the AnswerMap for further execution.
		iq.AnswersMap[SpecCommand] = fileSpecsCommand
	}

	switch specType {
	case Pattern:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getPatternMandatoryConf(iq.AnswersMap[SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getPatternOptionalConf(iq.AnswersMap[SpecCommand].(string))...)
	case Aql:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getAqlMandatoryConf(iq.AnswersMap[SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getAqlOptionalConf(iq.AnswersMap[SpecCommand].(string))...)
	case Build:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getBuildMandatoryConf(iq.AnswersMap[SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getBuildBundleOptionalConf(iq.AnswersMap[SpecCommand].(string))...)
	case Bundle:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getBundleMandatoryConf(iq.AnswersMap[SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getBuildBundleOptionalConf(iq.AnswersMap[SpecCommand].(string))...)
	default:
		return "", errors.New(fmt.Sprintf("unsupported %s was configured", SpecType))
	}

	// Clean specType value from final configuration
	delete(iq.AnswersMap, SpecType)
	// Clean specCommand value from final configuration
	delete(iq.AnswersMap, SpecCommand)

	return "", nil
}

func optionalKeyCallback(iq *corecommandutils.InteractiveQuestionnaire, key string) (value string, err error) {
	if key != corecommandutils.SaveAndExit {
		valueQuestion := iq.QuestionsMap[key]
		valueQuestion.MapKey = key
		// If prompt-prefix wasn't provided, use a default prompt.
		if valueQuestion.PromptPrefix == "" {
			valueQuestion.PromptPrefix = corecommandutils.InsertValuePromptMsg + key
			if valueQuestion.Options != nil {
				valueQuestion.PromptPrefix += corecommandutils.PressTabMsg
			}
			valueQuestion.PromptPrefix += " >"
		}
		value, err = iq.AskQuestion(valueQuestion)
	}
	return value, err
}

func getPatternMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case Search, Download, Move, Copy:
		mandatoryKeys = append(mandatoryKeys, Pattern)
	}
	return mandatoryKeys
}

func getPatternOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case Search, Delete, SetProps:
		optionalKeys = append(optionalKeys, SearchPatternOptionalKeys...)
	case Download:
		optionalKeys = append(optionalKeys, DownloadPatternOptionalKeys...)
	case Move, Copy:
		optionalKeys = append(optionalKeys, MoveCopyPatternOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, OptionalSuggestsMap)
}

func getAqlMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case Search, Download, Move, Copy, Delete, SetProps:
		mandatoryKeys = append(mandatoryKeys, Aql)
	}
	return mandatoryKeys
}

func getAqlOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case Search, Delete, SetProps:
		optionalKeys = append(optionalKeys, SearchAqlOptionalKeys...)
	case Download:
		optionalKeys = append(optionalKeys, DownloadAqlOptionalKeys...)
	case Move, Copy:
		optionalKeys = append(optionalKeys, MoveCopyAqlOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, OptionalSuggestsMap)
}

func getBuildMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case Search, Download, Move, Copy, Delete, SetProps:
		mandatoryKeys = append(mandatoryKeys, Build)
	}
	return mandatoryKeys
}

func getBundleMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case Search, Download, Move, Copy, Delete, SetProps:
		mandatoryKeys = append(mandatoryKeys, Bundle)
	}
	return mandatoryKeys
}

func getBuildBundleOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case Search, Delete, SetProps:
		optionalKeys = append(optionalKeys, SearchBuildBundleOptionalKeys...)
	case Download:
		optionalKeys = append(optionalKeys, DownloadBuildBundleOptionalKeys...)
	case Move, Copy:
		optionalKeys = append(optionalKeys, MoveCopyBuildBundleOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, OptionalSuggestsMap)
}

func getUploadOptionalConf() []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	optionalKeys = append(optionalKeys, UploadOptionalKeys...)
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, OptionalSuggestsMap)
}

var questionMap = map[string]corecommandutils.QuestionInfo{
	corecommandutils.OptionalKey: {
		PromptPrefix: "Select the next property >",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     optionalKeyCallback,
	},
	SpecCommand: {
		Options: []prompt.Suggest{
			{Text: Search, Description: "Search file-spec"},
			{Text: Download, Description: "Download file-spec"},
			{Text: Upload, Description: "Upload file-spec"},
			{Text: Move, Description: "Move file-spec"},
			{Text: Copy, Description: "Copy file-spec"},
			{Text: Delete, Description: "Delete file-spec"},
			{Text: SetProps, Description: "Set-props file-spec"},
		},
		PromptPrefix: "Select file-spec purpose" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       SpecCommand,
		Callback:     specCommandCallback,
	},
	SpecType: {
		Options: []prompt.Suggest{
			{Text: Pattern, Description: "File-spec with pattern"},
			{Text: Aql, Description: "File-spec with AQL"},
			{Text: Build, Description: "Build based file-spec"},
			{Text: Bundle, Description: "Bundle based file-spec"},
		},
		PromptPrefix: "Select the file-spec type" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       SpecType,
		Callback:     specTypeCallback,
	},
	Pattern: {
		PromptPrefix: "Insert the pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Pattern,
		Callback:     nil,
	},
	Aql: {
		PromptPrefix: "Insert the aql >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Aql,
		Callback:     nil,
	},
	Target: {
		PromptPrefix: "Insert the target >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Target,
		Callback:     nil,
	},
	Props: {
		PromptPrefix: "Enter \"key=value\" pairs separated by a semi-colon (key1=value1;key2=value2) >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Props,
		Callback:     nil,
	},
	ExcludeProps: {
		PromptPrefix: "Enter \"key=value\" pairs separated by a semi-colon (key1=value1;key2=value2) >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       ExcludeProps,
		Callback:     nil,
	},
	Recursive: {
		Options:      corecommandutils.GetBoolSuggests(),
		PromptPrefix: "Select if recursive" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Recursive,
		Callback:     nil,
	},
	Exclusions: {
		PromptPrefix: "Enter a comma separated list of exclusion patterns >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringArrayAnswer,
		MapKey:       Exclusions,
		Callback:     nil,
	},
	ArchiveEntries: {
		PromptPrefix: "Insert archive-entries pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       ArchiveEntries,
		Callback:     nil,
	},
	Build: {
		PromptPrefix: "Insert build pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Build,
		Callback:     nil,
	},
	Bundle: {
		Msg:          "",
		PromptPrefix: "Insert bundle >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       Bundle,
		Callback:     nil,
	},
	SortBy: {
		PromptPrefix: "Enter a comma separated list of sort-by values >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringArrayAnswer,
		MapKey:       SortBy,
		Callback:     nil,
	},
	SortOrder: {
		Options: []prompt.Suggest{
			{Text: Asc, Description: ""},
			{Text: Desc, Description: ""},
		},
		AllowVars: false,
		Writer:    corecommandutils.WriteStringAnswer,
		MapKey:    SortOrder,
		Callback:  nil,
	},
	Limit: {
		AllowVars: false,
		Writer:    corecommandutils.WriteStringAnswer,
		MapKey:    Limit,
		Callback:  nil,
	},
	Offset: {
		AllowVars: false,
		Writer:    corecommandutils.WriteStringAnswer,
		MapKey:    Offset,
		Callback:  nil,
	},
	Flat: {
		Options:   corecommandutils.GetBoolSuggests(),
		AllowVars: false,
		Writer:    corecommandutils.WriteStringAnswer,
		MapKey:    Flat,
		Callback:  nil,
	},
	ValidateSymlinks: {
		Options:      corecommandutils.GetBoolSuggests(),
		PromptPrefix: "Select if should validate symlinks" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       ValidateSymlinks,
		Callback:     nil,
	},
	Regexp: {
		Options:   corecommandutils.GetBoolSuggests(),
		AllowVars: false,
		Writer:    corecommandutils.WriteStringAnswer,
		MapKey:    Regexp,
		Callback:  nil,
	},
}

var OptionalSuggestsMap = map[string]prompt.Suggest{
	corecommandutils.SaveAndExit: {Text: corecommandutils.SaveAndExit},
	Props:                        {Text: Props},
	ExcludeProps:                 {Text: ExcludeProps},
	Target:                       {Text: Target},
	Recursive:                    {Text: Recursive},
	Exclusions:                   {Text: Exclusions},
	ArchiveEntries:               {Text: ArchiveEntries},
	Build:                        {Text: Build},
	Bundle:                       {Text: Bundle},
	SortBy:                       {Text: SortBy},
	SortOrder:                    {Text: SortOrder},
	Limit:                        {Text: Limit},
	Offset:                       {Text: Offset},
	Flat:                         {Text: Flat},
	ValidateSymlinks:             {Text: ValidateSymlinks},
	Regexp:                       {Text: Regexp},
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
