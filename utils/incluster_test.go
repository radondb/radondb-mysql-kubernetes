package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSleepFlag(t *testing.T) {

	{
		var file string = `/file_not_exist`
		// file not exist
		_, err := os.Stat(file)
		assert.Equal(t, false, !os.IsNotExist(err))
	}
	{
		var file string = `/`
		// file exist
		_, err := os.Stat(file)
		assert.Equal(t, true, !os.IsNotExist(err))
	}
}

func TestIsMySQLRunning(t *testing.T) {
	{
		want := false
		got := IsMySQLRunning()
		assert.Equal(t, want, got)
	}
}
