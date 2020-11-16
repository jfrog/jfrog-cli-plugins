package utils

import (
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

const DEFAULT_VALUE = "N/A"

func Optional(optionalValue string) (value string) {
	value = DEFAULT_VALUE
	if optionalValue != "" {
		value = optionalValue
	}
	return
}

func OptionalVcsUrl(vcs *buildinfo.Vcs) (value string) {
	value = DEFAULT_VALUE
	if vcs != nil {
		value = Optional(vcs.Url)
		if value != DEFAULT_VALUE && vcs.Revision != "" {
			value = strings.TrimSuffix(value, ".git") + "/commit/" + vcs.Revision
		}
	}
	return
}
