package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("curl command", func() {
	var ExpectHelpText = func(session *Session) {
		Eventually(session).Should(Say(`NAME:\n`))
		Eventually(session).Should(Say(`curl - Executes a request to the targeted API endpoint\n`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`USAGE:\n`))
		Eventually(session).Should(Say(`\s+cf curl PATH \[-iv\] \[-X METHOD\] \[-H HEADER\] \[-d DATA\] \[--output FILE\]`))
		Eventually(session).Should(Say(`\s+By default 'cf curl' will perform a GET to the specified PATH. If data`))
		Eventually(session).Should(Say(`\s+is provided via -d, a POST will be performed instead, and the Content-Type\n`))
		Eventually(session).Should(Say(`\s+will be set to application/json. You may override headers with -H and the\n`))
		Eventually(session).Should(Say(`\s+request method with -X.\n`))
		Eventually(session).Should(Say(`\s+For API documentation, please visit http://apidocs.cloudfoundry.org.\n`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`EXAMPLES:\n`))
		Eventually(session).Should(Say(`\s+cf curl \"/v2/apps\" -X GET -H \"Content-Type: application/x-www-form-urlencoded\" -d 'q=name:myapp'`))
		Eventually(session).Should(Say(`\s+cf curl \"/v2/apps\" -d @/path/to/file`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`OPTIONS:\n`))
		Eventually(session).Should(Say(`\s+-H\s+Custom headers to include in the request, flag can be specified multiple times`))
		Eventually(session).Should(Say(`\s+-X\s+HTTP method \(GET,POST,PUT,DELETE,etc\)`))
		Eventually(session).Should(Say(`\s+-d\s+HTTP data to include in the request body, or '@' followed by a file name to read the data from`))
		Eventually(session).Should(Say(`\s+-i\s+Include response headers in the output`))
		Eventually(session).Should(Say(`\s+--output\s+Write curl body to FILE instead of stdout`))
	}

	var ExpectRequestHeaders = func(session *Session) {
		Eventually(session).Should(Say(`REQUEST: .+`))
		Eventually(session).Should(Say(`(GET|POST|PUT|DELETE) /v2/apps HTTP/1.1`))
		Eventually(session).Should(Say(`Host: .+`))
		Eventually(session).Should(Say(`Accept: .+`))
		Eventually(session).Should(Say(`Authorization:\s+\[PRIVATE DATA HIDDEN\]`))
		Eventually(session).Should(Say(`Content-Type: .+`))
		Eventually(session).Should(Say(`User-Agent: .+`))
	}

	var ExpectResponseHeaders = func(session *Session) {
		Eventually(session).Should(Say("HTTP/1.1 200 OK"))
		Eventually(session).Should(Say(`Connection: .+`))
		Eventually(session).Should(Say(`Content-Length: .+`))
		Eventually(session).Should(Say(`Content-Type: .+`))
		Eventually(session).Should(Say(`Date: .+`))
		Eventually(session).Should(Say(`Server: .+`))
		Eventually(session).Should(Say(`X-Content-Type-Options: .+`))
		Eventually(session).Should(Say(`X-Vcap-Request-Id: .+`))
	}

	Describe("Help Text", func() {
		When("--help flag is set", func() {
			It("Displays command usage to the output", func() {
				session := helpers.CF("curl", "--help")
				ExpectHelpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("Incorrect Usage", func() {
		When("no arguments are provided", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PATH` was not provided"))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("unknown flag is specified", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl", "--test")
				// TODO Legacy cf uses a weird quote around test. This test needs be fixed for refactored command
				Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `test'"))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("more than one path is specified", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl", "/v2/apps", "/v2/apps")
				Eventually(session).Should(Say("FAILED\n"))
				// TODO Legacy code uses Incorrect Usage.(dot) instead of Incorrect Usage: (colon). Fix this test after refactor
				Eventually(session).Should(Say("Incorrect Usage. An argument is missing or not correctly enclosed."))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the user is not logged in", func() {
		It("makes the request and receives an unauthenticated error", func() {
			session := helpers.CF("curl", "/v2/apps")
			expectedJSON := `{
				 "description": "Authentication error",
				 "error_code": "CF-NotAuthenticated",
				 "code": 10002
			}`
			Eventually(session).Should(Exit(0))
			Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
		})
	})
	Context("User Agent", func() {
		It("sets the User-Agent Header to match the CLI version", func() {
			session := helpers.CF("curl", "/v2/info", "-v")
			Eventually(session).Should(Exit(0))

			//TODO The user agent is different in refactored commands, so this should when
			// we refactor cf curl
			versionPattern := `\d{1,}\.\d{2,}\.\d+\+[a-f0-9]+\.\d{4}-\d{2}-\d{2} / \w`
			Expect(session).To(Say(`User-Agent: go-cli %s`, versionPattern))
		})
	})

	When("the user is logged in", func() {
		var orgName string

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName := helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			helpers.SwitchToOrgRole(orgName, "OrgManager")
			helpers.TargetOrg(orgName)
		})

		When("the path is valid", func() {
			var expectedJSON string

			BeforeEach(func() {
				expectedJSON = `{
            "total_results": 0,
            "total_pages": 1,
            "prev_url": null,
            "next_url": null,
            "resources": []
				}`
			})

			When("the path has multiple initial slashes", func() {
				It("changes the path to use only one slash", func() {
					session := helpers.CF("curl", "////v2/apps", "-v")
					Eventually(session).Should(Exit(0))

					Eventually(session).Should(Say(`GET /v2/apps HTTP/1.1`))
				})
			})

			When("the path has no initial slashes", func() {
				It("prepends a slash to the path", func() {
					session := helpers.CF("curl", "v2/apps", "-v")
					Eventually(session).Should(Exit(0))

					Eventually(session).Should(Say(`GET /v2/apps HTTP/1.1`))
				})
			})

			When("no flag is set", func() {
				It("makes the request and displays the json response", func() {
					session := helpers.CF("curl", "/v2/apps")
					Eventually(session).Should(Exit(0))
					Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
				})
			})

			When("-i flag is set", func() {
				It("displays the response headers", func() {
					session := helpers.CF("curl", "/v2/apps", "-i")
					Eventually(session).Should(Exit(0))

					ExpectResponseHeaders(session)
					contents := string(session.Out.Contents())
					jsonStartsAt := strings.Index(contents, "{")

					actualJSON := contents[jsonStartsAt:]
					Expect(actualJSON).To(MatchJSON(expectedJSON))
				})
			})

			When("-v flag is set", func() {
				It("displays the request headers and response headers", func() {
					session := helpers.CF("curl", "/v2/apps", "-v")
					Eventually(session).Should(Exit(0))

					ExpectRequestHeaders(session)
					ExpectResponseHeaders(session)

					contents := string(session.Out.Contents())
					jsonStartsAt := strings.Index(contents, "{")

					actualJSON := contents[jsonStartsAt:]
					Expect(actualJSON).To(MatchJSON(expectedJSON))
				})
			})

			When("-H is passed with a custom header", func() {
				When("the custom header is valid", func() {
					It("add the custom header to the request", func() {
						session := helpers.CF("curl", "/v2/apps", "-H", "X-Foo: bar", "-v")
						Eventually(session).Should(Exit(0))

						Expect(session).To(Say("REQUEST:"))
						Expect(session).To(Say("X-Foo: bar"))
						Expect(session).To(Say("RESPONSE:"))
					})

					When("multiple headers are provided", func() {
						It("adds all the custom headers to the request", func() {
							session := helpers.CF("curl", "/v2/apps", "-H", "X-Bar: bar", "-H", "X-Foo: foo", "-v")
							Eventually(session).Should(Exit(0))

							Expect(session).To(Say("REQUEST:"))
							Expect(session).To(Say("X-Bar: bar"))
							Expect(session).To(Say("X-Foo: foo"))
							Expect(session).To(Say("RESPONSE:"))
						})

						When("the same header field is passed", func() {
							It("adds the same header field twice", func() {
								session := helpers.CF("curl", "/v2/apps", "-H", "X-Foo: bar", "-H", "X-Foo: foo", "-v")
								Eventually(session).Should(Exit(0))

								Expect(session).To(Say("REQUEST:"))
								Expect(session).To(Say("X-Foo: bar"))
								Expect(session).To(Say("X-Foo: foo"))
								Expect(session).To(Say("RESPONSE:"))
							})
						})
					})

					When("-H is provided with a default header", func() {
						It("overrides the value of User-Agent header", func() {
							session := helpers.CF("curl", "/v2/apps", "-H", "User-Agent: smith", "-v")
							Eventually(session).Should(Exit(0))

							Expect(session).To(Say("REQUEST:"))
							Expect(session).To(Say("User-Agent: smith"))
							Expect(session).To(Say("RESPONSE:"))
						})

						It("does not override the Host header", func() {
							session := helpers.CF("curl", "/v2/apps", "-H", "Host: example.com", "-v")
							Eventually(session).Should(Exit(0))

							Expect(session).ToNot(Say("Host: example.com"))
						})

						It("overrides the value of Accept header", func() {
							session := helpers.CF("curl", "/v2/apps", "-H", "Accept: application/xml", "-v")
							Eventually(session).Should(Exit(0))

							Expect(session).To(Say("REQUEST:"))
							Expect(session).To(Say("Accept: application/xml"))
							Expect(session).To(Say("RESPONSE:"))
						})

						It("overrides the value of Content-Type header", func() {
							session := helpers.CF("curl", "/v2/apps", "-H", "Content-Type: application/xml", "-v")
							Eventually(session).Should(Exit(0))

							Expect(session).To(Say("REQUEST:"))
							Expect(session).To(Say("Content-Type: application/xml"))
							Expect(session).To(Say("RESPONSE:"))
						})
					})
				})

				When("the custom header is not valid", func() {
					It("tells the user that the header is not valid", func() {
						session := helpers.CF("curl", "/v2/apps", "-H", "not-a-valid-header", "-v")
						Eventually(session).Should(Exit(1))

						Expect(session).Should(Say("FAILED"))
						Expect(session).Should(Say("Error creating request:"))
						Expect(session).Should(Say("Error parsing headers: malformed MIME header line: not-a-valid-header"))
					})
				})
			})

			When("-d is passed with a request body", func() {
				When("the request body is passed as a string", func() {
					It("sets the method to POST and sends the body", func() {
						orgGUID := helpers.GetOrgGUID(orgName)
						spaceName := helpers.NewSpaceName()
						jsonBody := fmt.Sprintf(`{"name":"%s", "organization_guid":"%s"}`, spaceName, orgGUID)
						session := helpers.CF("curl", "/v2/spaces", "-d", jsonBody)
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
					})
				})

				When("the request body is passed as a file", func() {
					var spaceName, filePath, dir string

					BeforeEach(func() {
						var err error
						dir, err = ioutil.TempDir("", "curl-command")
						Expect(err).ToNot(HaveOccurred())

						filePath = filepath.Join(dir, "request_body.json")
						orgGUID := helpers.GetOrgGUID(orgName)
						spaceName = helpers.NewSpaceName()

						jsonBody := fmt.Sprintf(`{"name":"%s", "organization_guid":"%s"}`, spaceName, orgGUID)
						err = ioutil.WriteFile(filePath, []byte(jsonBody), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						os.RemoveAll(dir)
					})

					It("sets the method to POST and sends the body", func() {
						session := helpers.CF("curl", "/v2/spaces", "-d", "@"+filePath)
						Eventually(session).Should(Exit(0))
						Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
					})

					When("the file does not exist", func() {
						It("fails and displays an error message", func() {
							_, err := os.Stat("this-file-does-not-exist")
							Expect(os.IsExist(err)).To(BeFalse())

							session := helpers.CF("curl", "/v2/spaces", "-d", "@this-file-does-not-exist")
							Eventually(session).Should(Exit(1))
							Expect(session).To(Say("no such file or directory"))
						})
					})

					When("the file is a symlink", func() {
						It("follows the symlink", func() {
							linkPath := filepath.Join(dir, "link-name.json")
							Expect(os.Symlink(filePath, linkPath)).To(Succeed())
							session := helpers.CF("curl", "-d", "@"+linkPath, "/v2/spaces")
							Eventually(session).Should(Exit(0))
							Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
						})
					})
				})
			})

			When("-X is passed with the HTTP method", func() {
				var spaceGUID, spaceName string

				BeforeEach(func() {
					spaceName = helpers.NewSpaceName()
					helpers.CreateSpace(spaceName)
					spaceGUID = helpers.GetSpaceGUID(spaceName)
				})

				It("changes the HTTP method of the request", func() {
					path := fmt.Sprintf("/v2/spaces/%s", spaceGUID)
					session := helpers.CF("curl", path, "-X", "DELETE", "-v")
					Eventually(session).Should(Exit(0))

					Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
				})
			})

			When("--output is passed with a file name", func() {
				It("writes the response body to the file", func() {
					outFile, err := ioutil.TempFile("", "output*.json")
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CF("curl", "/v2/apps", "-i", "--output", outFile.Name())
					Eventually(session).Should(Exit(0))
					ExpectResponseHeaders(session)
					body, err := ioutil.ReadFile(outFile.Name())
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(MatchJSON(expectedJSON))
				})
			})

			Context("Flag combinations", func() {
				When("-i and -v flags are set", func() {
					It("prints both the request and response headers", func() {
						session := helpers.CF("curl", "/v2/apps", "-v", "-i")
						Eventually(session).Should(Exit(0))

						ExpectRequestHeaders(session)
						ExpectResponseHeaders(session)

						contents := string(session.Out.Contents())
						jsonStartsAt := strings.Index(contents, "{")

						actualJSON := contents[jsonStartsAt:]
						Expect(actualJSON).To(MatchJSON(expectedJSON))
					})
				})

				XWhen("-v and --output flags are passed", func() {
					It("prints the headers to the terminal and the response to the file", func() {
						// TODO This is a bug in the legacy CLI. Please write the test and fix the bug after refactor [#162432878]
					})
				})

				When("-X, -H and -d flags are passed", func() {
					var spaceName, filePath, dir, jsonBody string

					BeforeEach(func() {
						var err error
						dir, err = ioutil.TempDir("", "curl-command")
						Expect(err).ToNot(HaveOccurred())

						filePath = filepath.Join(dir, "request_body.json")
						orgGUID := helpers.GetOrgGUID(orgName)
						spaceName = helpers.NewSpaceName()

						jsonBody = fmt.Sprintf(`{"name":"%s", "organization_guid":"%s"}`, spaceName, orgGUID)
						err = ioutil.WriteFile(filePath, []byte(jsonBody), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						os.RemoveAll(dir)
					})

					It("sets the custom header and use the request body from -d", func() {
						session := helpers.CF("curl", "/v2/spaces", "-X", "POST", "-H", "X-Foo: foo", "-H", "X-Bar: bar", "-d", "@"+filePath, "-v")
						Eventually(session).Should(Exit(0))

						Expect(session).Should(Say("REQUEST:"))
						Expect(session).Should(Say("POST"))

						Expect(session).Should(Say("X-Bar: bar"))
						Expect(session).Should(Say("X-Foo: foo"))

						Expect(session).Should(Say(jsonBody))

						Expect(session).Should(Say("RESPONSE:"))

						Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
					})
				})
			})

			Context("Refresh Token", func() {
				When("the auth token is invalid", func() {
					var spaceGUID, spaceName string

					BeforeEach(func() {
						spaceName = helpers.NewSpaceName()
						helpers.CreateSpace(spaceName)
						spaceGUID = helpers.GetSpaceGUID(spaceName)
					})

					It("generates a new auth token by using the refresh token", func() {
						path := fmt.Sprintf("/v2/spaces/%s", spaceGUID)
						session := helpers.CF("curl", path, "-H", "Authorization: bearer some-token", "-X", "DELETE", "-v")
						Eventually(session).Should(Exit(0))

						Expect(session).To(Say("POST /oauth/token"))

						Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
					})
				})
			})
		})

		When("PATH is invalid", func() {
			It("makes the request and displays the unknown request json", func() {
				expectedJSON := `{
				 "description": "Unknown request",
				 "error_code": "CF-NotFound",
				 "code": 10000
				}`
				session := helpers.CF("curl", "/some-random-path")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
			})
		})
	})
})
