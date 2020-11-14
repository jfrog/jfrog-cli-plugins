package commands

import (
	"fmt"
	"testing"

	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/stretchr/testify/assert"
)

var createAqlProvider = []struct {
	pattern  string
	expected string
}{
	{"repository/filename", `{"repo":"repository","path":".","name":"filename"}`},
	{"repository/dir/filename", `{"repo":"repository","path":"dir","name":"filename"}`},
	{"repository/dir1/dir2/filename", `{"repo":"repository","path":"dir1/dir2","name":"filename"}`},
}

func TestCreateAql(t *testing.T) {
	for _, triple := range createAqlProvider {
		t.Run(fmt.Sprintf("%v", triple), func(t *testing.T) {
			expected := utils.Aql{
				ItemsFind: triple.expected,
			}
			assert.Equal(t, expected, createAql(triple.pattern))
		})
	}
}
