package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	corecommandutils "github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-plugins/file-spec-gen/commands/keys"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type specslicetype map[string]interface{}

var once sync.Once
var fileSpecsCommand string

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
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("file-spec successfully created at %s", file))
	return nil
}

func validateSpecPath(templatePath string) error {
	exists, err := fileutils.IsDirExists(templatePath, false)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if exists || strings.HasSuffix(templatePath, string(os.PathSeparator)) {
		return errorutils.CheckError(errors.New("path cannot be a directory, please enter a path in which the new file-spec file will be created"))
	}
	exists, err = fileutils.IsFileExists(templatePath, false)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if exists {
		return errorutils.CheckError(errors.New("file already exists, please enter a path in which the new file-spec will be created"))
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
		MandatoryQuestionsKeys: []string{keys.SpecCommand},
		QuestionsMap:           questionMap,
	}
	err := fileSpecQuestionnaire.Perform()
	if err != nil {
		return nil, err
	}

	// Add more specs if required.
	addAnotherSpec := coreutils.AskYesNo("Do you want to add another file-spec?", false)
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
		addAnotherSpec = coreutils.AskYesNo("Do you want to add another file-spec?", false)
	}

	// Append map to result.
	return append(resultSlice, fileSpecQuestionnaire.AnswersMap), nil
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
	case keys.Search, keys.Download, keys.Delete, keys.SetProps:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, keys.SpecType)
	case keys.Upload:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, keys.Pattern)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getUploadOptionalConf()...)
	case keys.Move, keys.Copy:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, keys.SpecType, keys.Target)
	default:
		return "", errors.New(fmt.Sprintf("unsupported %s was configured", keys.SpecCommand))
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
	if _, ok := iq.AnswersMap[keys.SpecCommand]; !ok {
		if fileSpecsCommand == "" {
			// If fileSpecsCommand is empty, this is the first questionnaire run and the AnswerMap must contain
			// a spec-command.
			return "", errors.New(fmt.Sprintf("%s is missing in configuration map", keys.SpecCommand))
		}
		// This is not the first run, populate the AnswerMap for further execution.
		iq.AnswersMap[keys.SpecCommand] = fileSpecsCommand
	}

	switch specType {
	case keys.Pattern:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getPatternMandatoryConf(iq.AnswersMap[keys.SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getPatternOptionalConf(iq.AnswersMap[keys.SpecCommand].(string))...)
	case keys.Aql:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getAqlMandatoryConf(iq.AnswersMap[keys.SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getAqlOptionalConf(iq.AnswersMap[keys.SpecCommand].(string))...)
	case keys.Build:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getBuildMandatoryConf(iq.AnswersMap[keys.SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getBuildBundleOptionalConf(iq.AnswersMap[keys.SpecCommand].(string))...)
	case keys.Bundle:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, getBundleMandatoryConf(iq.AnswersMap[keys.SpecCommand].(string))...)
		iq.OptionalKeysSuggests = append(iq.OptionalKeysSuggests, getBuildBundleOptionalConf(iq.AnswersMap[keys.SpecCommand].(string))...)
	default:
		return "", errors.New(fmt.Sprintf("unsupported %s was configured", keys.SpecType))
	}

	// Clean specType value from final configuration
	delete(iq.AnswersMap, keys.SpecType)
	// Clean specCommand value from final configuration
	delete(iq.AnswersMap, keys.SpecCommand)

	return "", nil
}

func getPatternMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case keys.Search, keys.Download, keys.Move, keys.Copy:
		mandatoryKeys = append(mandatoryKeys, keys.Pattern)
	}
	return mandatoryKeys
}

func getPatternOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case keys.Search, keys.Delete, keys.SetProps:
		optionalKeys = append(optionalKeys, keys.SearchPatternOptionalKeys...)
	case keys.Download:
		optionalKeys = append(optionalKeys, keys.DownloadPatternOptionalKeys...)
	case keys.Move, keys.Copy:
		optionalKeys = append(optionalKeys, keys.MoveCopyPatternOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, keys.OptionalSuggestsMap)
}

func getAqlMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case keys.Search, keys.Download, keys.Move, keys.Copy, keys.Delete, keys.SetProps:
		mandatoryKeys = append(mandatoryKeys, keys.Aql)
	}
	return mandatoryKeys
}

func getAqlOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case keys.Search, keys.Delete, keys.SetProps:
		optionalKeys = append(optionalKeys, keys.SearchPatternOptionalKeys...)
	case keys.Download:
		optionalKeys = append(optionalKeys, keys.DownloadPatternOptionalKeys...)
	case keys.Move, keys.Copy:
		optionalKeys = append(optionalKeys, keys.MoveCopyAqlOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, keys.OptionalSuggestsMap)
}

func getBuildMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case keys.Search, keys.Download, keys.Move, keys.Copy, keys.Delete, keys.SetProps:
		mandatoryKeys = append(mandatoryKeys, keys.Build)
	}
	return mandatoryKeys
}

func getBundleMandatoryConf(commandType string) []string {
	var mandatoryKeys []string
	switch commandType {
	case keys.Search, keys.Download, keys.Move, keys.Copy, keys.Delete, keys.SetProps:
		mandatoryKeys = append(mandatoryKeys, keys.Bundle)
	}
	return mandatoryKeys
}

