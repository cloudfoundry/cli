package sharedaction

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/ykk"
	ignore "github.com/sabhiram/go-gitignore"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
	MaxResourceMatchChunkSize     = 1000
)

var DefaultIgnoreLines = []string{
	".cfignore",
	".DS_Store",
	".git",
	".gitignore",
	".hg",
	".svn",
	"_darcs",
	"manifest.yaml",
	"manifest.yml",
}

type Resource struct {
	Filename string      `json:"fn"`
	Mode     os.FileMode `json:"mode"`
	SHA1     string      `json:"sha1"`
	Size     int64       `json:"size"`
}

// GatherArchiveResources returns a list of resources for an archive.
func (actor Actor) GatherArchiveResources(archivePath string) ([]Resource, error) {
	var resources []Resource

	archive, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	reader, err := actor.newArchiveReader(archive)
	if err != nil {
		return nil, err
	}

	gitIgnore, err := actor.generateArchiveCFIgnoreMatcher(reader.File)
	if err != nil {
		log.Errorln("reading .cfignore file:", err)
		return nil, err
	}

	for _, archivedFile := range reader.File {
		filename := filepath.ToSlash(archivedFile.Name)
		if gitIgnore.MatchesPath(filename) {
			continue
		}

		resource := Resource{Filename: filename}
		info := archivedFile.FileInfo()

		switch {
		case info.IsDir():
			resource.Mode = DefaultFolderPermissions
		case info.Mode()&os.ModeSymlink == os.ModeSymlink:
			resource.Mode = info.Mode()
		default:
			fileReader, err := archivedFile.Open()
			if err != nil {
				return nil, err
			}
			defer fileReader.Close()

			hash := sha1.New()

			_, err = io.Copy(hash, fileReader)
			if err != nil {
				return nil, err
			}

			resource.Mode = DefaultArchiveFilePermissions
			resource.SHA1 = fmt.Sprintf("%x", hash.Sum(nil))
			resource.Size = archivedFile.FileInfo().Size()
		}

		resources = append(resources, resource)
	}
	return resources, nil
}

// GatherDirectoryResources returns a list of resources for a directory.
func (actor Actor) GatherDirectoryResources(sourceDir string) ([]Resource, error) {
	var (
		resources []Resource
		gitIgnore *ignore.GitIgnore
	)

	gitIgnore, err := actor.generateDirectoryCFIgnoreMatcher(sourceDir)
	if err != nil {
		log.Errorln("reading .cfignore file:", err)
		return nil, err
	}

	evalDir, err := filepath.EvalSymlinks(sourceDir)
	if err != nil {
		log.Errorln("evaluating symlink:", err)
		return nil, err
	}

	walkErr := filepath.Walk(evalDir, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(evalDir, fullPath)
		if err != nil {
			return err
		}

		// if file ignored contine to the next file
		if gitIgnore.MatchesPath(relPath) {
			return nil
		}

		if relPath == "." {
			return nil
		}

		resource := Resource{
			Filename: filepath.ToSlash(relPath),
		}

		switch {
		case info.IsDir():
			// If the file is a directory
			resource.Mode = DefaultFolderPermissions
		case info.Mode()&os.ModeSymlink == os.ModeSymlink:
			// If the file is a Symlink we just set the mode of the file
			// We won't be using any sha information since we don't do
			// any resource matching on symlinks.
			resource.Mode = fixMode(info.Mode())
		default:
			// If the file is regular we want to open
			// and calculate the sha of the file
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			sum := sha1.New()
			_, err = io.Copy(sum, file)
			if err != nil {
				return err
			}

			resource.Mode = fixMode(info.Mode())
			resource.SHA1 = fmt.Sprintf("%x", sum.Sum(nil))
			resource.Size = info.Size()
		}

		resources = append(resources, resource)
		return nil
	})

	if len(resources) == 0 {
		return nil, actionerror.EmptyDirectoryError{Path: sourceDir}
	}

	return resources, walkErr
}

