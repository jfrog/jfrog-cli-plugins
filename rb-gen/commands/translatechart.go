package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	rtcommands "github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/distribution"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/plugins/components"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	servicesutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	distributionServicesUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

const (
	OfferConfig = "JFROG_CLI_OFFER_CONFIG"
	CI          = "CI"
)

type TranslateChartCommand struct {
	rtDetails            *config.ArtifactoryDetails
	releaseBundlesParams distributionServicesUtils.ReleaseBundleParams
	sourceChartPath      string
	valuesFilePath       string
	dockerRepo           string
	dryRun               bool
}

func GetReleaseBundleTranslateChartCommand() components.Command {
	return components.Command{
		Name:        "from-chart",
		Description: "Generate a release bundle from an existing Helm chart.",
		Aliases:     []string{"fc"},
		Arguments:   getReleaseBundleTranslateChartArguments(),
		Flags:       getReleaseBundleTranslateChartFlags(),
		EnvVars:     []components.EnvVar{},
		Action: func(c *components.Context) error {
			return releaseBundleTranslateChartCmd(c)
		},
	}
}

func getReleaseBundleTranslateChartArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "release bundle name",
			Description: "The name of the release bundle.",
		},
		{
			Name:        "release bundle version",
			Description: "The release bundle version.",
		},
	}
}

func getReleaseBundleTranslateChartFlags() []components.Flag {
	return []components.Flag{
		components.StringFlag{
			Name:        "url",
			Description: "Artifactory URL.",
		},
		components.StringFlag{
			Name:        "dist-url",
			Description: "Distribution URL.",
		},
		components.StringFlag{
			Name:        "user",
			Description: "Artifactory username.",
		},
		components.StringFlag{
			Name:        "password",
			Description: "Artifactory password.",
		},
		components.StringFlag{
			Name:        "apikey",
			Description: "Artifactory API key.",
		},
		components.StringFlag{
			Name:        "access-token",
			Description: "Artifactory access token.",
		},
		components.StringFlag{
			Name:        "ssh-passphrase",
			Description: "SSH key passphrase.",
		},
		components.StringFlag{
			Name:        "ssh-key-path",
			Description: "SSH key file path.",
		},
		components.StringFlag{
			Name:        "server-id",
			Description: "Artifactory server ID configured using the config command.",
		},
		components.StringFlag{
			Name:        "chart-path",
			Description: "Path to a Helm chart in Artifactory, which should be translated to a release bundle.",
			Mandatory:   true,
		},
		components.StringFlag{
			Name:        "values-file",
			Description: "Local path to helm values file. Used when rendering the chart",
			Mandatory:   false,
		},
		components.StringFlag{
			Name:        "docker-repo",
			Description: "A Docker repository containing all the Docker images the Helm chart requires.",
			Mandatory:   true,
		},
		components.BoolFlag{
			Name:        "dry-run",
			Description: "Set to true to disable communication with JFrog Distribution.",
		},
		components.BoolFlag{
			Name:        "sign",
			Description: "If set to true, automatically signs the release bundle version.",
		},
		components.StringFlag{
			Name:        "desc",
			Description: "Description of the release bundle.",
		},
		components.StringFlag{
			Name:        "release-notes-path",
			Description: "Path to a file describes the release notes for the release bundle version.",
		},
		components.StringFlag{
			Name:         "release-notes-syntax",
			Description:  "The syntax for the release notes. Can be one of 'markdown', 'asciidoc', or 'plain_text'.",
			DefaultValue: "plain_text",
		},
		components.StringFlag{
			Name:        "exclusions",
			Description: "Semicolon-separated list of exclusions. Exclusions can include the * and the ? wildcards.",
		},
		components.StringFlag{
			Name:        "passphrase",
			Description: "The passphrase for the signing key.",
		},
		components.StringFlag{
			Name:        "repo",
			Description: "A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one.",
		},
	}
}

