package job

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	birelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type CompiledPackageRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

type DependencyCompiler interface {
	Compile([]bireljob.Job, biui.Stage) ([]CompiledPackageRef, error)
}

type dependencyCompiler struct {
	packageCompiler bistatepkg.Compiler

	logTag string
	logger boshlog.Logger
}

func NewDependencyCompiler(packageCompiler bistatepkg.Compiler, logger boshlog.Logger) DependencyCompiler {
	return &dependencyCompiler{
		packageCompiler: packageCompiler,

		logTag: "dependencyCompiler",
		logger: logger,
	}
}

// Compile resolves and compiles all transitive dependencies of multiple release jobs
func (c *dependencyCompiler) Compile(jobs []bireljob.Job, stage biui.Stage) ([]CompiledPackageRef, error) {
	compileOrderReleasePackages, err := c.resolveJobCompilationDependencies(jobs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving job package dependencies")
	}

	compiledPackageRefs, err := c.compilePackages(compileOrderReleasePackages, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies")
	}

	return compiledPackageRefs, nil
}

// resolveJobPackageCompilationDependencies returns all packages required by all specified jobs, in compilation order (reverse dependency order)
func (c *dependencyCompiler) resolveJobCompilationDependencies(jobs []bireljob.Job) ([]birelpkg.Compilable, error) {
	// collect and de-dupe all required packages (dependencies of jobs)
	packageMap := map[string]birelpkg.Compilable{}

	for _, releaseJob := range jobs {
		for _, releasePackage := range releaseJob.Packages {
			pkgKey := c.pkgKey(releasePackage)
			packageMap[pkgKey] = releasePackage
			c.resolvePackageDependencies(releasePackage, packageMap)
		}
	}

	// flatten map values to array
	packages := make([]birelpkg.Compilable, 0, len(packageMap))

	for _, releasePackage := range packageMap {
		packages = append(packages, releasePackage)
	}

	// sort in compilation order
	sortedPackages, err := birelpkg.Sort(packages)
	if err != nil {
		return nil, err
	}

	pkgs := []string{}

	for _, pkg := range sortedPackages {
		pkgs = append(pkgs, fmt.Sprintf("%s/%s", pkg.Name(), pkg.Fingerprint()))
	}

	c.logger.Debug(c.logTag, "Sorted dependencies:\n%s", strings.Join(pkgs, "\n"))

	return sortedPackages, nil
}

// resolvePackageDependencies adds the releasePackage's dependencies to the packageMap recursively
func (c *dependencyCompiler) resolvePackageDependencies(releasePackage birelpkg.Compilable, packageMap map[string]birelpkg.Compilable) {
	for _, dependency := range releasePackage.Deps() {
		// only add un-added packages, to avoid endless looping in case of cycles
		pkgKey := c.pkgKey(dependency)
		if _, found := packageMap[pkgKey]; !found {
			packageMap[pkgKey] = dependency
			c.resolvePackageDependencies(dependency, packageMap)
		}
	}
}

// compilePackages compiles the specified packages, in the order specified, uploads them to the Blobstore, and returns the blob references
func (c *dependencyCompiler) compilePackages(requiredPackages []birelpkg.Compilable, stage biui.Stage) ([]CompiledPackageRef, error) {
	packageRefs := make([]CompiledPackageRef, 0, len(requiredPackages))

	for _, pkg := range requiredPackages {
		stepName := fmt.Sprintf("Compiling package '%s/%s'", pkg.Name(), pkg.Fingerprint())

		err := stage.Perform(stepName, func() error {
			compiledPackageRecord, isAlreadyCompiled, err := c.packageCompiler.Compile(pkg)
			if err != nil {
				return err
			}

			packageRef := CompiledPackageRef{
				Name:        pkg.Name(),
				Version:     pkg.Fingerprint(),
				BlobstoreID: compiledPackageRecord.BlobID,
				SHA1:        compiledPackageRecord.BlobSHA1,
			}
			packageRefs = append(packageRefs, packageRef)

			if isAlreadyCompiled {
				return biui.NewSkipStageError(bosherr.Error(fmt.Sprintf("Package '%s' is already compiled. Skipped compilation", pkg.Name())), "Package already compiled")
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return packageRefs, nil
}

func (c *dependencyCompiler) pkgKey(pkg birelpkg.Compilable) string { return pkg.Name() }
