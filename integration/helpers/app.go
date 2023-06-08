package helpers

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/onsi/gomega/gbytes"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v2"
)

// CreateApp creates an empty app in CloudController with no package or droplet
func CreateApp(app string) {
	Eventually(CF("create-app", app)).Should(Exit(0))
}

// QuickDeleteApp deletes the app with the given name, if provided, using
// 'cf curl /v3/app... -X DELETE'.
func QuickDeleteApp(appName string) {
	guid := AppGUID(appName)
	url := fmt.Sprintf("/v3/apps/%s", guid)
	session := CF("curl", "-X", "DELETE", url)
	Eventually(session).Should(Exit(0))
}

// WithHelloWorldApp creates a simple application to use with your CLI command
// (typically CF Push). When pushing, be aware of specifying '-b
// staticfile_buildpack" so that your app will correctly start up with the
// proper buildpack.
func WithHelloWorldApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-app")
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// WithMultiEndpointApp creates a simple application to use with your CLI command
// (typically CF Push). It has multiple endpoints which are helpful when testing
// http healthchecks.
func WithMultiEndpointApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-app")
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
	Expect(err).ToNot(HaveOccurred())

	tempfile = filepath.Join(dir, "other_endpoint.html")
	err = ioutil.WriteFile(tempfile, []byte("other endpoint"), 0666)
	Expect(err).ToNot(HaveOccurred())

	tempfile = filepath.Join(dir, "third_endpoint.html")
	err = ioutil.WriteFile(tempfile, []byte("third endpoint"), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

func WithSidecarApp(f func(dir string), appName string) {
	withSidecarManifest := func(dir string) {
		err := ioutil.WriteFile(filepath.Join(dir, "manifest.yml"), []byte(fmt.Sprintf(`---
applications:
  - name: %s
    sidecars:
    - name: sidecar_name
      process_types: [web]
      command: sleep infinity`,
			appName)), 0666)
		Expect(err).ToNot(HaveOccurred())

		f(dir)
	}

	WithHelloWorldApp(withSidecarManifest)
}

func WithTaskApp(f func(dir string), appName string) {
	withTaskManifest := func(dir string) {
		err := ioutil.WriteFile(filepath.Join(dir, "manifest.yml"), []byte(fmt.Sprintf(`---
applications:
- name: %s
  processes:
  - type: task
    command: echo hi`,
			appName)), 0666)
		Expect(err).ToNot(HaveOccurred())

		f(dir)
	}

	WithHelloWorldApp(withTaskManifest)
}

// WithNoResourceMatchedApp creates a simple application to use with your CLI
// command (typically CF Push). When pushing, be aware of specifying '-b
// staticfile_buildpack" so that your app will correctly start up with the
// proper buildpack.
func WithNoResourceMatchedApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-app")
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")

	err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %s", strings.Repeat("a", 65*1024))), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// WithMultiBuildpackApp creates a multi-buildpack application to use with the CF push command.
func WithMultiBuildpackApp(f func(dir string)) {
	f("../../assets/go_calls_ruby")
}

// WithProcfileApp creates an application to use with your CLI command
// that contains Procfile defining web and worker processes.
func WithProcfileApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-ruby-app")
	defer os.RemoveAll(dir)

	err := ioutil.WriteFile(filepath.Join(dir, "Procfile"), []byte(`---
web: ruby -run -e httpd . -p $PORT
console: bundle exec irb`,
	), 0666)
	Expect(err).ToNot(HaveOccurred())

	//TODO: Remove the ruby version once bundler issue(https://github.com/rubygems/rubygems/issues/6280)
	// is solved
	err = ioutil.WriteFile(filepath.Join(dir, "Gemfile"), []byte(`source 'http://rubygems.org'
ruby '~> 3.0'
gem 'irb'
gem 'webrick'`,
	), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Gemfile.lock"), []byte(`
GEM
  remote: http://rubygems.org/
  specs:
    io-console (0.6.0)
    irb (1.6.4)
      reline (>= 0.3.0)
    reline (0.3.3)
      io-console (~> 0.5)
    webrick (1.8.1)

PLATFORMS
  ruby

DEPENDENCIES
  irb
  webrick
`,
	), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// WithCrashingApp creates an application to use with your CLI command
// that will not successfully start its `web` process
func WithCrashingApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "crashing-ruby-app")
	defer os.RemoveAll(dir)

	err := ioutil.WriteFile(filepath.Join(dir, "Procfile"), []byte(`---
web: bogus bogus`,
	), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Gemfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Gemfile.lock"), []byte(`
GEM
  specs:

PLATFORMS
  ruby

DEPENDENCIES

BUNDLED WITH
   2.1.4
	`), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// WithBananaPantsApp creates a simple application to use with your CLI command
// (typically CF Push). When pushing, be aware of specifying '-b
// staticfile_buildpack" so that your app will correctly start up with the
// proper buildpack.
func WithBananaPantsApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-app")
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err := ioutil.WriteFile(tempfile, []byte("Banana Pants"), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

func WithEmptyFilesApp(f func(dir string)) {
	dir := TempDirAbsolutePath("", "simple-app")
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err := ioutil.WriteFile(tempfile, nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// AppGUID returns the GUID for an app in the currently targeted space.
func AppGUID(appName string) string {
	session := CF("app", appName, "--guid")
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

// AppJSON returns the JSON representation of an app by name.
func AppJSON(appName string) string {
	appGUID := AppGUID(appName)
	session := CF("curl", fmt.Sprintf("/v3/apps/%s", appGUID))
	Eventually(session).Should(Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}

// WriteManifest will write out a YAML manifest file at the specified path.
func WriteManifest(path string, manifest map[string]interface{}) {
	body, err := yaml.Marshal(manifest)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(path, body, 0666)
	Expect(err).ToNot(HaveOccurred())
}

func ReadManifest(path string) map[string]interface{} {
	manifestBytes, err := ioutil.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())
	var manifest map[string]interface{}
	err = yaml.Unmarshal(manifestBytes, &manifest)
	Expect(err).ToNot(HaveOccurred())
	return manifest
}

// Zipit zips the source into a .zip file in the target dir.
func Zipit(source, target, prefix string) error {
	// Thanks to Svett Ralchev
	// http://blog.ralch.com/tutorial/golang-working-with-zip/

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	if prefix != "" {
		_, err = io.WriteString(zipfile, prefix)
		if err != nil {
			return err
		}
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == source {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(header.Name)

		if info.IsDir() {
			header.Name += "/"
			header.SetMode(0755)
		} else {
			header.Method = zip.Deflate
			header.SetMode(0744)
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// ConfirmStagingLogs checks session for log output
// indicating that staging is working.
func ConfirmStagingLogs(session *Session) {
	Eventually(session).Should(gbytes.Say(`(?i)Creating container|Successfully created container|Staging\.\.\.|Staging process started \.\.\.|Staging Complete|Exit status 0|Uploading droplet\.\.\.|Uploading complete`))
}

func WaitForAppMemoryToTakeEffect(appName string, processIndex int, instanceIndex int, shouldRestartFirst bool, expectedMemory string) {
	if shouldRestartFirst {
		session := CF("restart", appName)
		Eventually(session).Should(Exit(0))
	}

	Eventually(func() string {
		session := CF("app", appName)
		Eventually(session).Should(Exit(0))
		appTable := ParseV3AppProcessTable(session.Out.Contents())
		return appTable.Processes[processIndex].Instances[instanceIndex].Memory
	}).Should(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, expectedMemory)))
}

func WaitForAppDiskToTakeEffect(appName string, processIndex int, instanceIndex int, shouldRestartFirst bool, expectedDisk string) {
	if shouldRestartFirst {
		session := CF("restart", appName)
		Eventually(session).Should(Exit(0))
	}

	Eventually(func() string {
		session := CF("app", appName)
		Eventually(session).Should(Exit(0))
		appTable := ParseV3AppProcessTable(session.Out.Contents())
		return appTable.Processes[processIndex].Instances[instanceIndex].Disk
	}).Should(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, expectedDisk)))
}

func WaitForLogRateLimitToTakeEffect(appName string, processIndex int, instanceIndex int, shouldRestartFirst bool, expectedLogRateLimit string) {
	if shouldRestartFirst {
		session := CF("restart", appName)
		Eventually(session).Should(Exit(0))
	}

	Eventually(func() string {
		session := CF("app", appName)
		Eventually(session).Should(Exit(0))
		appTable := ParseV3AppProcessTable(session.Out.Contents())
		return appTable.Processes[processIndex].Instances[instanceIndex].LogRate
	}).Should(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]?/s of %s/s`, expectedLogRateLimit)))
}
