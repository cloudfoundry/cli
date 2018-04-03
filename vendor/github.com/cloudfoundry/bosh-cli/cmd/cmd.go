package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cppforlife/go-patch/patch"

	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	"github.com/cloudfoundry/bosh-cli/crypto"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshuit "github.com/cloudfoundry/bosh-cli/ui/task"

	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
)

type Cmd struct {
	BoshOpts BoshOpts
	Opts     interface{}

	deps BasicDeps
}

func NewCmd(boshOpts BoshOpts, opts interface{}, deps BasicDeps) Cmd {
	return Cmd{boshOpts, opts, deps}
}

type cmdConveniencePanic struct {
	Err error
}

func (c Cmd) Execute() (cmdErr error) {
	// Catch convenience panics from panicIfErr
	defer func() {
		if r := recover(); r != nil {
			if cp, ok := r.(cmdConveniencePanic); ok {
				cmdErr = cp.Err
			} else {
				panic(r)
			}
		}
	}()

	c.configureUI()
	c.configureFS()

	if c.BoshOpts.Sha2 {
		c.deps = c.deps.WithSha2CheckSumming()
	}

	deps := c.deps

	switch opts := c.Opts.(type) {
	case *EnvironmentOpts:
		return NewEnvironmentCmd(deps.UI, c.director()).Run()

	case *EnvironmentsOpts:
		return NewEnvironmentsCmd(c.config(), deps.UI).Run()

	case *CreateEnvOpts:
		envProvider := func(manifestPath string, statePath string, vars boshtpl.Variables, op patch.Op) DeploymentPreparer {
			return NewEnvFactory(deps, manifestPath, statePath, vars, op).Preparer()
		}

		stage := boshui.NewStage(deps.UI, deps.Time, deps.Logger)
		return NewCreateEnvCmd(deps.UI, envProvider).Run(stage, *opts)

	case *DeleteEnvOpts:
		envProvider := func(manifestPath string, statePath string, vars boshtpl.Variables, op patch.Op) DeploymentDeleter {
			return NewEnvFactory(deps, manifestPath, statePath, vars, op).Deleter()
		}

		stage := boshui.NewStage(deps.UI, deps.Time, deps.Logger)
		return NewDeleteCmd(deps.UI, envProvider).Run(stage, *opts)

	case *AliasEnvOpts:
		sessionFactory := func(config cmdconf.Config) Session {
			return NewSessionFromOpts(c.BoshOpts, config, deps.UI, true, false, deps.FS, deps.Logger)
		}

		return NewAliasEnvCmd(sessionFactory, c.config(), deps.UI).Run(*opts)

	case *LogInOpts:
		sessionFactory := func(config cmdconf.Config) Session {
			return NewSessionFromOpts(c.BoshOpts, config, deps.UI, true, true, deps.FS, deps.Logger)
		}

		config := c.config()
		basicStrategy := NewBasicLoginStrategy(sessionFactory, config, deps.UI)
		uaaStrategy := NewUAALoginStrategy(sessionFactory, config, deps.UI, deps.Logger)

		sess := NewSessionFromOpts(c.BoshOpts, c.config(), deps.UI, true, true, deps.FS, deps.Logger)

		anonDirector, err := sess.AnonymousDirector()
		if err != nil {
			return err
		}

		return NewLogInCmd(basicStrategy, uaaStrategy, anonDirector).Run()

	case *LogOutOpts:
		config := c.config()
		sess := NewSessionFromOpts(c.BoshOpts, config, deps.UI, true, true, deps.FS, deps.Logger)
		return NewLogOutCmd(sess.Environment(), config, deps.UI).Run()

	case *TaskOpts:
		eventsTaskReporter := boshuit.NewReporter(deps.UI, true)
		plainTaskReporter := boshuit.NewReporter(deps.UI, false)
		return NewTaskCmd(eventsTaskReporter, plainTaskReporter, c.director()).Run(*opts)

	case *TasksOpts:
		return NewTasksCmd(deps.UI, c.director()).Run(*opts)

	case *CancelTaskOpts:
		return NewCancelTaskCmd(c.director()).Run(*opts)

	case *DeploymentOpts:
		sessionFactory := func(config cmdconf.Config) Session {
			return NewSessionFromOpts(c.BoshOpts, config, deps.UI, true, false, deps.FS, deps.Logger)
		}

		return NewDeploymentCmd(sessionFactory, c.config(), deps.UI).Run()

	case *DeploymentsOpts:
		return NewDeploymentsCmd(deps.UI, c.director()).Run()

	case *DeleteDeploymentOpts:
		return NewDeleteDeploymentCmd(deps.UI, c.deployment()).Run(*opts)

	case *ReleasesOpts:
		return NewReleasesCmd(deps.UI, c.director()).Run()

	case *UploadReleaseOpts:
		relProv, relDirProv := c.releaseProviders()

		releaseDirFactory := func(dir DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
			releaseReader := relDirProv.NewReleaseReader(dir.Path, c.BoshOpts.Parallel)
			releaseDir := relDirProv.NewFSReleaseDir(dir.Path, c.BoshOpts.Parallel)
			return releaseReader, releaseDir
		}

		releaseWriter := relProv.NewArchiveWriter()

		releaseArchiveFactory := func(path string) boshdir.ReleaseArchive {
			return boshdir.NewFSReleaseArchive(path, deps.FS)
		}

		cmd := NewUploadReleaseCmd(
			releaseDirFactory,
			releaseWriter,
			c.director(),
			releaseArchiveFactory,
			deps.CmdRunner,
			deps.FS,
			deps.UI,
		)

		return cmd.Run(*opts)

	case *DeleteReleaseOpts:
		return NewDeleteReleaseCmd(deps.UI, c.director()).Run(*opts)

	case *StemcellsOpts:
		return NewStemcellsCmd(deps.UI, c.director()).Run()

	case *UploadStemcellOpts:
		stemcellArchiveFactory := func(path string) boshdir.StemcellArchive {
			return boshdir.NewFSStemcellArchive(path, deps.FS)
		}

		return NewUploadStemcellCmd(c.director(), stemcellArchiveFactory, deps.UI).Run(*opts)

	case *DeleteStemcellOpts:
		return NewDeleteStemcellCmd(deps.UI, c.director()).Run(*opts)

	case *RepackStemcellOpts:
		stemcellReader := bistemcell.NewReader(deps.Compressor, deps.FS)
		stemcellExtractor := bistemcell.NewExtractor(stemcellReader, deps.FS)

		return NewRepackStemcellCmd(deps.UI, deps.FS, stemcellExtractor).Run(*opts)

	case *LocksOpts:
		return NewLocksCmd(deps.UI, c.director()).Run()

	case *ErrandsOpts:
		return NewErrandsCmd(deps.UI, c.deployment()).Run()

	case *RunErrandOpts:
		director, deployment := c.directorAndDeployment()
		downloader := NewUIDownloader(director, deps.Time, deps.FS, deps.UI)
		return NewRunErrandCmd(deployment, downloader, deps.UI).Run(*opts)

	case *AttachDiskOpts:
		return NewAttachDiskCmd(c.deployment()).Run(*opts)

	case *DisksOpts:
		return NewDisksCmd(deps.UI, c.director()).Run(*opts)

	case *DeleteDiskOpts:
		return NewDeleteDiskCmd(deps.UI, c.director()).Run(*opts)

	case *OrphanDiskOpts:
		return NewOrphanDiskCmd(deps.UI, c.director()).Run(*opts)

	case *SnapshotsOpts:
		return NewSnapshotsCmd(deps.UI, c.deployment()).Run(*opts)

	case *TakeSnapshotOpts:
		return NewTakeSnapshotCmd(c.deployment()).Run(*opts)

	case *DeleteSnapshotOpts:
		return NewDeleteSnapshotCmd(deps.UI, c.deployment()).Run(*opts)

	case *DeleteSnapshotsOpts:
		return NewDeleteSnapshotsCmd(deps.UI, c.deployment()).Run()

	case *DeleteVMOpts:
		return NewDeleteVMCmd(deps.UI, c.deployment()).Run(*opts)

	case *InterpolateOpts:
		return NewInterpolateCmd(deps.UI).Run(*opts)

	case *ConfigOpts:
		return NewConfigCmd(deps.UI, c.director()).Run(*opts)

	case *ConfigsOpts:
		return NewConfigsCmd(deps.UI, c.director()).Run(*opts)

	case *DiffConfigOpts:
		return NewDiffConfigCmd(deps.UI, c.director()).Run(*opts)

	case *UpdateConfigOpts:
		return NewUpdateConfigCmd(deps.UI, c.director()).Run(*opts)

	case *DeleteConfigOpts:
		return NewDeleteConfigCmd(deps.UI, c.director()).Run(*opts)

	case *CloudConfigOpts:
		return NewCloudConfigCmd(deps.UI, c.director()).Run()

	case *UpdateCloudConfigOpts:
		return NewUpdateCloudConfigCmd(deps.UI, c.director()).Run(*opts)

	case *CPIConfigOpts:
		return NewCPIConfigCmd(deps.UI, c.director()).Run()

	case *UpdateCPIConfigOpts:
		return NewUpdateCPIConfigCmd(deps.UI, c.director()).Run(*opts)

	case *RuntimeConfigOpts:
		return NewRuntimeConfigCmd(deps.UI, c.director()).Run(*opts)

	case *UpdateRuntimeConfigOpts:
		director := c.director()
		releaseManager := c.releaseManager(director)
		return NewUpdateRuntimeConfigCmd(deps.UI, director, releaseManager).Run(*opts)

	case *ManifestOpts:
		return NewManifestCmd(deps.UI, c.deployment()).Run()

	case *EventsOpts:
		return NewEventsCmd(deps.UI, c.director()).Run(*opts)

	case *EventOpts:
		return NewEventCmd(deps.UI, c.director()).Run(*opts)

	case *InspectReleaseOpts:
		return NewInspectReleaseCmd(deps.UI, c.director()).Run(*opts)

	case *VMsOpts:
		return NewVMsCmd(deps.UI, c.director(), c.BoshOpts.Parallel).Run(*opts)

	case *InstancesOpts:
		return NewInstancesCmd(deps.UI, c.director(), c.BoshOpts.Parallel).Run(*opts)

	case *UpdateResurrectionOpts:
		return NewUpdateResurrectionCmd(c.director()).Run(*opts)

	case *IgnoreOpts:
		return NewIgnoreCmd(c.deployment()).Run(*opts)

	case *UnignoreOpts:
		return NewUnignoreCmd(c.deployment()).Run(*opts)

	case *DeployOpts:
		director, deployment := c.directorAndDeployment()
		releaseManager := c.releaseManager(director)
		return NewDeployCmd(deps.UI, deployment, releaseManager).Run(*opts)

	case *StartOpts:
		return NewStartCmd(deps.UI, c.deployment()).Run(*opts)

	case *StopOpts:
		return NewStopCmd(deps.UI, c.deployment()).Run(*opts)

	case *RestartOpts:
		return NewRestartCmd(deps.UI, c.deployment()).Run(*opts)

	case *RecreateOpts:
		return NewRecreateCmd(deps.UI, c.deployment()).Run(*opts)

	case *CloudCheckOpts:
		return NewCloudCheckCmd(c.deployment(), deps.UI).Run(*opts)

	case *CleanUpOpts:
		return NewCleanUpCmd(deps.UI, c.director()).Run(*opts)

	case *LogsOpts:
		director, deployment := c.directorAndDeployment()
		downloader := NewUIDownloader(director, deps.Time, deps.FS, deps.UI)
		sshProvider := boshssh.NewProvider(deps.CmdRunner, deps.FS, deps.UI, deps.Logger)
		nonIntSSHRunner := sshProvider.NewSSHRunner(false)
		return NewLogsCmd(deployment, downloader, deps.UUIDGen, nonIntSSHRunner).Run(*opts)

	case *SSHOpts:
		sshProvider := boshssh.NewProvider(deps.CmdRunner, deps.FS, deps.UI, deps.Logger)
		intSSHRunner := sshProvider.NewSSHRunner(true)
		nonIntSSHRunner := sshProvider.NewSSHRunner(false)
		resultsSSHRunner := sshProvider.NewResultsSSHRunner(false)
		return NewSSHCmd(c.deployment(), deps.UUIDGen, intSSHRunner, nonIntSSHRunner, resultsSSHRunner, deps.UI).Run(*opts)

	case *SCPOpts:
		sshProvider := boshssh.NewProvider(deps.CmdRunner, deps.FS, deps.UI, deps.Logger)
		scpRunner := sshProvider.NewSCPRunner()
		return NewSCPCmd(c.deployment(), deps.UUIDGen, scpRunner, deps.UI).Run(*opts)

	case *ExportReleaseOpts:
		director, deployment := c.directorAndDeployment()
		downloader := NewUIDownloader(director, deps.Time, deps.FS, deps.UI)
		return NewExportReleaseCmd(deployment, downloader).Run(*opts)

	case *InitReleaseOpts:
		return NewInitReleaseCmd(c.releaseDir(opts.Directory)).Run(*opts)

	case *ResetReleaseOpts:
		return NewResetReleaseCmd(c.releaseDir(opts.Directory)).Run(*opts)

	case *GenerateJobOpts:
		return NewGenerateJobCmd(c.releaseDir(opts.Directory)).Run(*opts)

	case *GeneratePackageOpts:
		return NewGeneratePackageCmd(c.releaseDir(opts.Directory)).Run(*opts)

	case *VendorPackageOpts:
		return NewVendorPackageCmd(c.releaseDir, deps.UI).Run(*opts)

	case *FinalizeReleaseOpts:
		_, relDirProv := c.releaseProviders()
		releaseReader := relDirProv.NewReleaseReader(opts.Directory.Path, c.BoshOpts.Parallel)
		releaseDir := relDirProv.NewFSReleaseDir(opts.Directory.Path, c.BoshOpts.Parallel)
		return NewFinalizeReleaseCmd(releaseReader, releaseDir, deps.UI).Run(*opts)

	case *CreateReleaseOpts:
		relProv, relDirProv := c.releaseProviders()

		releaseDirFactory := func(dir DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
			releaseReader := relDirProv.NewReleaseReader(dir.Path, c.BoshOpts.Parallel)
			releaseDir := relDirProv.NewFSReleaseDir(dir.Path, c.BoshOpts.Parallel)
			return releaseReader, releaseDir
		}

		_, err := NewCreateReleaseCmd(
			releaseDirFactory,
			relProv.NewArchiveWriter(),
			c.deps.FS,
			c.deps.UI,
		).Run(*opts)
		return err

	case *Sha1ifyReleaseOpts:
		relProv, _ := c.releaseProviders()

		return NewRedigestReleaseCmd(
			relProv.NewArchiveReader(),
			relProv.NewArchiveWriter(),
			crypto.NewDigestCalculator(c.deps.FS, []boshcrypto.Algorithm{boshcrypto.DigestAlgorithmSHA1}),
			boshfu.NewFileMover(c.deps.FS),
			c.deps.FS,
			c.deps.UI,
		).Run(opts.Args)

	case *Sha2ifyReleaseOpts:
		relProv, _ := c.releaseProviders()

		return NewRedigestReleaseCmd(
			relProv.NewArchiveReader(),
			relProv.NewArchiveWriter(),
			crypto.NewDigestCalculator(c.deps.FS, []boshcrypto.Algorithm{boshcrypto.DigestAlgorithmSHA256}),
			boshfu.NewFileMover(c.deps.FS),
			c.deps.FS,
			c.deps.UI,
		).Run(opts.Args)

	case *BlobsOpts:
		return NewBlobsCmd(c.blobsDir(opts.Directory), deps.UI).Run()

	case *AddBlobOpts:
		return NewAddBlobCmd(c.blobsDir(opts.Directory), deps.FS, deps.UI).Run(*opts)

	case *RemoveBlobOpts:
		return NewRemoveBlobCmd(c.blobsDir(opts.Directory), deps.UI).Run(*opts)

	case *UploadBlobsOpts:
		return NewUploadBlobsCmd(c.blobsDir(opts.Directory)).Run()

	case *SyncBlobsOpts:
		return NewSyncBlobsCmd(c.blobsDir(opts.Directory), c.BoshOpts.Parallel).Run()

	case *MessageOpts:
		deps.UI.PrintBlock([]byte(opts.Message))
		return nil

	case *VariablesOpts:
		return NewVariablesCmd(deps.UI, c.deployment()).Run()

	default:
		return fmt.Errorf("Unhandled command: %#v", c.Opts)
	}
}
func (c Cmd) configureUI() {
	c.deps.UI.EnableTTY(c.BoshOpts.TTYOpt)

	if !c.BoshOpts.NoColorOpt {
		c.deps.UI.EnableColor()
	}

	if c.BoshOpts.JSONOpt {
		c.deps.UI.EnableJSON()
	}

	if c.BoshOpts.NonInteractiveOpt {
		c.deps.UI.EnableNonInteractive()
	}

	if len(c.BoshOpts.ColumnOpt) > 0 {
		headers := []boshtbl.Header{}
		for _, columnOpt := range c.BoshOpts.ColumnOpt {
			headers = append(headers, columnOpt.Header)
		}

		c.deps.UI.ShowColumns(headers)
	}
}