// ZipArchiveResources zips an archive and a sorted (based on full
// path/filename) list of resources and returns the location. On Windows, the
// filemode for user is forced to be readable and executable.
func (actor Actor) ZipArchiveResources(sourceArchivePath string, filesToInclude []Resource) (string, error) {
	log.WithField("sourceArchive", sourceArchivePath).Info("zipping source files from archive")
	zipFile, err := ioutil.TempFile("", "cf-cli-")
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	source, err := os.Open(sourceArchivePath)
	if err != nil {
		return "", err
	}
	defer source.Close()

	reader, err := actor.newArchiveReader(source)
	if err != nil {
		return "", err
	}

	for _, archiveFile := range reader.File {
		resource, ok := actor.findInResources(archiveFile.Name, filesToInclude)
		if !ok {
			log.WithField("archiveFileName", archiveFile.Name).Debug("skipping file")
			continue
		}

		log.WithField("archiveFileName", archiveFile.Name).Debug("zipping file")
		// archiveFile.Open opens the symlink file, not the file it points too
		reader, openErr := archiveFile.Open()
		if openErr != nil {
			log.WithField("archiveFile", archiveFile.Name).Errorln("opening path in dir:", openErr)
			return "", openErr
		}
		defer reader.Close()

		err = actor.addFileToZipFromFileSystem(
			resource.Filename, reader, archiveFile.FileInfo(),
			resource, writer,
		)
		if err != nil {
			log.WithField("archiveFileName", archiveFile.Name).Errorln("zipping file:", err)
			return "", err
		}
		reader.Close()
	}

	log.WithFields(log.Fields{
		"zip_file_location": zipFile.Name(),
		"zipped_file_count": len(filesToInclude),
	}).Info("zip file created")
	return zipFile.Name(), nil
}

// ZipDirectoryResources zips a directory and a sorted (based on full
// path/filename) list of resources and returns the location. On Windows, the
// filemode for user is forced to be readable and executable.
func (actor Actor) ZipDirectoryResources(sourceDir string, filesToInclude []Resource) (string, error) {
	log.WithField("sourceDir", sourceDir).Info("zipping source files from directory")
	zipFile, err := ioutil.TempFile("", "cf-cli-")
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	for _, resource := range filesToInclude {
		fullPath := filepath.Join(sourceDir, resource.Filename)
		log.WithField("fullPath", fullPath).Debug("zipping file")

		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			log.WithField("fullPath", fullPath).Errorln("stat error in dir:", err)
			return "", err
		}

		log.WithField("file-mode", fileInfo.Mode().String()).Debug("resource file info")
		if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			// we need to user os.Readlink to read a symlink file from a directory
			err = actor.addLinkToZipFromFileSystem(fullPath, fileInfo, resource, writer)
			if err != nil {
				log.WithField("fullPath", fullPath).Errorln("zipping file:", err)
				return "", err
			}
		} else {
			srcFile, err := os.Open(fullPath)
			defer srcFile.Close()
			if err != nil {
				log.WithField("fullPath", fullPath).Errorln("opening path in dir:", err)
				return "", err
			}

			err = actor.addFileToZipFromFileSystem(
				fullPath, srcFile, fileInfo,
				resource, writer,
			)
			srcFile.Close()
			if err != nil {
				log.WithField("fullPath", fullPath).Errorln("zipping file:", err)
				return "", err
			}
		}
	}

	log.WithFields(log.Fields{
		"zip_file_location": zipFile.Name(),
		"zipped_file_count": len(filesToInclude),
	}).Info("zip file created")
	return zipFile.Name(), nil
}

func (Actor) addLinkToZipFromFileSystem(srcPath string,
	fileInfo os.FileInfo, resource Resource,
	zipFile *zip.Writer,
) error {
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("getting file info in dir:", err)
		return err
	}

	header.Name = resource.Filename
	header.Method = zip.Deflate

	log.WithFields(log.Fields{
		"srcPath":  srcPath,
		"destPath": header.Name,
		"mode":     header.Mode().String(),
	}).Debug("setting mode for file")

	destFileWriter, err := zipFile.CreateHeader(header)
	if err != nil {
		log.Errorln("creating header:", err)
		return err
	}

	pathInSymlink, err := os.Readlink(srcPath)
	if err != nil {
		return err
	}
	log.WithField("path", pathInSymlink).Debug("resolving symlink")
	symLinkContents := strings.NewReader(pathInSymlink)
	if _, err := io.Copy(destFileWriter, symLinkContents); err != nil {
		log.WithField("srcPath", srcPath).Errorln("copying data in dir:", err)
		return err
	}

	return nil
}