func getBuildBundleOptionalConf(commandType string) []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	switch commandType {
	case keys.Search, keys.Delete, keys.SetProps:
		optionalKeys = append(optionalKeys, keys.SearchBuildBundleOptionalKeys...)
	case keys.Download:
		optionalKeys = append(optionalKeys, keys.DownloadBuildBundleOptionalKeys...)
	case keys.Move, keys.Copy:
		optionalKeys = append(optionalKeys, keys.MoveCopyBuildBundleOptionalKeys...)
	}
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, keys.OptionalSuggestsMap)
}

func getUploadOptionalConf() []prompt.Suggest {
	optionalKeys := []string{corecommandutils.SaveAndExit}
	optionalKeys = append(optionalKeys, keys.UploadOptionalKeys...)
	return corecommandutils.GetSuggestsFromKeys(optionalKeys, keys.OptionalSuggestsMap)
}

var questionMap = map[string]corecommandutils.QuestionInfo{
	corecommandutils.OptionalKey: {
		Msg:          "",
		PromptPrefix: "Select the next property >",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     corecommandutils.OptionalKeyCallback,
	},
	keys.SpecCommand: {
		Options: []prompt.Suggest{
			{Text: keys.Search, Description: "Search file-spec"},
			{Text: keys.Download, Description: "Download file-spec"},
			{Text: keys.Upload, Description: "Upload file-spec"},
			{Text: keys.Move, Description: "Move file-spec"},
			{Text: keys.Copy, Description: "Copy file-spec"},
			{Text: keys.Delete, Description: "Delete file-spec"},
			{Text: keys.SetProps, Description: "Set-props file-spec"},
		},
		Msg:          "",
		PromptPrefix: "Select file-spec purpose" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.SpecCommand,
		Callback:     specCommandCallback,
	},
	keys.SpecType: {
		Options: []prompt.Suggest{
			{Text: keys.Pattern, Description: "File-spec with pattern"},
			{Text: keys.Aql, Description: "File-spec with AQL"},
			{Text: keys.Build, Description: "Build based file-spec"},
			{Text: keys.Bundle, Description: "Bundle based file-spec"},
		},
		Msg:          "",
		PromptPrefix: "Select the file-spec type" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.SpecType,
		Callback:     specTypeCallback,
	},
	keys.Pattern: {
		Msg:          "",
		PromptPrefix: "Insert the pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Pattern,
		Callback:     nil,
	},
	keys.Aql: {
		Msg:          "",
		PromptPrefix: "Insert the aql >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Aql,
		Callback:     nil,
	},
	keys.Target: {
		Msg:          "",
		PromptPrefix: "Insert the target >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Target,
		Callback:     nil,
	},
	keys.Props: {
		Msg:          "",
		PromptPrefix: "Enter \"key=value\" pairs separated by a semi-colon (key1=value1;key2=value2) >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Props,
		Callback:     nil,
	},
	keys.ExcludeProps: {
		Msg:          "",
		PromptPrefix: "Enter \"key=value\" pairs separated by a semi-colon (key1=value1;key2=value2) >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.ExcludeProps,
		Callback:     nil,
	},
	keys.Recursive: {
		Options:      corecommandutils.GetBoolSuggests(),
		Msg:          "",
		PromptPrefix: "Select if recursive" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Recursive,
		Callback:     nil,
	},
	keys.Exclusions: {
		Msg:          "",
		PromptPrefix: "Enter a comma separated list of exclusion patterns >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringArrayAnswer,
		MapKey:       keys.Exclusions,
		Callback:     nil,
	},
	keys.ArchiveEntries: {
		Msg:          "",
		PromptPrefix: "Insert archive-entries pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.ArchiveEntries,
		Callback:     nil,
	},
	keys.Build: {
		Msg:          "",
		PromptPrefix: "Insert build pattern >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Build,
		Callback:     nil,
	},
	keys.Bundle: {
		Msg:          "",
		PromptPrefix: "Insert bundle >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Bundle,
		Callback:     nil,
	},
	keys.SortBy: {
		Msg:          "",
		PromptPrefix: "Enter a comma separated list of sort-by values >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringArrayAnswer,
		MapKey:       keys.SortBy,
		Callback:     nil,
	},
	keys.SortOrder: {
		Options: []prompt.Suggest{
			{Text: keys.Asc, Description: ""},
			{Text: keys.Desc, Description: ""},
		},
		Msg:          "",
		PromptPrefix: "Select sort-order >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.SortOrder,
		Callback:     nil,
	},
	keys.Limit: {
		Msg:          "",
		PromptPrefix: "Insert limit >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Limit,
		Callback:     nil,
	},
	keys.Offset: {
		Msg:          "",
		PromptPrefix: "Insert offset >",
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Offset,
		Callback:     nil,
	},
	keys.Flat: {
		Options:      corecommandutils.GetBoolSuggests(),
		Msg:          "",
		PromptPrefix: "Select if flat" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Flat,
		Callback:     nil,
	},
	keys.ValidateSymlinks: {
		Options:      corecommandutils.GetBoolSuggests(),
		Msg:          "",
		PromptPrefix: "Select if should validate symlinks" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.ValidateSymlinks,
		Callback:     nil,
	},
	keys.Regexp: {
		Options:      corecommandutils.GetBoolSuggests(),
		Msg:          "",
		PromptPrefix: "Select if regexp" + corecommandutils.PressTabMsg,
		AllowVars:    false,
		Writer:       corecommandutils.WriteStringAnswer,
		MapKey:       keys.Regexp,
		Callback:     nil,
	},
}
