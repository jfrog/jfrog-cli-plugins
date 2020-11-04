package commands

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var trimPrefixProvider = []struct {
	pattern  string
	path     string
	expected string
}{
	{"abc", "abc", "abc"},
	{"abc", "abc/def", "def"},
	{"abc/", "abc/def", "def"},
	{"abc/def", "abc/def", "def"},
	{"abc/def/", "abc/def", "def"},
}

func TestTrimPrefixFromPath(t *testing.T) {
	for _, triple := range trimPrefixProvider {
		t.Run(fmt.Sprintf("%v", triple), func(t *testing.T) {
			assert.Equal(t, triple.expected, trimFoldersFromPath(triple.pattern, triple.path))
		})
	}
}