func (c Cmd) configureFS() {
	tmpDirPath, err := c.deps.FS.ExpandPath(filepath.Join("~", ".bosh", "tmp"))
	c.panicIfErr(err)

	err = c.deps.FS.ChangeTempRoot(tmpDirPath)
	c.panicIfErr(err)
}

func (c Cmd) config() cmdconf.Config {
	config, err := cmdconf.NewFSConfigFromPath(c.BoshOpts.ConfigPathOpt, c.deps.FS)
	c.panicIfErr(err)

	return config
}

func (c Cmd) session() Session {
	return NewSessionFromOpts(c.BoshOpts, c.config(), c.deps.UI, true, true, c.deps.FS, c.deps.Logger)
}

func (c Cmd) director() boshdir.Director {
	director, err := c.session().Director()
	c.panicIfErr(err)

	return director
}

func (c Cmd) deployment() boshdir.Deployment {
	deployment, err := c.session().Deployment()
	c.panicIfErr(err)

	return deployment
}

func (c Cmd) directorAndDeployment() (boshdir.Director, boshdir.Deployment) {
	sess := c.session()

	director, err := sess.Director()
	c.panicIfErr(err)

	deployment, err := sess.Deployment()
	c.panicIfErr(err)

	return director, deployment
}

func (c Cmd) releaseProviders() (boshrel.Provider, boshreldir.Provider) {
	indexReporter := boshui.NewIndexReporter(c.deps.UI)
	blobsReporter := boshui.NewBlobsReporter(c.deps.UI)
	releaseIndexReporter := boshui.NewReleaseIndexReporter(c.deps.UI)

	releaseProvider := boshrel.NewProvider(
		c.deps.CmdRunner, c.deps.Compressor, c.deps.DigestCalculator, c.deps.FS, c.deps.Logger)

	releaseDirProvider := boshreldir.NewProvider(
		indexReporter, releaseIndexReporter, blobsReporter, releaseProvider,
		c.deps.DigestCalculator, c.deps.CmdRunner, c.deps.UUIDGen, c.deps.Time, c.deps.FS, c.deps.DigestCreationAlgorithms, c.deps.Logger)

	return releaseProvider, releaseDirProvider
}

