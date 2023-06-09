package cmd

import (
	"bytes"
	"context"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"

	"github.com/TheFriendlyCoder/rejigger/cmd/shared"
	ao "github.com/TheFriendlyCoder/rejigger/lib/applicationOptions"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setHome helper function that overloads the default user home folder
// to point to a custom location for our test
func setHome(t *testing.T, newHome string) string {
	envVar := "HOME"
	switch runtime.GOOS {
	case "windows":
		envVar = "USERPROFILE"
	case "plan9":
		envVar = "home"
	}
	oldHome, err := os.UserHomeDir()
	require.NoError(t, err)
	require.NoError(t, os.Setenv(envVar, newHome))
	return oldHome
}

// restoreHome restores the users home folder configuration
// used as a deferred operation in unit tests to undo the changes
// made by the setHome helper method
func restoreHome(t *testing.T, newHome string) {
	envVar := "HOME"
	switch runtime.GOOS {
	case "windows":
		envVar = "USERPROFILE"
	case "plan9":
		envVar = "home"
	}
	require.NoError(t, os.Setenv(envVar, newHome))
}

func Test_helpCommand(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	// Redirect our user home folder to an empty temp dir
	// to force the command to load default options
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// When we run the help command
	rootCmd := RootCmd()
	actual := new(bytes.Buffer)
	rootCmd.SetOut(actual)
	rootCmd.SetErr(actual)

	rootCmd.SetArgs([]string{"help"})
	err = Execute(&rootCmd)

	r.NoError(err)
	a.Contains(actual.String(), "rejigger")
}

func Test_rootSetupValidation(t *testing.T) {
	// We set up the app environment through Execute now, and the Cobra code should
	// have some sanity checks in the PreRun to validate that we haven't messed something
	// up

	r := require.New(t)
	// Redirect our user home folder to an empty temp dir
	// to ensure test doesn't try to use the current users
	// config file
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// Construct partial app contexts to test various scenarios
	// TODO: Find some way to detect number of values in ContextKey enum
	//		 and then make sure we are testing every permutation here
	ctxNoViper := context.Background()
	ctxNoViper = context.WithValue(ctxNoViper, shared.CkOptions, ao.AppOptions{})

	ctxNoAppOptions := context.Background()
	ctxNoAppOptions = context.WithValue(ctxNoAppOptions, shared.CkViper, viper.New())

	tests := map[string]struct {
		context context.Context
	}{
		"Missing Viper config": {
			context: ctxNoViper,
		},
		"Missing app config": {
			context: ctxNoAppOptions,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {

			// When we run the help command
			rootCmd := RootCmd()
			rootCmd.SetArgs([]string{"help"})
			r.Panics(func() { r.NoError(rootCmd.ExecuteContext(data.context)) })
		})
	}
}

func Test_rootInitInvalidConfig(t *testing.T) {
	r := require.New(t)

	// Redirect our user home folder to an empty temp dir
	// to force the command to load default options
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// Create an invalid config file
	configFile := path.Join(tmpDir, ".rejig")
	fh, err := os.Create(configFile)
	r.NoError(err)
	_, err = fh.WriteString("templates: fubar")
	r.NoError(err)
	r.NoError(fh.Close())

	// When we run the help command
	rootCmd := RootCmd()
	actual := new(bytes.Buffer)
	rootCmd.SetOut(actual)
	rootCmd.SetErr(actual)

	rootCmd.SetArgs([]string{"help"})
	err = Execute(&rootCmd)

	r.Error(err)
}

func Test_executeConfigFileInvalidYAML(t *testing.T) {
	r := require.New(t)

	// Redirect our user home folder to a temp dir
	// to force the command to load default options
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// create an app config with invalid YAML
	configFile := path.Join(tmpDir, ".rejig")
	fh, err := os.Create(configFile)
	r.NoError(err)
	// _, err = fh.WriteString("templates: fubar")
	_, err = fh.WriteString("not valid yaml")
	r.NoError(err)
	r.NoError(fh.Close())

	// When we run the root command with a custom config file
	actual := new(bytes.Buffer)
	rootCmd := RootCmd()
	rootCmd.SetOut(actual)
	rootCmd.SetErr(actual)

	rootCmd.SetArgs([]string{"help"})
	err = Execute(&rootCmd)
	r.Error(err)
}

func Test_initViperDefaults(t *testing.T) {
	r := require.New(t)

	// Redirect our user home folder to an empty temp dir
	// to avoid conflicts with current users home folder
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// When we try to init a new config
	v := viper.New()
	r.NoError(initViper(v))
}

func Test_initViperMissingHome(t *testing.T) {
	r := require.New(t)

	// Remove the default users home folder from the environment
	// This case should trigger an immediate panic
	oldHome := setHome(t, "")
	defer restoreHome(t, oldHome)

	// When we try to init a new config
	// We expect the operation to panic
	v := viper.New()
	r.Panics(func() {
		r.NoError(initViper(v))
	})
}

func Test_initViperConfigFileNotExist(t *testing.T) {
	r := require.New(t)

	// Redirect our user home folder, which is where the application
	// will look for an app config file
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// When we attempt to initialize our app config the operation
	// should complete successfully (ie: a user defined config file
	// should be optional)
	v := viper.New()
	r.NoError(initViper(v))
}

func Test_initViperFileNoPermission(t *testing.T) {
	r := require.New(t)

	// Redirect our user home folder, which is where the application
	// will look for an app config file
	tmpDir, err := os.MkdirTemp("", "")
	r.NoError(err)
	defer os.RemoveAll(tmpDir)

	oldHome := setHome(t, tmpDir)
	defer restoreHome(t, oldHome)

	// Create a valid config file but make it read-only so the app
	// can read the contents
	configFile := path.Join(tmpDir, ".rejig")
	fh, err := os.Create(configFile)
	r.NoError(err)
	_, err = fh.WriteString("")
	r.NoError(err)
	r.NoError(fh.Chmod(0200))
	r.NoError(fh.Close())

	// When we attempt to initialize our app
	v := viper.New()
	err = initViper(v)

	// The operation should fail
	r.Error(err)

	// TODO: validate error message
	// TODO: Make sure error has a stack that includes our application
}

func Test_allColorThemesExist(t *testing.T) {
	r := require.New(t)

	numThemes := int64(ao.ThtDark) + 1

	for i := int64(0); i < numThemes; i++ {
		t.Run("Processing theme "+strconv.FormatInt(i, 10), func(t *testing.T) {
			curType := ao.ThemeType(i)
			r.Contains(CobraThemes, curType)
		})
	}
}