func releaseBundleTranslateChartCmd(c *components.Context) error {
	chartpath := c.GetStringFlagValue("chart-path")
	valuesFilePath := c.GetStringFlagValue("values-file")
	dockerrepo := c.GetStringFlagValue("docker-repo")
	if !(len(c.Arguments) == 2 && chartpath != "" && dockerrepo != "") {
		return errors.New("Wrong number of arguments.")
	}
	params, err := createReleaseBundleCreateUpdateParams(c, c.Arguments[0], c.Arguments[1])
	if err != nil {
		return err
	}
	translateChartCmd := NewTranslateChartCommand()
	rtDetails, err := createArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}
	translateChartCmd.SetRtDetails(rtDetails).SetReleaseBundleCreateParams(params).SetSourceChartPath(chartpath).SetValuesFilePath(valuesFilePath).SetDockerRepo(dockerrepo).SetDryRun(c.GetBoolFlagValue("dry-run"))
	return rtcommands.Exec(translateChartCmd)
}

func createReleaseBundleCreateUpdateParams(c *components.Context, bundleName, bundleVersion string) (distributionServicesUtils.ReleaseBundleParams, error) {
	releaseBundleParams := distributionServicesUtils.NewReleaseBundleParams(bundleName, bundleVersion)
	releaseBundleParams.SignImmediately = c.GetBoolFlagValue("sign")
	releaseBundleParams.StoringRepository = c.GetStringFlagValue("repo")
	releaseBundleParams.GpgPassphrase = c.GetStringFlagValue("passphrase")
	releaseBundleParams.Description = c.GetStringFlagValue("desc")
	notespath := c.GetStringFlagValue("release-notes-path")
	if notespath != "" {
		bytes, err := ioutil.ReadFile(notespath)
		if err != nil {
			return releaseBundleParams, errorutils.CheckError(err)
		}
		releaseBundleParams.ReleaseNotes = string(bytes)
		releaseBundleParams.ReleaseNotesSyntax, err = populateReleaseNotesSyntax(c)
		if err != nil {
			return releaseBundleParams, err
		}
	}
	return releaseBundleParams, nil
}

func populateReleaseNotesSyntax(c *components.Context) (distributionServicesUtils.ReleaseNotesSyntax, error) {
	// If release notes syntax is set, use it
	releaseNotexSyntax := c.GetStringFlagValue("release-notes-syntax")
	if releaseNotexSyntax != "" {
		switch releaseNotexSyntax {
		case "markdown":
			return distributionServicesUtils.Markdown, nil
		case "asciidoc":
			return distributionServicesUtils.Asciidoc, nil
		case "plain_text":
			return distributionServicesUtils.PlainText, nil
		default:
			return distributionServicesUtils.PlainText, errorutils.CheckError(errors.New("--release-notes-syntax must be one of: markdown, asciidoc or plain_text."))
		}
	}
	// If the file extension is ".md" or ".markdown", use the markdonwn syntax
	extension := strings.ToLower(filepath.Ext(c.GetStringFlagValue("release-notes-path")))
	if extension == ".md" || extension == ".markdown" {
		return distributionServicesUtils.Markdown, nil
	}
	return distributionServicesUtils.PlainText, nil
}

func createArtifactoryDetailsByFlags(c *components.Context) (*config.ArtifactoryDetails, error) {
	artDetails, err := createArtifactoryDetails(c, true)
	if err != nil {
		return nil, err
	}
	if artDetails.DistributionUrl == "" {
		return nil, errors.New("the --dist-url option is mandatory")
	}
	if artDetails.Url == "" {
		return nil, errors.New("the --url option is mandatory")
	}

	return artDetails, nil
}

