package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	repo = "testRepo"
	time = "17mo"
	aql  = `items.find({` +
		`"type":"file",` +
		`"repo":"` + repo + `",` +
		`"$or":[` +
		`{"$and":[` +
		`{"modified":{"$before":"` + time + `"}},` +
		`{"stat.downloaded":{"$before":"` + time + `"}},` +
		`{"stat.downloads":{"$gt":"0"}}` +
		`]},` +
		`{"$and":[` +
		`{"modified":{"$before":"` + time + `"}},` +
		`{"stat.downloads":{"$eq":null}}` +
		`]}` +
		`]` +
		`})`
)

func TestBuildAQL(t *testing.T) {
	conf := &cleanConfiguration{
		repository:       repo,
		noDownloadedTime: time,
	}
	assert.Equal(t, buildAQL(conf), aql)
}

func TestParseTimeFlags(t *testing.T) {
	var timeFlags = []struct {
		timeUnit  string
		timeValue string
		expected  string
	}{
		{"day", "3", "3d"},
		{"month", "99", "99mo"},
		{"year", "7", "7y"},
		{"non-valid-unit", "3", ""},
		{"day", "non-valid-int", ""},
	}
	for _, v := range timeFlags {
		result, _ := parseTimeFlags(v.timeValue, v.timeUnit)
		if result != v.expected {
			t.Errorf("parseTimeFlags(%q,%q) => '%s', expected '%s'", v.timeValue, v.timeUnit, result, v.expected)
		}
	}

}