func (Actor) addFileToZipFromFileSystem(srcPath string,
	srcFile io.Reader, fileInfo os.FileInfo, resource Resource,
	zipFile *zip.Writer,
) error {
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("getting file info in dir:", err)
		return err
	}

	header.Name = resource.Filename

	// An extra '/' indicates that this file is a directory
	if fileInfo.IsDir() && !strings.HasSuffix(resource.Filename, "/") {
		header.Name += "/"
	}
	header.Method = zip.Deflate
	header.SetMode(resource.Mode)

	log.WithFields(log.Fields{
		"srcPath":  srcPath,
		"destPath": header.Name,
		"mode":     header.Mode().String(),
	}).Debug("setting mode for file")

	destFileWriter, err := zipFile.CreateHeader(header)
	if err != nil {
		log.Errorln("creating header:", err)
		return err
	}

	if fileInfo.Mode().IsRegular() {
		sum := sha1.New()
		multi := io.MultiWriter(sum, destFileWriter)

		if _, err := io.Copy(multi, srcFile); err != nil {
			log.WithField("srcPath", srcPath).Errorln("copying data in dir:", err)
			return err
		}

		if currentSum := fmt.Sprintf("%x", sum.Sum(nil)); resource.SHA1 != currentSum {
			log.WithFields(log.Fields{
				"expected":   resource.SHA1,
				"currentSum": currentSum,
			}).Error("setting mode for file")
			return actionerror.FileChangedError{Filename: srcPath}
		}
	} else if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		io.Copy(destFileWriter, srcFile)
	}

	return nil
}

func (Actor) generateArchiveCFIgnoreMatcher(files []*zip.File) (*ignore.GitIgnore, error) {
	for _, item := range files {
		if strings.HasSuffix(item.Name, ".cfignore") {
			fileReader, err := item.Open()
			if err != nil {
				return nil, err
			}
			defer fileReader.Close()

			raw, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}
			s := append(DefaultIgnoreLines, strings.Split(string(raw), "\n")...)
			return ignore.CompileIgnoreLines(s...)
		}
	}
	return ignore.CompileIgnoreLines(DefaultIgnoreLines...)
}

func (actor Actor) generateDirectoryCFIgnoreMatcher(sourceDir string) (*ignore.GitIgnore, error) {
	pathToCFIgnore := filepath.Join(sourceDir, ".cfignore")

	additionalIgnoreLines := DefaultIgnoreLines

	// If verbose logging has files in the current dir, ignore them
	_, traceFiles := actor.Config.Verbose()
	for _, traceFilePath := range traceFiles {
		if relPath, err := filepath.Rel(sourceDir, traceFilePath); err == nil {
			additionalIgnoreLines = append(additionalIgnoreLines, relPath)
		}
	}

	if _, err := os.Stat(pathToCFIgnore); !os.IsNotExist(err) {
		return ignore.CompileIgnoreFileAndLines(pathToCFIgnore, additionalIgnoreLines...)
	} else {
		return ignore.CompileIgnoreLines(additionalIgnoreLines...)
	}
}

func (Actor) findInResources(path string, filesToInclude []Resource) (Resource, bool) {
	for _, resource := range filesToInclude {
		if resource.Filename == filepath.ToSlash(path) {
			log.WithField("resource", resource.Filename).Debug("found resource in files to include")
			return resource, true
		}
	}

	log.WithField("path", path).Debug("did not find resource in files to include")
	return Resource{}, false
}

func (Actor) newArchiveReader(archive *os.File) (*zip.Reader, error) {
	info, err := archive.Stat()
	if err != nil {
		return nil, err
	}

	return ykk.NewReader(archive, info.Size())
}