func createArtifactoryDetails(c *components.Context, includeConfig bool) (details *config.ArtifactoryDetails, err error) {
	if includeConfig {
		details, err := offerConfig(c)
		if err != nil {
			return nil, err
		}
		if details != nil {
			return details, err
		}
	}
	details = new(config.ArtifactoryDetails)
	details.Url = c.GetStringFlagValue("url")
	details.DistributionUrl = c.GetStringFlagValue("dist-url")
	details.ApiKey = c.GetStringFlagValue("apikey")
	details.User = c.GetStringFlagValue("user")
	details.Password = c.GetStringFlagValue("password")
	details.SshKeyPath = c.GetStringFlagValue("ssh-key-path")
	details.SshPassphrase = c.GetStringFlagValue("ssh-passphrase")
	details.AccessToken = c.GetStringFlagValue("access-token")
	details.ClientCertPath = c.GetStringFlagValue("client-cert-path")
	details.ClientCertKeyPath = c.GetStringFlagValue("client-cert-key-path")
	details.ServerId = c.GetStringFlagValue("server-id")
	details.InsecureTls = c.GetBoolFlagValue("insecure-tls")
	if details.ApiKey != "" && details.User != "" && details.Password == "" {
		// The API Key is deprecated, use password option instead.
		details.Password = details.ApiKey
		details.ApiKey = ""
	}

	if includeConfig && !credentialsChanged(details) {
		confDetails, err := rtcommands.GetConfig(details.ServerId, false)
		if err != nil {
			return nil, err
		}

		if details.Url == "" {
			details.Url = confDetails.Url
		}
		if details.DistributionUrl == "" {
			details.DistributionUrl = confDetails.DistributionUrl
		}

		if !isAuthMethodSet(details) {
			if details.ApiKey == "" {
				details.ApiKey = confDetails.ApiKey
			}
			if details.User == "" {
				details.User = confDetails.User
			}
			if details.Password == "" {
				details.Password = confDetails.Password
			}
			if details.SshKeyPath == "" {
				details.SshKeyPath = confDetails.SshKeyPath
			}
			if details.AccessToken == "" {
				details.AccessToken = confDetails.AccessToken
			}
			if details.RefreshToken == "" {
				details.RefreshToken = confDetails.RefreshToken
			}
			if details.TokenRefreshInterval == coreutils.TokenRefreshDisabled {
				details.TokenRefreshInterval = confDetails.TokenRefreshInterval
			}
			if details.ClientCertPath == "" {
				details.ClientCertPath = confDetails.ClientCertPath
			}
			if details.ClientCertKeyPath == "" {
				details.ClientCertKeyPath = confDetails.ClientCertKeyPath
			}
		}
	}
	details.Url = clientutils.AddTrailingSlashIfNeeded(details.Url)
	details.DistributionUrl = clientutils.AddTrailingSlashIfNeeded(details.DistributionUrl)

	err = config.CreateInitialRefreshableTokensIfNeeded(details)
	return
}

func offerConfig(c *components.Context) (*config.ArtifactoryDetails, error) {
	var exists bool
	exists, err := config.IsArtifactoryConfExists()
	if err != nil || exists {
		return nil, err
	}

	var ci bool
	if ci, err = clientutils.GetBoolEnvValue(CI, false); err != nil {
		return nil, err
	}
	var offerConfig bool
	if offerConfig, err = clientutils.GetBoolEnvValue(OfferConfig, !ci); err != nil {
		return nil, err
	}
	if !offerConfig {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return nil, nil
	}

	msg := fmt.Sprintf("To avoid this message in the future, set the %s environment variable to false.\n"+
		"The CLI commands require the Artifactory URL and authentication details\n"+
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n"+
		"You can also configure these parameters later using the 'jfrog rt c' command.\n"+
		"Configure now?", OfferConfig)
	confirmed := InteractiveConfirm(msg)
	if !confirmed {
		config.SaveArtifactoryConf(make([]*config.ArtifactoryDetails, 0))
		return nil, nil
	}
	details, err := createArtifactoryDetails(c, false)
	if err != nil {
		return nil, err
	}
	encPassword := c.GetBoolFlagValue("enc-password")
	configCmd := rtcommands.NewConfigCommand().SetDefaultDetails(details).SetInteractive(true).SetEncPassword(encPassword)
	err = configCmd.Config()
	if err != nil {
		return nil, err
	}

	return configCmd.RtDetails()
}

func InteractiveConfirm(message string) bool {
	var confirm string
	fmt.Print(message + " (y/n): ")
	fmt.Scanln(&confirm)
	return confirmAnswer(confirm)
}

