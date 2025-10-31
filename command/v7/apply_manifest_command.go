package v7

import (
	"os"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/cf/errors"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

type ApplyManifestCommand struct {
	BaseCommand

	PathToManifest   flag.ManifestPathWithExistenceCheck `short:"f" description:"Path to app manifest"`
	Vars             []template.VarKV                    `long:"var" description:"Variable key value pair for variable substitution, (e.g., name=app1); can specify multiple times"`
	PathsToVarsFiles []flag.PathWithExistenceCheck       `long:"vars-file" description:"Path to a variable substitution file for manifest; can specify multiple times"`
	RedactEnv        bool                                `long:"redact-env" description:"Do not print values for environment vars set in the application manifest"`
	usage            interface{}                         `usage:"CF_NAME apply-manifest -f APP_MANIFEST_PATH"`
	relatedCommands  interface{}                         `related_commands:"create-app, create-app-manifest, push"`

	ManifestLocator ManifestLocator
	ManifestParser  ManifestParser

	DiffDisplayer DiffDisplayer
	CWD           string
}

func (cmd *ApplyManifestCommand) Setup(config command.Config, ui command.UI) error {
	cmd.ManifestLocator = manifestparser.NewLocator()
	cmd.ManifestParser = manifestparser.ManifestParser{}
	cmd.DiffDisplayer = &shared.ManifestDiffDisplayer{
		UI:        ui,
		RedactEnv: cmd.RedactEnv,
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd.CWD = currentDir

	return cmd.BaseCommand.Setup(config, ui)
}

func (cmd ApplyManifestCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	readPath := cmd.CWD
	if cmd.PathToManifest != "" {
		readPath = string(cmd.PathToManifest)
	}

	pathToManifest, exists, err := cmd.ManifestLocator.Path(readPath)
	if err != nil {
		return err
	}

	if !exists {
		return translatableerror.ManifestFileNotFoundInDirectoryError{PathToManifest: readPath}
	}

	cmd.UI.DisplayTextWithFlavor("Applying manifest {{.ManifestPath}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ManifestPath": pathToManifest,
		"OrgName":      cmd.Config.TargetedOrganization().Name,
		"SpaceName":    cmd.Config.TargetedSpace().Name,
		"Username":     user.Name,
	})

	spaceGUID := cmd.Config.TargetedSpace().GUID

	var pathsToVarsFiles []string
	for _, varFilePath := range cmd.PathsToVarsFiles {
		pathsToVarsFiles = append(pathsToVarsFiles, string(varFilePath))
	}

	interpolatedManifestBytes, err := cmd.ManifestParser.InterpolateManifest(pathToManifest, pathsToVarsFiles, cmd.Vars)
	if err != nil {
		return err
	}

	manifest, err := cmd.ManifestParser.ParseManifest(pathToManifest, interpolatedManifestBytes)
	if err != nil {
		if _, ok := err.(*yaml.TypeError); ok {
			return errors.New("Unable to apply manifest because its format is invalid.")
		}
		return err
	}

	manifestBytes, err := cmd.ManifestParser.MarshalManifest(manifest)
	if err != nil {
		return err
	}

	diff, warnings, err := cmd.Actor.DiffSpaceManifest(spaceGUID, manifestBytes)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, isUnexpectedError := err.(ccerror.V3UnexpectedResponseError); isUnexpectedError {
			cmd.UI.DisplayWarning("Unable to generate diff. Continuing to apply manifest...")
		} else {
			return err
		}
	} else {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Updating with these attributes...")

		err = cmd.DiffDisplayer.DisplayDiff(manifestBytes, diff)
		if err != nil {
			return err
		}
	}

	warnings, err = cmd.Actor.SetSpaceManifest(spaceGUID, manifestBytes)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayOK()

	return nil
}
