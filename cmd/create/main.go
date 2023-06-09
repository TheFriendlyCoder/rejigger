package create

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/TheFriendlyCoder/rejigger/cmd/shared"
	"github.com/TheFriendlyCoder/rejigger/lib"
	ao "github.com/TheFriendlyCoder/rejigger/lib/applicationOptions"
	e "github.com/TheFriendlyCoder/rejigger/lib/errors"
	"github.com/TheFriendlyCoder/rejigger/lib/templateManager"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// rootArgs parsed command line arguments
type rootArgs struct {
	// targetPath path to the folder where the new project is to be created
	targetPath string
	// templateName name of the template to use to create the new project from
	templateName string
}

// findTemplate looks up a specific template in the template inventory
// Name names must be of one of the forms:
// <template_name>
// <inventory_namespace>.<template_name>
func findTemplate(appOptions ao.AppOptions, name string) (ao.TemplateOptions, error) {
	parts := strings.Split(name, ".")
	if len(parts) > 2 {
		return ao.TemplateOptions{}, e.AOInvalidTemplateNameError()
	}
	var templates []ao.TemplateOptions
	var newName string
	if len(parts) == 1 {
		templates = appOptions.Templates
		newName = name
	} else {
		inv := appOptions.FindInventory(parts[0])
		if inv == nil {
			return ao.TemplateOptions{}, e.NewUnknownTemplateError(name)
		}
		// TODO: validate remote inventory definition (ie: check for duplication template names, etc.)

		// iterate through all inventory templates
		var err error
		templates, err = inv.GetTemplateDefinitions()
		if err != nil {
			// TODO: keep record of all failed inventory queries and report them
			//		 as an aggregate
			// TODO: ignore errors from template definitions if a match for the
			//		 template can be found elsewhere
			return ao.TemplateOptions{}, err
		}
		newName = parts[1]
	}

	for _, t := range templates {
		if t.GetName() == newName {
			return t, nil
		}
	}

	return ao.TemplateOptions{}, e.NewUnknownTemplateError(name)
}

// run Primary entry point function for our generator
func run(cmd *cobra.Command, args rootArgs) error {
	// We have to use cmd.OutOrStdout() to ensure output is redirected to Cobra
	// stream handler, to facilitate testing (ie: it allows us to capture output
	// during unit testing to validate results of CLI operations)
	lib.SNF(fmt.Fprintf(cmd.OutOrStdout(), "Loading template %s...\n", args.templateName))

	appOptions, ok := cmd.Context().Value(shared.CkOptions).(ao.AppOptions)
	if !ok {
		return e.NewInternalError("Failed to retrieve app options")
	}

	curTemplate, err := findTemplate(appOptions, args.templateName)
	if err != nil {
		return err
	}
	tm, err := templateManager.New(curTemplate)
	if err != nil {
		return err
	}

	if err = tm.GatherParams(cmd); err != nil {
		return err
	}
	lib.SNF(fmt.Fprintf(cmd.OutOrStdout(), "Generating project %s from template %s...\n", args.targetPath, curTemplate.GetName()))

	// Make sure our output folder exists
	if err = os.MkdirAll(args.targetPath, 0700); err != nil {
		return errors.WithStack(err)
	}

	return tm.Generate(args.targetPath)
	// TODO: after generating, put an archive file in the root folder summarizing what we did so we
	//		 can regenerate or update the project later
	// TODO: make terminology consistent (ie: config file for the app, manifest file for the template,
	//		 and something else for storing status of generated project - maybe audit file?
	//	file in home folder with user options: app options file / user options file
	//  file in root folder of template: manifest file / template manifest file
	//  file in the root folder of a template inventory: inventory file
	//  file generated in a project folder linking it to the original template: archive file
}

// validateArgs checks to see if the command line args provided to the app are valid
func validateArgs(options ao.AppOptions, args []string) error {
	if lib.DirExists(args[0]) {
		contents, err := os.ReadDir(args[0])
		if err != nil {
			log.Panic(err)
		}
		if len(contents) != 0 {
			return e.NewPathError(args[0], e.PePathNotEmpty)
		}
	}

	// Validate template name
	_, err := findTemplate(options, args[1])
	if err != nil {
		return err
	}

	return nil
}

// CreateCmd instantiates the "create" subcommand
func CreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   generateUsageLine(),
		Short: "create a new project from a template",
		Long:  `Creates a new project in an empty folder using content defined in a template`,
		Args:  cobra.MinimumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize our command context from the root command
			appOptions, ok := cmd.Context().Value(shared.CkOptions).(ao.AppOptions)
			if !ok {
				return e.CommandContextNotDefined()
			}
			return validateArgs(appOptions, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedArgs := rootArgs{
				targetPath:   args[0],
				templateName: args[1],
			}
			err := run(cmd, parsedArgs)
			if err != nil {
				// https://pkg.go.dev/github.com/pkg/errors#hdr-Retrieving_the_stack_trace_of_an_error_or_wrapper
				type stackTracer interface {
					StackTrace() errors.StackTrace
				}
				if temp, ok := interface{}(err).(stackTracer); ok {
					for _, f := range temp.StackTrace() {
						lib.SNF(fmt.Fprintf(cmd.ErrOrStderr(), "%+s:%d\n", f, f))
					}
				}
				lib.SNF(fmt.Fprintln(cmd.ErrOrStderr(), "Failed to generate project"))
				lib.SNF(fmt.Fprintln(cmd.ErrOrStderr(), err.Error()))
			}
			return err
		},
	}
}

// generateUsageLine dynamically generates a usage line for the app based on the contents
// of the rootArgs struct, using reflection
func generateUsageLine() string {
	var retval string
	retval = "create "
	temp := reflect.ValueOf(&rootArgs{}).Elem()

	for i := 0; i < temp.NumField(); i++ {
		varName := temp.Type().Field(i).Name
		retval += varName + " "
	}
	return strings.TrimSpace(retval)
}