func confirmAnswer(answer string) bool {
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes"
}

func credentialsChanged(details *config.ArtifactoryDetails) bool {
	return details.Url != "" || details.User != "" || details.Password != "" ||
		details.ApiKey != "" || details.SshKeyPath != "" || details.AccessToken != ""
}

func isAuthMethodSet(details *config.ArtifactoryDetails) bool {
	return (details.User != "" && details.Password != "") || details.SshKeyPath != "" || details.ApiKey != "" || details.AccessToken != ""
}

func NewTranslateChartCommand() *TranslateChartCommand {
	return &TranslateChartCommand{}
}

func (tc *TranslateChartCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *TranslateChartCommand {
	tc.rtDetails = rtDetails
	return tc
}

func (tc *TranslateChartCommand) SetReleaseBundleCreateParams(params distributionServicesUtils.ReleaseBundleParams) *TranslateChartCommand {
	tc.releaseBundlesParams = params
	return tc
}

func (tc *TranslateChartCommand) SetSourceChartPath(sourceChartPath string) *TranslateChartCommand {
	tc.sourceChartPath = sourceChartPath
	return tc
}

func (tc *TranslateChartCommand) SetValuesFilePath(valuesFilePath string) *TranslateChartCommand {
	tc.valuesFilePath = valuesFilePath
	return tc
}

func (tc *TranslateChartCommand) SetDockerRepo(dockerRepo string) *TranslateChartCommand {
	tc.dockerRepo = dockerRepo
	return tc
}

func (tc *TranslateChartCommand) SetDryRun(dryRun bool) *TranslateChartCommand {
	tc.dryRun = dryRun
	return tc
}

func (tc *TranslateChartCommand) Run() error {
	body, err := readFileFromArtifactory(tc.rtDetails, tc.sourceChartPath)
	if err != nil {
		return err
	}
	defer body.Close()
	chrt, err := chartutil.LoadArchive(body)
	if err != nil {
		return err
	}
	specstr, expected, err := createFilespec(chrt, extractRepo(tc.sourceChartPath), tc.dockerRepo, tc.valuesFilePath)
	if err != nil {
		return err
	}
	specfiles := new(spec.SpecFiles)
	err = json.Unmarshal([]byte(specstr), specfiles)
	if err != nil {
		return err
	}
	createBundle := distribution.NewReleaseBundleCreateCommand()
	createBundle.SetRtDetails(tc.rtDetails)
	createBundle.SetReleaseBundleCreateParams(tc.releaseBundlesParams)
	createBundle.SetSpec(specfiles)
	createBundle.SetDryRun(tc.dryRun)
	err = createBundle.Run()
	if err != nil {
		return err
	}
	actual, err := checkExisting(tc.rtDetails, specfiles)
	if err != nil {
		return err
	}
	missing := make([]string, 0)
	fmt.Println("Found:")
	for _, name := range expected {
		line := "/" + name
		found := false
		if !strings.HasSuffix(line, ".tgz") {
			line = strings.ReplaceAll(line, ":", "/") + "/"
		}
		for _, path := range actual {
			if strings.Contains(path, line) {
				fmt.Println("- " + name)
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, name)
		}
	}
	fmt.Println("Missing:")
	if len(missing) <= 0 {
		missing = append(missing, "none")
	}
	for _, line := range missing {
		fmt.Println("- " + line)
	}
	return nil
}

func (tc *TranslateChartCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return tc.rtDetails, nil
}

func (tc *TranslateChartCommand) CommandName() string {
	return "rt_translate_chart"
}

func extractRepo(path string) string {
	if path[0] == '/' {
		path = path[1:]
	}
	return strings.SplitN(path, "/", 2)[0]
}

func readFileFromArtifactory(artDetails *config.ArtifactoryDetails, downloadPath string) (io.ReadCloser, error) {
	downloadUrl := urlAppend(artDetails.Url, downloadPath)
	auth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	securityDir, err := coreutils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	client, err := rthttpclient.ArtifactoryClientBuilder().
		SetCertificatesPath(securityDir).
		SetInsecureTls(artDetails.InsecureTls).
		SetServiceDetails(&auth).
		Build()
	if err != nil {
		return nil, err
	}
	httpClientDetails := auth.CreateHttpClientDetails()
	body, resp, err := client.ReadRemoteFile(downloadUrl, &httpClientDetails)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New(resp.Status + " received when attempting to download " + downloadUrl))
	}
	return body, err
}

