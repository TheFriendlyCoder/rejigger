package lib

import (
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
//									DATA STRUCTURES

// VersionData stores version information about the template
type VersionData struct {
	// Schema version describing the format of the manifest file and its contents
	Schema version.Version `yaml:"schema"`
	// Jigger the minimum version of the Rejigger app needed to process to the template
	Jigger version.Version `yaml:"rejigger"`
	// Template the version number associated with the current template
	Template version.Version `yaml:"template"`
}

// ArgData metadata describing the input args supported by the template. Values for these
// args must be provided by the user to customize the content produced by the template
type ArgData struct {
	// Name of the argument, exactly as will be used in the template contents
	Name string `yaml:"name"`
	// Description descriptive text explaining the purpose of the argument
	Description string `yaml:"description"`
}

// TemplateData metadata describing the template being processed
type TemplateData struct {
	// Args list of input parameters supported by the template. These provide user configurable
	// options that customize the content produced by the template
	Args []ArgData `yaml:"args"`
}

// ManifestData parsed content of the manifest file associated with a template
type ManifestData struct {
	// Versions version identifiers for various aspects of the template
	Versions VersionData `yaml:"versions"`
	// Template metadata describing the template
	Template TemplateData `yaml:"template"`
	// MiscParams all unparsed values in the manifest will be dumped into a simple map structure
	MiscParams map[string]interface{} `yaml:"-,flow"`
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
//									PARSING LOGIC

// UnmarshalYAML custom YAML decoding method compatible with the YAML parsing library
func (m *ManifestData) UnmarshalYAML(value *yaml.Node) error {
	// Start by parsing version info from the manifest file
	var versionFields struct {
		Versions VersionData `yaml:"versions"`
	}
	if err := value.Decode(&versionFields); err != nil {
		return errors.Wrap(err, "Failed to parse version information")
	}
	m.Versions = versionFields.Versions

	// Then parse template metadata
	var templateFields struct {
		Template TemplateData `yaml:"template"`
	}
	if err := value.Decode(&templateFields); err != nil {
		return errors.Wrap(err, "Failed to parse template metadata")
	}
	m.Template = templateFields.Template

	// Dump all remaining content into a simple map
	var remaining map[string]interface{}
	if err := value.Decode(&remaining); err != nil {
		return errors.Wrap(err, "Failed to parse additional config options")
	}
	delete(remaining, "versions")
	delete(remaining, "template")
	m.MiscParams = remaining
	return nil
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
//									PUBLIC INTERFACE

// ParseManifest parses a template manifest file and returns a reference to
// the parsed representation of the contents of the file
func ParseManifest(path string) (*ManifestData, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open manifest file")
	}

	retval := &ManifestData{}
	// TODO: Find some way to get "Strict" mode to work properly (aka: KnownFields in v3)
	//		https://github.com/go-yaml/yaml/issues/460
	//		https://github.com/go-yaml/yaml/issues/642
	err = yaml.Unmarshal(buf, retval)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse YAML content from manifest file")
	}

	return retval, nil
}