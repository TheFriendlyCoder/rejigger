package lib

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DirExists(t *testing.T) {
	result := DirExists(".")
	require.True(t, result)
}

func Test_DirNotExists(t *testing.T) {
	r := require.New(t)

	// Given an empty temp folder
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	// When we request a check on a file system object that doesn't exist
	result := DirExists(path.Join(tmpDir, "fubar"))
	r.False(result)
}

func Test_DirExistsButIsFile(t *testing.T) {
	r := require.New(t)

	// Given a test folder with a file in it
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	var filename = path.Join(tmpDir, "fubar")
	_, err = os.Create(filename)
	r.NoError(err)

	// When we check to see if that file exists as a directory
	result := DirExists(filename)

	// Then, the response should be False because the dir already exists
	r.False(result)
}

func Test_InvalidDirExists(t *testing.T) {
	// The stat method doesn't accept strings with nested null characters in it
	// but we expect that error to get swallowed and just return a false result
	result := DirExists(".\x00asdf")
	require.False(t, result)
}
