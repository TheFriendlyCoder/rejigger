package templateManager

import (
	"io/fs"
	"os"
	"path/filepath"

	ao "github.com/TheFriendlyCoder/rejigger/lib/applicationOptions"
	"github.com/flosch/pongo2/v6"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// generate applies a set of user defined options (ie: the 'context') to a set of template
// files stored in srcPath, and produces a complete project in the targetPath with the
// user defined parameters applied throughout
func generate(srcFS afero.Fs, templateOptions ao.TemplateOptions, targetPath string, context map[string]any) error {
	rootDir := templateOptions.GetProjectRoot()

	// loop through all files
	err := afero.Walk(srcFS, rootDir, func(path string, info fs.FileInfo, err error) error {
		// If walk encountered an error attempting to enumerate the file system object
		// we are processing, it tells us here. For now we just assume we can not proceed
		// if we hit this condition.
		// TODO: Consider how best to handle error conditions
		//		https://pkg.go.dev/io/fs#WalkDirFunc
		if err != nil {
			return err
		}

		// Skip excluded files
		if templateOptions.IsFileExcluded(path) {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return errors.WithStack(err)
		}
		// Skip processing the root dir
		if relPath == "." {
			return nil
		}
		// Skip Rejigger manifest file
		if relPath == ".rejig.yml" {
			return nil
		}

		// apply template to the Path being processed
		newOutputPath, err := processPath(relPath, targetPath, context)
		if err != nil {
			return err
		}

		// Generate output content
		if info.IsDir() {
			err = createOutputDir(newOutputPath, info.Mode())
		} else {
			err = createOutputFile(srcFS, path, newOutputPath, info.Mode(), context)
		}
		return err
	})
	if err != nil {
		return errors.Wrap(err, "Failed generating project")
	}
	return nil
}

// processPath applies template processor to a folder name
func processPath(relPath string, targetPath string, context map[string]any) (string, error) {
	tpl, err := pongo2.FromString(relPath)
	if err != nil {
		return "", errors.Wrap(err, "Failed to load template from path "+relPath)
	}
	newDirName, err := tpl.Execute(context)
	if err != nil {
		return "", errors.Wrap(err, "Failed applying template to path "+relPath)
	}
	return filepath.Join(targetPath, newDirName), nil
}

// createOutputDir applies template processor to a directory
func createOutputDir(newOutputPath string, mode os.FileMode) error {
	// Make sure to preserve the file mode
	err := os.MkdirAll(newOutputPath, mode)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// createOutputFile applies template processor to a file
func createOutputFile(srcFS afero.Fs, originalPath string, newOutputPath string, mode os.FileMode, context map[string]any) error {
	// Read in the original file contents
	var data []byte
	data, err := afero.ReadFile(srcFS, originalPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Apply our template to the file contents
	tpl, err := pongo2.FromString(string(data))
	if err != nil {
		return errors.Wrap(err, "Error loading template file "+originalPath)
	}

	var newData string
	newData, err = tpl.Execute(context)
	if err != nil {
		return errors.Wrap(err, "Error applying template file "+originalPath)
	}

	// Write processed output to new file location
	// making sure to preserve the file mode in the process
	err = os.WriteFile(newOutputPath, []byte(newData), mode)
	if err != nil {
		return errors.Wrap(err, "Failed to generate project file "+newOutputPath)
	}
	return nil
}
