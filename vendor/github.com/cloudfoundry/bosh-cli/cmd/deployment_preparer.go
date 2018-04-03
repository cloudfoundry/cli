package cmd

import (
	bihttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	bihttpclient "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/go-patch/patch"

	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-cli/deployment"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/installation"
	boshinst "github.com/cloudfoundry/bosh-cli/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

func NewDeploymentPreparer(
	ui biui.UI,
	logger boshlog.Logger,
	logTag string,
	deploymentStateService biconfig.DeploymentStateService,
	legacyDeploymentStateMigrator biconfig.LegacyDeploymentStateMigrator,
	releaseManager boshinst.ReleaseManager,
	deploymentRecord bidepl.Record,
	cloudFactory bicloud.Factory,
	stemcellManagerFactory bistemcell.ManagerFactory,
	agentClientFactory bihttpagent.AgentClientFactory,
	vmManagerFactory bivm.ManagerFactory,
	blobstoreFactory biblobstore.Factory,
	deployer bidepl.Deployer,
	deploymentManifestPath string,
	deploymentVars boshtpl.Variables,
	deploymentOp patch.Op,
	cpiInstaller bicpirel.CpiInstaller,
	releaseFetcher boshinst.ReleaseFetcher,
	stemcellFetcher bistemcell.Fetcher,
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser,
	deploymentManifestParser DeploymentManifestParser,
	tempRootConfigurator TempRootConfigurator,
	targetProvider biinstall.TargetProvider,
) DeploymentPreparer {
	return DeploymentPreparer{
		ui:                                      ui,
		logger:                                  logger,
		logTag:                                  logTag,
		deploymentStateService:                  deploymentStateService,
		legacyDeploymentStateMigrator:           legacyDeploymentStateMigrator,
		releaseManager:                          releaseManager,
		deploymentRecord:                        deploymentRecord,
		cloudFactory:                            cloudFactory,
		stemcellManagerFactory:                  stemcellManagerFactory,
		agentClientFactory:                      agentClientFactory,
		vmManagerFactory:                        vmManagerFactory,
		blobstoreFactory:                        blobstoreFactory,
		deployer:                                deployer,
		deploymentManifestPath:                  deploymentManifestPath,
		deploymentVars:                          deploymentVars,
		deploymentOp:                            deploymentOp,
		cpiInstaller:                            cpiInstaller,
		releaseFetcher:                          releaseFetcher,
		stemcellFetcher:                         stemcellFetcher,
		releaseSetAndInstallationManifestParser: releaseSetAndInstallationManifestParser,
		deploymentManifestParser:                deploymentManifestParser,
		tempRootConfigurator:                    tempRootConfigurator,
		targetProvider:                          targetProvider,
	}
}

type DeploymentPreparer struct {
	ui                                      biui.UI
	logger                                  boshlog.Logger
	logTag                                  string
	deploymentStateService                  biconfig.DeploymentStateService
	legacyDeploymentStateMigrator           biconfig.LegacyDeploymentStateMigrator
	releaseManager                          boshinst.ReleaseManager
	deploymentRecord                        bidepl.Record
	cloudFactory                            bicloud.Factory
	stemcellManagerFactory                  bistemcell.ManagerFactory
	agentClientFactory                      bihttpagent.AgentClientFactory
	vmManagerFactory                        bivm.ManagerFactory
	blobstoreFactory                        biblobstore.Factory
	deployer                                bidepl.Deployer
	deploymentManifestPath                  string
	deploymentVars                          boshtpl.Variables
	deploymentOp                            patch.Op
	cpiInstaller                            bicpirel.CpiInstaller
	releaseFetcher                          boshinst.ReleaseFetcher
	stemcellFetcher                         bistemcell.Fetcher
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser
	deploymentManifestParser                DeploymentManifestParser
	tempRootConfigurator                    TempRootConfigurator
	targetProvider                          biinstall.TargetProvider
}

func (c *DeploymentPreparer) PrepareDeployment(stage biui.Stage, recreate bool) (err error) {
	c.ui.BeginLinef("Deployment state: '%s'\n", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		migrated, err := c.legacyDeploymentStateMigrator.MigrateIfExists(biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		if err != nil {
			return bosherr.WrapError(err, "Migrating legacy deployment state file")
		}
		if migrated {
			c.ui.BeginLinef("Migrated legacy deployments file: '%s'\n", biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		}
	}

	deploymentState, err := c.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment state")
	}

	target, err := c.targetProvider.NewTarget()
	if err != nil {
		return bosherr.WrapError(err, "Determining installation target")
	}

	err = c.tempRootConfigurator.PrepareAndSetTempRoot(target.TmpPath(), c.logger)
	if err != nil {
		return bosherr.WrapError(err, "Setting temp root")
	}

	defer func() {
		err := c.releaseManager.DeleteAll()
		if err != nil {
			c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
		}
	}()

	var (
		extractedStemcell    bistemcell.ExtractedStemcell
		deploymentManifest   bideplmanifest.Manifest
		installationManifest biinstallmanifest.Manifest
		manifestSHA          string
	)
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		var releaseSetManifest birelsetmanifest.Manifest
		releaseSetManifest, installationManifest, err = c.releaseSetAndInstallationManifestParser.ReleaseSetAndInstallationManifest(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp)
		if err != nil {
			return err
		}

		for _, releaseRef := range releaseSetManifest.Releases {
			err = c.releaseFetcher.DownloadAndExtract(releaseRef, stage)
			if err != nil {
				return err
			}
		}

		err := c.cpiInstaller.ValidateCpiRelease(installationManifest, stage)
		if err != nil {
			return err
		}

		deploymentManifest, manifestSHA, err = c.deploymentManifestParser.GetDeploymentManifest(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp, releaseSetManifest, stage)
		if err != nil {
			return err
		}

		extractedStemcell, err = c.stemcellFetcher.GetStemcell(deploymentManifest, stage)
		return err
	})
	if err != nil {
		return err
	}
	defer func() {
		deleteErr := extractedStemcell.Cleanup()
		if deleteErr != nil {
			c.logger.Warn(c.logTag, "Failed to delete extracted stemcell: %s", deleteErr.Error())
		}
	}()

	isDeployed, err := c.deploymentRecord.IsDeployed(manifestSHA, c.releaseManager.List(), extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed && !recreate {
		c.ui.BeginLinef("No deployment, stemcell or release changes. Skipping deploy.\n")
		return nil
	}

	err = c.cpiInstaller.WithInstalledCpiRelease(installationManifest, target, stage, func(installation biinstall.Installation) error {
		return installation.WithRunningRegistry(c.logger, stage, func() error {
			return c.deploy(
				installation,
				deploymentState,
				extractedStemcell,
				installationManifest,
				deploymentManifest,
				manifestSHA,
				stage)
		})
	})

	return err

}

func (c *DeploymentPreparer) deploy(
	installation biinstall.Installation,
	deploymentState biconfig.DeploymentState,
	extractedStemcell bistemcell.ExtractedStemcell,
	installationManifest biinstallmanifest.Manifest,
	deploymentManifest bideplmanifest.Manifest,
	manifestSHA string,
	stage biui.Stage,
) (err error) {
	cloud, err := c.cloudFactory.NewCloud(installation, deploymentState.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)

	cloudStemcell, err := stemcellManager.Upload(extractedStemcell, stage)
	if err != nil {
		return err
	}

	agentClient, err := c.agentClientFactory.NewAgentClient(deploymentState.DirectorID, installationManifest.Mbus, installationManifest.Cert.CA)
	if err != nil {
		return err
	}
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient)

	blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus, bihttpclient.CreateDefaultClientInsecureSkipVerify())
	if err != nil {
		return bosherr.WrapError(err, "Creating blobstore client")
	}

	err = stage.PerformComplex("deploying", func(deployStage biui.Stage) error {
		err = c.deploymentRecord.Clear()
		if err != nil {
			return bosherr.WrapError(err, "Clearing deployment record")
		}

		_, err = c.deployer.Deploy(
			cloud,
			deploymentManifest,
			cloudStemcell,
			installationManifest.Registry,
			vmManager,
			blobstore,
			deployStage,
		)
		if err != nil {
			return bosherr.WrapError(err, "Deploying")
		}

		err = c.deploymentRecord.Update(manifestSHA, c.releaseManager.List())
		if err != nil {
			return bosherr.WrapError(err, "Updating deployment record")
		}

		return nil
	})
	if err != nil {
		return err
	}

	// TODO: cleanup unused disks here?

	err = stemcellManager.DeleteUnused(stage)
	if err != nil {
		return err
	}

	return nil
}
