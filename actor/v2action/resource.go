package v2action

import (
	"archive/zip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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

type FileChangedError struct {
	Filename string
}

func (e FileChangedError) Error() string {
	return fmt.Sprint("SHA1 mismatch for:", e.Filename)
}

type EmptyDirectoryError struct {
	Path string
}

func (e EmptyDirectoryError) Error() string {
	return fmt.Sprint(e.Path, "is empty")
}

type Resource ccv2.Resource

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
		if archivedFile.FileInfo().IsDir() {
			resource.Mode = DefaultFolderPermissions
		} else {
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

	walkErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if file ignored contine to the next file
		if gitIgnore.MatchesPath(path) {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		resource := Resource{
			Filename: filepath.ToSlash(relPath),
		}

		if info.IsDir() {
			resource.Mode = DefaultFolderPermissions
		} else {
			file, err := os.Open(path)
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
		return nil, EmptyDirectoryError{Path: sourceDir}
	}

	return resources, walkErr
}

// ResourceMatch returns a set of matched resources and unmatched resources in
// the order they were given in allResources.
func (actor Actor) ResourceMatch(allResources []Resource) ([]Resource, []Resource, Warnings, error) {
	resourcesToSend := [][]ccv2.Resource{{}}
	var currentList, sendCount int
	for _, resource := range allResources {
		if resource.Size == 0 {
			continue
		}

		resourcesToSend[currentList] = append(
			resourcesToSend[currentList],
			ccv2.Resource(resource),
		)
		sendCount += 1

		if len(resourcesToSend[currentList]) == MaxResourceMatchChunkSize {
			currentList += 1
			resourcesToSend = append(resourcesToSend, []ccv2.Resource{})
		}
	}

	log.WithFields(log.Fields{
		"total_resources":    len(allResources),
		"resources_to_match": sendCount,
		"chunks":             len(resourcesToSend),
	}).Debug("sending resource match stats")

	matchedCCResources := map[string]ccv2.Resource{}
	var allWarnings Warnings
	for _, chunk := range resourcesToSend {
		if len(chunk) == 0 {
			log.Debug("chunk size 0, stopping resource match requests")
			break
		}

		returnedResources, warnings, err := actor.CloudControllerClient.ResourceMatch(chunk)
		allWarnings = append(allWarnings, warnings...)

		if err != nil {
			log.Errorln("during resource matching", err)
			return nil, nil, allWarnings, err
		}

		for _, resource := range returnedResources {
			matchedCCResources[resource.SHA1] = resource
		}
	}
	log.WithField("matched_resource_count", len(matchedCCResources)).Debug("total number of matched resources")

	var matchedResources, unmatchedResources []Resource
	for _, resource := range allResources {
		if _, ok := matchedCCResources[resource.SHA1]; ok {
			matchedResources = append(matchedResources, resource)
		} else {
			unmatchedResources = append(unmatchedResources, resource)
		}
	}

	return matchedResources, unmatchedResources, allWarnings, nil
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
		reader, openErr := archiveFile.Open()
		if openErr != nil {
			log.WithField("archiveFile", archiveFile.Name).Errorln("opening path in dir:", openErr)
			return "", openErr
		}

		err = actor.addFileToZipFromFileSystem(
			resource.Filename, reader, archiveFile.FileInfo(),
			resource.Filename, resource.SHA1, resource.Mode, writer,
		)
		if err != nil {
			log.WithField("archiveFileName", archiveFile.Name).Errorln("zipping file:", err)
			return "", err
		}
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

		srcFile, err := os.Open(fullPath)
		if err != nil {
			log.WithField("fullPath", fullPath).Errorln("opening path in dir:", err)
			return "", err
		}

		fileInfo, err := srcFile.Stat()
		if err != nil {
			log.WithField("fullPath", fullPath).Errorln("stat error in dir:", err)
			return "", err
		}

		err = actor.addFileToZipFromFileSystem(
			fullPath, srcFile, fileInfo,
			resource.Filename, resource.SHA1, resource.Mode, writer,
		)
		if err != nil {
			log.WithField("fullPath", fullPath).Errorln("zipping file:", err)
			return "", err
		}
	}

	log.WithFields(log.Fields{
		"zip_file_location": zipFile.Name(),
		"zipped_file_count": len(filesToInclude),
	}).Info("zip file created")
	return zipFile.Name(), nil
}

func (Actor) actorToCCResources(resources []Resource) []ccv2.Resource {
	apiResources := make([]ccv2.Resource, 0, len(resources)) // Explicitly done to prevent nils

	for _, resource := range resources {
		apiResources = append(apiResources, ccv2.Resource(resource))
	}

	return apiResources
}

func (Actor) addFileToZipFromFileSystem(
	srcPath string, srcFile io.ReadCloser, fileInfo os.FileInfo,
	destPath string, sha1Sum string, mode os.FileMode, zipFile *zip.Writer,
) error {
	defer srcFile.Close()

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		log.WithField("srcPath", srcPath).Errorln("getting file info in dir:", err)
		return err
	}

	// An extra '/' indicates that this file is a directory
	if fileInfo.IsDir() && !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}

	header.Name = destPath
	header.Method = zip.Deflate

	header.SetMode(mode)
	log.WithFields(log.Fields{
		"srcPath":  srcPath,
		"destPath": destPath,
		"mode":     mode,
	}).Debug("setting mode for file")

	destFileWriter, err := zipFile.CreateHeader(header)
	if err != nil {
		log.Errorln("creating header:", err)
		return err
	}

	if !fileInfo.IsDir() {
		sum := sha1.New()

		multi := io.MultiWriter(sum, destFileWriter)
		if _, err := io.Copy(multi, srcFile); err != nil {
			log.WithField("srcPath", srcPath).Errorln("copying data in dir:", err)
			return err
		}

		if currentSum := fmt.Sprintf("%x", sum.Sum(nil)); sha1Sum != currentSum {
			log.WithFields(log.Fields{
				"expected":   sha1Sum,
				"currentSum": currentSum,
			}).Error("setting mode for file")
			return FileChangedError{Filename: srcPath}
		}
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
