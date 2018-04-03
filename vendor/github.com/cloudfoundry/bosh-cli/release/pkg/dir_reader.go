package pkg

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"errors"

	. "github.com/cloudfoundry/bosh-cli/release/pkg/manifest"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

type DirReaderImpl struct {
	archiveFactory ArchiveFunc

	srcDirPath   string
	blobsDirPath string

	fs boshsys.FileSystem
}

var (
	fileNotFoundError = errors.New("File Not Found")
)

func NewDirReaderImpl(
	archiveFactory ArchiveFunc,
	srcDirPath string,
	blobsDirPath string,
	fs boshsys.FileSystem,
) DirReaderImpl {
	return DirReaderImpl{
		archiveFactory: archiveFactory,
		srcDirPath:     srcDirPath,
		blobsDirPath:   blobsDirPath,
		fs:             fs,
	}
}

func (r DirReaderImpl) Read(path string) (*Package, error) {
	manifestLock, err := r.collectLock(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting package spec lock")
	}

	if manifestLock != nil {
		resource := NewExistingResource(manifestLock.Name, manifestLock.Fingerprint, "")
		return NewPackage(resource, manifestLock.Dependencies), nil
	}

	manifest, files, prepFiles, err := r.collectFiles(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting package files")
	}

	// Note that files do not include package's spec file,
	// but rather specify dependencies as additional chunks for the fingerprint.
	archive := r.archiveFactory(ArchiveFactoryArgs{Files: files, PrepFiles: prepFiles, Chunks: manifest.Dependencies})

	fp, err := archive.Fingerprint()
	if err != nil {
		return nil, err
	}

	resource := NewResource(manifest.Name, fp, archive)

	return NewPackage(resource, manifest.Dependencies), nil
}

func (r DirReaderImpl) collectLock(path string) (*ManifestLock, error) {
	path = filepath.Join(path, "spec.lock")

	if r.fs.FileExists(path) {
		manifestLock, err := NewManifestLockFromPath(path, r.fs)
		if err != nil {
			return nil, err
		}

		return &manifestLock, nil
	}

	return nil, nil
}

func (r DirReaderImpl) collectFiles(path string) (Manifest, []File, []File, error) {
	var files, prepFiles []File

	specPath := filepath.Join(path, "spec")

	manifest, err := NewManifestFromPath(specPath, r.fs)
	if err != nil {
		return Manifest{}, nil, nil, err
	}

	packagingPath := filepath.Join(path, "packaging")
	files, err = r.checkAndFilterDir(packagingPath, path)
	if err != nil {
		if err == fileNotFoundError {
			return manifest, nil, nil, bosherr.Errorf(
				"Expected to find '%s' for package '%s'", packagingPath, manifest.Name)
		}

		return manifest, nil, nil, bosherr.Errorf("Unexpected error occurred: %s", err)
	}

	prePackagingPath := filepath.Join(path, "pre_packaging")
	prepFiles, err = r.checkAndFilterDir(prePackagingPath, path) //can proceed if there is no pre_packaging
	if err != nil && err != fileNotFoundError {
		return manifest, nil, nil, bosherr.Errorf("Unexpected error occurred: %s", err)
	}

	files = append(files, prepFiles...)

	filesByRelPath, err := r.applyFilesPattern(manifest)
	if err != nil {
		return manifest, nil, nil, err
	}

	excludedFiles, err := r.applyExcludedFilesPattern(manifest)
	if err != nil {
		return manifest, nil, nil, err
	}

	for _, excludedFile := range excludedFiles {
		delete(filesByRelPath, excludedFile.RelativePath)
	}

	for _, specialFileName := range []string{"packaging", "pre_packaging"} {
		if _, ok := filesByRelPath[specialFileName]; ok {
			errMsg := "Expected special '%s' file to not be included via 'files' key for package '%s'"
			return manifest, nil, nil, bosherr.Errorf(errMsg, specialFileName, manifest.Name)
		}
	}

	for _, file := range filesByRelPath {
		files = append(files, file)
	}

	return manifest, files, prepFiles, nil
}

func (r DirReaderImpl) applyFilesPattern(manifest Manifest) (map[string]File, error) {
	filesByRelPath := map[string]File{}

	for _, glob := range manifest.Files {
		matchingFilesFound := false

		srcDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.srcDirPath, glob))
		if err != nil {
			return map[string]File{}, bosherr.WrapErrorf(err, "Listing package files in src")
		}

		for _, path := range srcDirMatches {
			isPackageableFile, err := r.isPackageableFile(path)
			if err != nil {
				return map[string]File{}, bosherr.WrapErrorf(err, "Checking file packageability")
			}

			if isPackageableFile {
				matchingFilesFound = true
				file := NewFile(path, r.srcDirPath)
				if _, found := filesByRelPath[file.RelativePath]; !found {
					filesByRelPath[file.RelativePath] = file
				}
			}
		}

		blobsDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.blobsDirPath, glob))
		if err != nil {
			return map[string]File{}, bosherr.WrapErrorf(err, "Listing package files in blobs")
		}

		for _, path := range blobsDirMatches {
			isPackageableFile, err := r.isPackageableFile(path)
			if err != nil {
				return map[string]File{}, bosherr.WrapErrorf(err, "Checking file packageability")
			}

			if isPackageableFile {
				matchingFilesFound = true
				file := NewFile(path, r.blobsDirPath)
				if _, found := filesByRelPath[file.RelativePath]; !found {
					filesByRelPath[file.RelativePath] = file
				}
			}
		}

		if !matchingFilesFound {
			return nil, bosherr.Errorf("Missing files for pattern '%s'", glob)
		}
	}

	return filesByRelPath, nil
}

func (r DirReaderImpl) applyExcludedFilesPattern(manifest Manifest) ([]File, error) {
	var excludedFiles []File

	for _, glob := range manifest.ExcludedFiles {
		srcDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.srcDirPath, glob))
		if err != nil {
			return []File{}, bosherr.WrapErrorf(err, "Listing package excluded files in src")
		}

		for _, path := range srcDirMatches {
			file := NewFile(path, r.srcDirPath)
			excludedFiles = append(excludedFiles, file)
		}

		blobsDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.blobsDirPath, glob))
		if err != nil {
			return []File{}, bosherr.WrapErrorf(err, "Listing package excluded files in blobs")
		}

		for _, path := range blobsDirMatches {
			file := NewFile(path, r.blobsDirPath)
			excludedFiles = append(excludedFiles, file)
		}
	}

	return excludedFiles, nil
}

func (r DirReaderImpl) checkAndFilterDir(packagePath, path string) ([]File, error) {
	var files []File

	if r.fs.FileExists(packagePath) {
		isPackageableFile, err := r.isPackageableFile(packagePath)
		if err != nil {
			return nil, err
		}

		if isPackageableFile {
			file := NewFile(packagePath, path)
			file.ExcludeMode = true
			files = append(files, file)
		}
		return files, nil
	}

	return []File{}, fileNotFoundError
}

func (r DirReaderImpl) isPackageableFile(path string) (bool, error) {
	info, err := r.fs.Lstat(path)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSymlink != 0 || !info.IsDir(), nil
}
