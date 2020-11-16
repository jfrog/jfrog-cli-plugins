package utils

import (
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

const DefaultValue = "N/A"

func Optional(optionalValue string) (value string) {
	value = DefaultValue
	if optionalValue != "" {
		value = optionalValue
	}
	return
}

func OptionalVcsUrl(vcs *buildinfo.Vcs) (value string) {
	value = DefaultValue
	if vcs != nil {
		value = Optional(vcs.Url)
		if value != DefaultValue && vcs.Revision != "" {
			value = strings.TrimSuffix(value, ".git") + "/commit/" + vcs.Revision
		}
	}
	return
}
