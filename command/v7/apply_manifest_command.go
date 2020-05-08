package v7

import (
	"os"

	"code.cloudfoundry.org/cli/cf/errors"
	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

type ApplyManifestCommand struct {
	command.BaseCommand

	PathToManifest   flag.ManifestPathWithExistenceCheck `short:"f" description:"Path to app manifest"`
	Vars             []template.VarKV                    `long:"var" description:"Variable key value pair for variable substitution, (e.g., name=app1); can specify multiple times"`
	PathsToVarsFiles []flag.PathWithExistenceCheck       `long:"vars-file" description:"Path to a variable substitution file for manifest; can specify multiple times"`
	usage            interface{}                         `usage:"CF_NAME apply-manifest -f APP_MANIFEST_PATH"`
	relatedCommands  interface{}                         `related_commands:"create-app, create-app-manifest, push"`

	ManifestLocator ManifestLocator
	ManifestParser  ManifestParser
	CWD             string
}

func (cmd *ApplyManifestCommand) Setup(config command.Config, ui command.UI) error {
	cmd.ManifestLocator = manifestparser.NewLocator()
	cmd.ManifestParser = manifestparser.ManifestParser{}

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

	user, err := cmd.Config.CurrentUser()
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

	var pathsToVarsFiles []string
	for _, varFilePath := range cmd.PathsToVarsFiles {
		pathsToVarsFiles = append(pathsToVarsFiles, string(varFilePath))
	}

	manifest, err := cmd.ManifestParser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, cmd.Vars)
	if err != nil {
		if _, ok := err.(*yaml.TypeError); ok {
			return errors.New("Unable to apply manifest because its format is invalid.")
		}
		return err
	}

	rawManifest, err := cmd.ManifestParser.MarshalManifest(manifest)
	if err != nil {
		return err
	}

	warnings, err := cmd.Actor.SetSpaceManifest(cmd.Config.TargetedSpace().GUID, rawManifest)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}
