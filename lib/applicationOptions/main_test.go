package applicationOptions

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_successfulValidation(t *testing.T) {
	a := assert.New(t)

	tests := map[string]struct {
		templateOptions  []TemplateOptions
		inventoryOptions []InventoryOptions
	}{
		"Local template file system type": {
			templateOptions: []TemplateOptions{{
				Alias:  "My Template",
				Source: "https://some/location",
				Type:   TstLocal,
			}},
		},
		"Git template file system type": {
			templateOptions: []TemplateOptions{{
				Alias:  "My Template",
				Source: "https://some/location",
				Type:   TstGit,
			}},
		},
		"Local inventory file system type": {
			inventoryOptions: []InventoryOptions{{
				Namespace: "Fubar",
				Source:    "https://some/location",
				Type:      IstLocal,
			}},
		},
		"Git inventory file system type": {
			inventoryOptions: []InventoryOptions{{
				Namespace: "Fubar",
				Source:    "https://some/location",
				Type:      IstGit,
			}},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {

			opts := AppOptions{
				Templates:   data.templateOptions,
				Inventories: data.inventoryOptions,
			}
			err := opts.Validate()
			a.NoError(err)
		})
	}
}

func Test_successfulValidationEmptyConfig(t *testing.T) {
	r := require.New(t)

	opts := AppOptions{}

	err := opts.Validate()
	r.NoError(err, "Validation should have succeeded")
}

func Test_validationFailures(t *testing.T) {
	r := require.New(t)

	tests := map[string]struct {
		templateOptions  []TemplateOptions
		inventoryOptions []InventoryOptions
	}{
		"Template missing type": {
			templateOptions: []TemplateOptions{{
				Alias:  "My Template",
				Source: "https://some/location",
			}},
		},
		"Template missing alias": {
			templateOptions: []TemplateOptions{{
				Source: "https://some/location",
				Type:   TstLocal,
			}},
		},
		"Template missing source": {
			templateOptions: []TemplateOptions{{
				Alias: "My Template",
				Type:  TstGit,
			}},
		},
		"Inventory missing type": {
			inventoryOptions: []InventoryOptions{{
				Namespace: "Fubar",
				Source:    "https://some/location",
			}},
		},
		"Inventory missing source": {
			inventoryOptions: []InventoryOptions{{
				Namespace: "Fubar",
				Type:      IstLocal,
			}},
		},
		"Inventory missing namespsce": {
			inventoryOptions: []InventoryOptions{{
				Source: "https://some/location",
				Type:   IstLocal,
			}},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			opts := AppOptions{
				Templates:   data.templateOptions,
				Inventories: data.inventoryOptions,
			}
			err := opts.Validate()
			r.Error(err)
		})
	}
}

func Test_validationTemplateCompoundError(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	opts := AppOptions{
		Templates:   []TemplateOptions{{}},
		Inventories: []InventoryOptions{{}},
	}

	err := opts.Validate()
	r.Error(err)

	a.Contains(err.Error(), "template 0 source is undefined")
	a.Contains(err.Error(), "template 0 type is undefined")
	a.Contains(err.Error(), "template 0 alias is undefined")

	a.Contains(err.Error(), "inventory 0 source is undefined")
	a.Contains(err.Error(), "inventory 0 type is undefined")
	a.Contains(err.Error(), "inventory 0 namespace is undefined")
}

func Test_fromViper(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	tests := map[string]struct {
		TypeStr  string
		TypeEnum TemplateSourceType
		Source   string
		Alias    string
	}{
		"Default local template": {
			TypeStr:  "local",
			TypeEnum: TstLocal,
			Source:   "/path/to/template",
			Alias:    "test1",
		},
		"Default Git template": {
			TypeStr:  "git",
			TypeEnum: TstGit,
			Source:   "https://some/url",
			Alias:    "test1",
		},
		"Default Unsupported template": {
			TypeStr:  "other",
			TypeEnum: TstUnknown,
			Source:   "https://some/url",
			Alias:    "test1",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Given an empty temp folder
			tmpDir, err := os.MkdirTemp("", "")
			r.NoError(err)
			defer os.RemoveAll(tmpDir)

			// And a sample application options file
			cfgFilePath := path.Join(tmpDir, "sample.yml")
			fh, err := os.Create(cfgFilePath)
			r.NoError(err)
			cfgData := fmt.Sprintf(`
templates:
  - type: %s
    source: %s
    alias: %s
`, data.TypeStr, data.Source, data.Alias)
			_, err = fh.WriteString(cfgData)
			r.NoError(err)

			// Point viper to our config file
			v := viper.New()
			v.SetConfigFile(cfgFilePath)
			r.NoError(v.ReadInConfig())

			// When we try instantiating our app options from Viper
			options, err := FromViper(v)

			// We expect the operation to succeed
			r.NoError(err)
			a.Equal(1, len(options.Templates))
			a.Equal(data.Alias, options.Templates[0].Alias)
			a.Equal(data.Source, options.Templates[0].Source)
			a.Equal(data.TypeEnum, options.Templates[0].Type)
		})
	}
}

func Test_fromViperParseFail(t *testing.T) {
	r := require.New(t)

	tests := map[string]struct {
		config string
	}{
		"Valid yaml with missing template type": {
			config: `
templates:
  - source: /some/path2
    alias: test2
`,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Given an empty temp folder
			tmpDir, err := os.MkdirTemp("", "")
			r.NoError(err)
			defer os.RemoveAll(tmpDir)

			// And a sample application options file
			cfgFilePath := path.Join(tmpDir, "sample.yml")
			fh, err := os.Create(cfgFilePath)
			r.NoError(err)
			_, err = fh.WriteString(data.config)
			r.NoError(err)

			// Point viper to our config file
			v := viper.New()
			v.SetConfigFile(cfgFilePath)
			r.NoError(v.ReadInConfig())

			// When we try instantiating our app options from Viper
			_, err = FromViper(v)

			// The operation should fail
			r.Error(err)

		})
	}
}