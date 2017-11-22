package scp_test

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/diego-ssh/scp"
	"code.cloudfoundry.org/diego-ssh/scp/atime"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestCopier interface {
	scp.SecureCopier

	SendDirectory(dir string, dirInfo os.FileInfo) error
	ReceiveDirectory(dir string, timeStampMessage *scp.TimeMessage) error

	SendFile(file *os.File, fileInfo os.FileInfo) error
	ReceiveFile(path string, pathIsDir bool, timeMessage *scp.TimeMessage) error
}

var fileInfos map[string]os.FileInfo

var _ = Describe("scp", func() {
	var (
		stdin, stdoutSource io.ReadCloser
		stdinSource, stdout io.WriteCloser
		stderr              io.Writer

		sourceDir                   string
		sourceDirInfo               os.FileInfo
		targetDir                   string
		nestedTempDir               string
		nestedTempDirInfo           os.FileInfo
		generatedTextFile           string
		generatedTextFileInfo       os.FileInfo
		generatedNestedTextFile     string
		generatedNestedTextFileInfo os.FileInfo
		generatedBinaryFile         string
		generatedBinaryFileInfo     os.FileInfo

		secureCopier scp.SecureCopier
		logger       *lagertest.TestLogger

		testCopier TestCopier
	)

	newTestCopier := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, preserveTimeAndMode bool) TestCopier {
		options := &scp.Options{
			PreserveTimesAndMode: preserveTimeAndMode,
		}
		secureCopier, ok := scp.New(options, stdin, stdout, stderr, logger).(TestCopier)
		Expect(ok).To(BeTrue())
		return secureCopier
	}

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		fileInfos = make(map[string]os.FileInfo)

		stdin, stdinSource = io.Pipe()
		stdoutSource, stdout = io.Pipe()
		stderr = ioutil.Discard

		var err error
		sourceDir, err = ioutil.TempDir("", "scp-source")
		Expect(err).NotTo(HaveOccurred())

		fileContents := []byte("---\nthis is a simple file\n\n")
		generatedTextFile = filepath.Join(sourceDir, "textfile.txt")

		err = ioutil.WriteFile(generatedTextFile, fileContents, 0664)
		Expect(err).NotTo(HaveOccurred())

		fileContents = make([]byte, 1024)
		generatedBinaryFile = filepath.Join(sourceDir, "binary.dat")

		_, err = rand.Read(fileContents)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(generatedBinaryFile, fileContents, 0400)
		Expect(err).NotTo(HaveOccurred())

		nestedTempDir, err = ioutil.TempDir(sourceDir, "nested")
		Expect(err).NotTo(HaveOccurred())

		nestedFileContents := []byte("---\nthis is a simple nested file\n\n")
		generatedNestedTextFile = filepath.Join(nestedTempDir, "nested-textfile.txt")

		err = ioutil.WriteFile(generatedNestedTextFile, nestedFileContents, 0664)
		Expect(err).NotTo(HaveOccurred())

		// save off file infos
		sourceDirInfo, err = os.Stat(sourceDir)
		Expect(err).NotTo(HaveOccurred())
		fileInfos[sourceDir] = sourceDirInfo

		generatedTextFileInfo, err = os.Stat(generatedTextFile)
		Expect(err).NotTo(HaveOccurred())
		fileInfos[generatedTextFile] = generatedTextFileInfo

		generatedBinaryFileInfo, err = os.Stat(generatedBinaryFile)
		Expect(err).NotTo(HaveOccurred())
		fileInfos[generatedBinaryFile] = generatedBinaryFileInfo

		nestedTempDirInfo, err = os.Stat(nestedTempDir)
		Expect(err).NotTo(HaveOccurred())
		fileInfos[nestedTempDir] = nestedTempDirInfo

		generatedNestedTextFileInfo, err = os.Stat(generatedNestedTextFile)
		Expect(err).NotTo(HaveOccurred())
		fileInfos[generatedNestedTextFile] = generatedNestedTextFileInfo

		targetDir, err = ioutil.TempDir("", "scp-target")
		Expect(err).NotTo(HaveOccurred())

		secureCopier = nil
	})

	AfterEach(func() {
		os.RemoveAll(sourceDir)
		os.RemoveAll(targetDir)
	})

	Context("source mode", func() {
		var preserveTimestamps bool
		Context("when no files are requested", func() {
			It("fails construct the copier", func() {
				_, err := scp.NewFromCommand("scp -f", stdin, stdout, stderr, logger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when files are requested", func() {
			var sourceFileInfo os.FileInfo

			BeforeEach(func() {
				preserveTimestamps = false
			})

			Context("when the requested file exists", func() {
				JustBeforeEach(func() {
					var err error

					command := fmt.Sprintf("scp -f %s", generatedTextFile)
					if preserveTimestamps {
						command = fmt.Sprintf("scp -fp %s", generatedTextFile)
					}

					secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					done := make(chan struct{})
					go func() {
						err := secureCopier.Copy()
						Expect(err).NotTo(HaveOccurred())
						close(done)
					}()

					_, err = stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					session := scp.NewSession(stdoutSource, stdinSource, nil, preserveTimestamps, logger)

					var timestampMessage *scp.TimeMessage
					if preserveTimestamps {
						timestampMessage = &scp.TimeMessage{}
						err = timestampMessage.Receive(session)
						Expect(err).NotTo(HaveOccurred())
					}

					testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)
					err = testCopier.ReceiveFile(targetDir, true, timestampMessage)
					Expect(err).NotTo(HaveOccurred())
					Eventually(done).Should(BeClosed())

					sourceFileInfo, err = os.Stat(generatedTextFile)
					Expect(err).NotTo(HaveOccurred())
				})

				It("sends the file", func() {
					compareFile(filepath.Join(targetDir, sourceFileInfo.Name()), generatedTextFile, preserveTimestamps)
				})

				Context("when -p (preserve times) is specified", func() {
					BeforeEach(func() {
						preserveTimestamps = true
					})

					It("sends the timestamp information before the file", func() {
						compareFile(filepath.Join(targetDir, sourceFileInfo.Name()), generatedTextFile, preserveTimestamps)
					})
				})
			})

			Context("when the requested file does not exist", func() {
				BeforeEach(func() {
					os.RemoveAll(generatedTextFile)
				})

				It("returns an error and continues sending", func() {
					testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)

					command := fmt.Sprintf("scp -f %s %s", generatedTextFile, generatedBinaryFile)
					secureCopier, err := scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					errCh := make(chan error)
					go func() {
						errCh <- secureCopier.Copy()
					}()

					_, err = stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					stdoutReader := bufio.NewReader(stdoutSource)

					errCode, err := stdoutReader.ReadByte()
					Expect(err).NotTo(HaveOccurred())
					Expect(errCode).To(BeEquivalentTo(1))

					errMessage, err := stdoutReader.ReadString('\n')
					Expect(err).NotTo(HaveOccurred())
					Expect(errMessage).To(ContainSubstring("no such file or directory"))

					err = testCopier.ReceiveFile(targetDir, true, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(errCh).Should(Receive(HaveOccurred()))

					compareFile(filepath.Join(targetDir, "binary.dat"), generatedBinaryFile, false)
				})
			})
		})

		Context("when a directory is requested", func() {
			Context("when the -r (recursive) flag is not specified", func() {
				BeforeEach(func() {
					var err error
					command := fmt.Sprintf("scp -f %s %s", sourceDir, generatedTextFile)
					secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error and continues sending sources", func() {
					testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)

					errCh := make(chan error)
					go func() {
						errCh <- secureCopier.Copy()
					}()

					_, err := stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					stdoutReader := bufio.NewReader(stdoutSource)

					errCode, err := stdoutReader.ReadByte()
					Expect(err).NotTo(HaveOccurred())
					Expect(errCode).To(BeEquivalentTo(1))

					errMessage, err := stdoutReader.ReadString('\n')
					Expect(err).NotTo(HaveOccurred())
					Expect(errMessage).To(ContainSubstring("not a regular file"))

					err = testCopier.ReceiveFile(targetDir, true, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(errCh).Should(Receive(HaveOccurred()))
					compareFile(filepath.Join(targetDir, "textfile.txt"), generatedTextFile, false)
				})
			})

			Context("when the -r (recursive) flag is specified", func() {
				var sourceDirInfo os.FileInfo

				BeforeEach(func() {
					preserveTimestamps = false
				})

				JustBeforeEach(func() {
					var err error

					command := fmt.Sprintf("scp -rf %s", sourceDir)
					if preserveTimestamps {
						command = fmt.Sprintf("scp -rfp %s", sourceDir)
					}

					secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					done := make(chan struct{})
					go func() {
						err := secureCopier.Copy()
						Expect(err).NotTo(HaveOccurred())
						close(done)
					}()

					_, err = stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					session := scp.NewSession(stdoutSource, stdinSource, nil, preserveTimestamps, logger)

					timestampMessage := &scp.TimeMessage{}
					if preserveTimestamps {
						err = timestampMessage.Receive(session)
						Expect(err).NotTo(HaveOccurred())
					}

					testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)
					err = testCopier.ReceiveDirectory(targetDir, timestampMessage)
					Expect(err).NotTo(HaveOccurred())
					Eventually(done).Should(BeClosed())

					sourceDirInfo, err = os.Stat(sourceDir)
					Expect(err).NotTo(HaveOccurred())
				})

				It("sends the directory and all the files", func() {
					compareDir(filepath.Join(targetDir, sourceDirInfo.Name()), sourceDir, preserveTimestamps)
				})

				Context("when the -p is specified", func() {
					BeforeEach(func() {
						preserveTimestamps = true
					})

					It("sends timestamp information before files and directories", func() {
						compareDir(filepath.Join(targetDir, sourceDirInfo.Name()), sourceDir, preserveTimestamps)
					})
				})
			})
		})

		Context("when a glob is requested", func() {
			var (
				command string
			)

			BeforeEach(func() {
				command = fmt.Sprintf("scp -f %s/[bt]*", sourceDir)
			})

			Context("when the glob is valid", func() {

				JustBeforeEach(func() {
					var err error

					secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					done := make(chan struct{})
					go func() {
						err := secureCopier.Copy()
						Expect(err).NotTo(HaveOccurred())
						close(done)
					}()

					_, err = stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					testCopier = newTestCopier(stdoutSource, stdinSource, nil, false)

					// Receive File 1
					err = testCopier.ReceiveFile(targetDir, true, nil)
					Expect(err).NotTo(HaveOccurred())

					// Receive File 2
					err = testCopier.ReceiveFile(targetDir, true, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(done).Should(BeClosed())
				})

				It("properly matches the glob against a single filename", func() {
					compareFile(filepath.Join(targetDir, "textfile.txt"), generatedTextFile, false)
					compareFile(filepath.Join(targetDir, "binary.dat"), generatedBinaryFile, false)
				})
			})

			Context("when the glob does not match any sources", func() {
				var generatedBadGlobFile string

				JustBeforeEach(func() {
					fileContents := []byte("---\nthis is a bad glob file\n\n")

					err := ioutil.WriteFile(generatedBadGlobFile, fileContents, 0664)
					Expect(err).NotTo(HaveOccurred())

					generatedBadGlobFileInfo, err := os.Stat(generatedBadGlobFile)
					Expect(err).NotTo(HaveOccurred())
					fileInfos[generatedBadGlobFile] = generatedBadGlobFileInfo

					command = fmt.Sprintf("scp -f %s", generatedBadGlobFile)
					secureCopier, err := scp.NewFromCommand(command, stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					done := make(chan struct{})
					go func() {
						err := secureCopier.Copy()
						Expect(err).NotTo(HaveOccurred())
						close(done)
					}()

					_, err = stdinSource.Write([]byte{0})
					Expect(err).NotTo(HaveOccurred())

					testCopier = newTestCopier(stdoutSource, stdinSource, nil, false)

					// Receive File 1
					err = testCopier.ReceiveFile(targetDir, true, nil)
					Expect(err).NotTo(HaveOccurred())

					Eventually(done).Should(BeClosed())
				})

				Context("because it is malformed", func() {
					BeforeEach(func() {
						generatedBadGlobFile = filepath.Join(sourceDir, "[")
					})

					It("attempts to match the glob literally", func() {
						compareFile(filepath.Join(targetDir, "["), generatedBadGlobFile, false)
					})
				})

				Context("because nothing matches the glob", func() {
					BeforeEach(func() {
						generatedBadGlobFile = filepath.Join(sourceDir, "[a].txt")
					})

					It("attempts to match the glob literally", func() {
						compareFile(filepath.Join(targetDir, "[a].txt"), generatedBadGlobFile, false)
					})
				})
			})
		})
	})

	Context("target mode", func() {
		Context("when no target is specified", func() {
			It("fails construct the copier", func() {
				_, err := scp.NewFromCommand("scp -t", stdin, stdout, stderr, logger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when multiple targets are specified", func() {
			It("fails construct the copier", func() {
				_, err := scp.NewFromCommand("scp -t a b", stdin, stdout, stderr, logger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the target is not a directory", func() {
			Context("and the target is specified as a directory", func() {
				It("fails when the target does not exist", func() {
					secureCopier, err := scp.NewFromCommand("scp -td bogus", stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					err = secureCopier.Copy()
					Expect(err).To(HaveOccurred())
				})

				It("fails when the target is not a directory", func() {
					tempFile, err := ioutil.TempFile(targetDir, "target")
					Expect(err).NotTo(HaveOccurred())

					secureCopier, err := scp.NewFromCommand("scp -td "+tempFile.Name(), stdin, stdout, stderr, logger)
					Expect(err).NotTo(HaveOccurred())

					err = secureCopier.Copy()
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when a file is specified as the target", func() {
			var (
				targetFile         string
				preserveTimestamps bool
			)

			BeforeEach(func() {
				preserveTimestamps = false
				targetFile = filepath.Join(targetDir, "targetFile")
			})

			JustBeforeEach(func() {
				var err error

				args := "-t"
				if preserveTimestamps {
					args += "p"
				}
				command := fmt.Sprintf("scp %s %s", args, targetFile)

				secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
				Expect(err).NotTo(HaveOccurred())

				done := make(chan struct{})
				go func() {
					defer GinkgoRecover()
					err := secureCopier.Copy()
					Expect(err).NotTo(HaveOccurred())
					close(done)
				}()

				bytes := make([]byte, 1)
				_, err = stdoutSource.Read(bytes)
				Expect(err).NotTo(HaveOccurred())

				textFile, err := os.Open(generatedTextFile)
				Expect(err).NotTo(HaveOccurred())

				textFileInfo, err := textFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)
				err = testCopier.SendFile(textFile, textFileInfo)
				Expect(err).NotTo(HaveOccurred())
				stdinSource.Close()
				Eventually(done).Should(BeClosed())

				_, err = os.Stat(targetFile)
				Expect(err).NotTo(HaveOccurred())
			})

			It("allows a file to be sent", func() {
				compareFile(targetFile, generatedTextFile, preserveTimestamps)
			})

			Context("when preserving timestamps and mode", func() {
				BeforeEach(func() {
					preserveTimestamps = true
				})

				It("sets the mode and timestamp", func() {
					compareFile(targetFile, generatedTextFile, preserveTimestamps)
				})

				Context("when the target file exists", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(targetFile, []byte{'a'}, 0640)
						Expect(err).NotTo(HaveOccurred())

						modificationTime := time.Unix(123456789, 12345678)
						accessTime := time.Unix(987654321, 987654321)
						err = os.Chtimes(targetFile, accessTime, modificationTime)
						Expect(err).NotTo(HaveOccurred())
					})

					It("sets the mode and timestamp", func() {
						targetFileInfo, err := os.Stat(targetFile)
						Expect(err).NotTo(HaveOccurred())
						compareFileInfo(targetFileInfo, generatedTextFileInfo, preserveTimestamps)
					})
				})
			})
		})

		Context("when a directory is specified as the target", func() {
			var (
				dir                string
				preserveTimestamps bool
				targetIsDirectory  bool
				done               chan struct{}
			)

			BeforeEach(func() {
				dir = targetDir
				preserveTimestamps = false
				targetIsDirectory = false
			})

			JustBeforeEach(func() {
				var err error

				args := "-t"
				if preserveTimestamps {
					args += "p"
				}
				if targetIsDirectory {
					args += "d"
				}
				command := fmt.Sprintf("scp %s %s", args, dir)

				secureCopier, err = scp.NewFromCommand(command, stdin, stdout, stderr, logger)
				Expect(err).NotTo(HaveOccurred())

				done = make(chan struct{})
				go func() {
					defer GinkgoRecover()
					err := secureCopier.Copy()
					Expect(err).NotTo(HaveOccurred())
					close(done)
				}()

				bytes := make([]byte, 1)
				_, err = stdoutSource.Read(bytes)
				Expect(err).NotTo(HaveOccurred())

				scp.NewSession(stdoutSource, stdinSource, nil, preserveTimestamps, logger)
				testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)
			})

			Context("and a file is sent", func() {
				var file *os.File
				var err error

				JustBeforeEach(func() {
					file, err = os.Open(generatedTextFile)
					Expect(err).NotTo(HaveOccurred())

					fileInfo, err := file.Stat()
					Expect(err).NotTo(HaveOccurred())

					testCopier = newTestCopier(stdoutSource, stdinSource, nil, preserveTimestamps)
					err = testCopier.SendFile(file, fileInfo)
					Expect(err).NotTo(HaveOccurred())

					stdinSource.Close()
					Eventually(done).Should(BeClosed())
				})

				It("copies the file and its contents into the target", func() {
					compareFile(filepath.Join(dir, filepath.Base(file.Name())), generatedTextFile, preserveTimestamps)
				})
			})

			Context("and a directory is sent", func() {
				JustBeforeEach(func() {
					sourceDirInfo, err := os.Stat(sourceDir)
					Expect(err).NotTo(HaveOccurred())

					err = testCopier.SendDirectory(sourceDir, sourceDirInfo)
					Expect(err).NotTo(HaveOccurred())
					stdinSource.Close()
					Eventually(done).Should(BeClosed())
				})

				It("receives the directory and its content", func() {
					compareDir(filepath.Join(dir, filepath.Base(sourceDir)), sourceDir, preserveTimestamps)
				})

				Context("when the target directory does not exist but its parent directory does", func() {
					BeforeEach(func() {
						dir = filepath.Join(targetDir, "newdir")
					})

					It("makes the target directory and populates with the source directories contents", func() {
						compareDir(dir, sourceDir, preserveTimestamps)
					})
				})
			})
		})

		Context("when an unknown message type is sent", func() {
			It("returns an error", func() {
				secureCopier, err := scp.NewFromCommand("scp -t /tmp/foo", stdin, stdout, stderr, logger)
				Expect(err).NotTo(HaveOccurred())

				errCh := make(chan error)
				go func() {
					defer GinkgoRecover()
					errCh <- secureCopier.Copy()
				}()

				bytes := make([]byte, 1)
				_, err = stdoutSource.Read(bytes)
				Expect(err).NotTo(HaveOccurred())

				_, err = stdinSource.Write([]byte("F this protocol message does not exist"))
				Expect(err).NotTo(HaveOccurred())

				stdoutReader := bufio.NewReader(stdoutSource)

				errCode, err := stdoutReader.ReadByte()
				Expect(err).NotTo(HaveOccurred())
				Expect(errCode).To(BeEquivalentTo(1))

				errMessage, err := stdoutReader.ReadString('\n')
				Expect(err).NotTo(HaveOccurred())
				Expect(errMessage).To(ContainSubstring("unexpected message type: F"))

				Eventually(errCh).Should(Receive(HaveOccurred()))
			})
		})
	})
})

func compareDir(actualDir, expectedDir string, compareTimestamps bool) {
	actualDirInfo, err := os.Stat(actualDir)
	Expect(err).NotTo(HaveOccurred())

	expectedDirInfo, ok := fileInfos[expectedDir]
	Expect(ok).To(BeTrue())

	Expect(actualDirInfo.Mode()).To(Equal(expectedDirInfo.Mode()))
	if compareTimestamps {
		compareTimestampsFromInfo(actualDirInfo, expectedDirInfo)
	}

	actualFiles, err := ioutil.ReadDir(actualDir)
	Expect(err).NotTo(HaveOccurred())

	expectedFiles, err := ioutil.ReadDir(expectedDir)
	Expect(err).NotTo(HaveOccurred())

	Expect(len(actualFiles)).To(Equal(len(expectedFiles)))
	for i, actualFile := range actualFiles {
		expectedFile := expectedFiles[i]
		if actualFile.IsDir() {
			compareDir(filepath.Join(actualDir, actualFile.Name()), filepath.Join(expectedDir, expectedFile.Name()), compareTimestamps)
		} else {
			compareFile(filepath.Join(actualDir, actualFile.Name()), filepath.Join(expectedDir, expectedFile.Name()), compareTimestamps)
		}
	}
}

func compareFile(actualFile, expectedFile string, compareTimestamps bool) {
	actualFileInfo, err := os.Stat(actualFile)
	Expect(err).NotTo(HaveOccurred())

	expectedFileInfo, ok := fileInfos[expectedFile]
	Expect(ok).To(BeTrue())

	compareFileInfo(actualFileInfo, expectedFileInfo, compareTimestamps)

	actualContents, err := ioutil.ReadFile(actualFile)
	Expect(err).NotTo(HaveOccurred())

	expectedContents, err := ioutil.ReadFile(expectedFile)
	Expect(err).NotTo(HaveOccurred())

	Expect(actualContents).To(Equal(expectedContents))
}

func compareFileInfo(actualFileInfo os.FileInfo, expectedFileInfo os.FileInfo, compareTimestamps bool) {
	Expect(actualFileInfo.Mode()).To(Equal(expectedFileInfo.Mode()))
	Expect(actualFileInfo.Size()).To(Equal(expectedFileInfo.Size()))
	if compareTimestamps {
		compareTimestampsFromInfo(actualFileInfo, expectedFileInfo)
	}
}

func compareTimestampsFromInfo(actualInfo, expectedInfo os.FileInfo) {
	actualAccessTime, err := atime.AccessTime(actualInfo)
	Expect(err).NotTo(HaveOccurred())

	expectedAccessTime, err := atime.AccessTime(expectedInfo)
	Expect(err).NotTo(HaveOccurred())

	Expect(actualInfo.ModTime().Unix()).To(Equal(expectedInfo.ModTime().Unix()))
	Expect(actualAccessTime.Unix()).To(Equal(expectedAccessTime.Unix()))
}