func (c Cmd) releaseManager(director boshdir.Director) ReleaseManager {
	relProv, relDirProv := c.releaseProviders()

	releaseDirFactory := func(dir DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
		releaseReader := relDirProv.NewReleaseReader(dir.Path, c.BoshOpts.Parallel)
		releaseDir := relDirProv.NewFSReleaseDir(dir.Path, c.BoshOpts.Parallel)
		return releaseReader, releaseDir
	}

	releaseWriter := relProv.NewArchiveWriter()

	createReleaseCmd := NewCreateReleaseCmd(
		releaseDirFactory,
		releaseWriter,
		c.deps.FS,
		c.deps.UI,
	)

	releaseArchiveFactory := func(path string) boshdir.ReleaseArchive {
		return boshdir.NewFSReleaseArchive(path, c.deps.FS)
	}

	uploadReleaseCmd := NewUploadReleaseCmd(
		releaseDirFactory,
		releaseWriter,
		director,
		releaseArchiveFactory,
		c.deps.CmdRunner,
		c.deps.FS,
		c.deps.UI,
	)

	return NewReleaseManager(createReleaseCmd, uploadReleaseCmd, c.BoshOpts.Parallel)
}

func (c Cmd) blobsDir(dir DirOrCWDArg) boshreldir.BlobsDir {
	_, relDirProv := c.releaseProviders()
	return relDirProv.NewFSBlobsDir(dir.Path)
}

func (c Cmd) releaseDir(dir DirOrCWDArg) boshreldir.ReleaseDir {
	_, relDirProv := c.releaseProviders()
	return relDirProv.NewFSReleaseDir(dir.Path, c.BoshOpts.Parallel)
}

func (c Cmd) panicIfErr(err error) {
	if err != nil {
		panic(cmdConveniencePanic{err})
	}
}
