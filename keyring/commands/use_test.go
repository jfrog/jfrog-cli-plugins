package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindFirstFlagIndex(t *testing.T) {
	tests := []testFindIndex{
		{args: []string{"rt", "ping"}, index: 2},
		{args: []string{"rt", "u", "--flat"}, index: 2},
		{args: []string{"rt", "u", "--flat", "./*"}, index: 2},
		{args: []string{"rt", "u", "./*", "--flat"}, index: 3},
		{args: []string{"rt", "u", "./*", "--flat", "--recursive"}, index: 3},
	}

	for _, test := range tests {
		assert.Equal(t, test.index, findFirstFlagIndex(test.args))
	}
}

type testFindIndex struct {
	args  []string
	index int
}

func TestBuildArgs(t *testing.T) {
	conf := getTestConf()
	tests := []testBuildArgs{
		{inArgs: []string{"rt", "ping"}, outArgs: []string{"rt", "ping", "--url", conf.Url, "--user", conf.User, "--password", conf.Password}},
		{inArgs: []string{"rt", "u", "--flat"}, outArgs: []string{"rt", "u", "--url", conf.Url, "--user", conf.User, "--password", conf.Password, "--flat"}},
		{inArgs: []string{"rt", "u", "--flat", "./*"}, outArgs: []string{"rt", "u", "--url", conf.Url, "--user", conf.User, "--password", conf.Password, "--flat", "./*"}},
		{inArgs: []string{"rt", "u", "./*", "--flat"}, outArgs: []string{"rt", "u", "./*", "--url", conf.Url, "--user", conf.User, "--password", conf.Password, "--flat"}},
		{inArgs: []string{"rt", "u", "./*", "--flat", "--recursive"}, outArgs: []string{"rt", "u", "./*", "--url", conf.Url, "--user", conf.User, "--password", conf.Password, "--flat", "--recursive"}},
	}

	for _, test := range tests {
		args := buildArgs(test.inArgs, &conf)
		assert.True(t, equals(test.outArgs, args))
	}
}

type testBuildArgs struct {
	inArgs  []string
	outArgs []string
}
