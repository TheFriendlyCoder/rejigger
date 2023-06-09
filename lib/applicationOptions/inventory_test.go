package applicationOptions

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseInventoryFile(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	// Given an empty temp folder
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	// And a sample config file
	outputFile := path.Join(tmpDir, ".rejig.inv.yml")
	expName := "test1"
	expType := TstGit
	expTypeStr := "git"
	expSource := "http://some/repo"
	invData := fmt.Sprintf(`
templates:
  - name: %s
    source: %s
    type: %s
`, expName, expSource, expTypeStr)
	fh, err := os.Create(outputFile)
	r.NoError(err)
	_, err = fh.WriteString(invData)
	r.NoError(err)
	r.NoError(fh.Close())

	// When we try parsing in the inventory
	fs := afero.NewOsFs()
	result, err := parseInventory(fs, outputFile)

	// The inventory should be successfully parsed
	r.NoError(err)
	r.Equal(1, len(result.Templates))
	a.Equal(expType, result.Templates[0].Type)
	a.Equal(expSource, result.Templates[0].GetSource())
	a.Equal(expName, result.Templates[0].GetName())
}
