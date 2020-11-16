package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupItems(t *testing.T) {
	slice := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	results := GroupItems(slice, 3)
	assert.ElementsMatch(t, results, [][]string{{"0", "1", "2"}, {"3", "4", "5"}, {"6", "7", "8"}, {"9"}})
}