func createFilespec(chrt *chart.Chart, helmrepo, dockerrepo string, valuesFilePath string) (string, []string, error) {
	flist := make([]string, 0)
	chartConfig := &chart.Config{Raw: "{}"}
	if valuesFilePath != "" {
		yfile, err := ioutil.ReadFile(valuesFilePath)
		if err != nil {
			return "", flist, err
		}
		chartConfig = &chart.Config{Raw: string(yfile)}
	}
	files, err := renderutil.Render(chrt, chartConfig, renderutil.Options{})
	if err != nil {
		return "", flist, err
	}
	spec := "{\"files\":["
	lines := extractImages(files)
	for _, line := range sortStringMap(lines) {
		splits := strings.SplitN(line, "/", 2)
		image := splits[len(splits)-1]
		cname := strings.ReplaceAll(splits[len(splits)-1], ":", "/") + "/"
		path1, _ := json.Marshal(dockerrepo + "/" + cname)
		path2, _ := json.Marshal(dockerrepo + "/*/" + cname)
		spec = spec + "{\"pattern\":" + string(path1) + "},"
		spec = spec + "{\"pattern\":" + string(path2) + "},"
		flist = append(flist, image)
	}
	deps := map[string]*chart.Chart{}
	crawlRequirements(deps, chrt)
	for _, c := range sortChartMap(deps) {
		cname := c.Metadata.Name + "-" + c.Metadata.Version + ".tgz"
		path1, _ := json.Marshal(helmrepo + "/" + cname)
		path2, _ := json.Marshal(helmrepo + "/*/" + cname)
		spec = spec + "{\"pattern\":" + string(path1) + "},"
		spec = spec + "{\"pattern\":" + string(path2) + "},"
		flist = append(flist, cname)
	}
	spec = spec[:len(spec)-1]
	spec = spec + "]}"
	return spec, flist, nil
}

func checkExisting(rtDetails *config.ArtifactoryDetails, spec *spec.SpecFiles) ([]string, error) {
	flist := make([]string, 0)
	cmd := generic.NewSearchCommand()
	cmd.SetRtDetails(rtDetails).SetSpec(spec)
	results, err := cmd.Search()
	if err != nil {
		return flist, err
	}
	for result := new(servicesutils.ResultItem); results.NextRecord(result) == nil; result = new(servicesutils.ResultItem) {
		flist = append(flist, result.Path)
	}
	return flist, nil
}

func sortStringMap(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	vals := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, in[k])
	}
	return vals
}

func sortChartMap(in map[string]*chart.Chart) []*chart.Chart {
	keys := make([]string, 0, len(in))
	vals := make([]*chart.Chart, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, in[k])
	}
	return vals
}

func urlAppend(url, path string) string {
	if url[len(url)-1] != '/' {
		url = url + "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return url + path
}

func extractImages(files map[string]string) map[string]string {
	lines := map[string]string{}
	for _, v := range files {
		for _, line := range strings.Split(v, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "image:") {
				continue
			}
			line = strings.TrimPrefix(line, "image:")
			line = strings.TrimSpace(line)
			if len(line) <= 0 {
				continue
			}
			if line[0] == '\'' && line[len(line)-1] == '\'' ||
				line[0] == '"' && line[len(line)-1] == '"' {
				line = line[1 : len(line)-1]
			}
			lines[line] = line
		}
	}
	return lines
}

func crawlRequirements(reqs map[string]*chart.Chart, chrt *chart.Chart) {
	reqs[chrt.Metadata.Name] = chrt
	for _, req := range chrt.GetDependencies() {
		crawlRequirements(reqs, req)
	}
}
