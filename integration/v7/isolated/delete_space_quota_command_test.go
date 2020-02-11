package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-space-quota command", func() {
	Describe("help text", func() {
		When("the --help flag is passed", func() {
			It("Displays the appropriate help text", func() {
				session := helpers.CF("delete-space-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("delete-space-quota - Delete a space quota"))
				Eventually(session).Should(Say("\n"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf delete-space-quota QUOTA \[-f]`))
				Eventually(session).Should(Say("\n"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--force, -f\s+Force deletion without confirmation`))
				Eventually(session).Should(Say("\n"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space-quotas"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("command behavior", func() {
		var (
			quotaName string
			userName  string
			orgName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			helpers.SetupCFWithOrgOnly(orgName)
			quotaName = helpers.QuotaName()
			userName, _ = helpers.GetCredentials()
		})

		When("deleting a space quota", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				session := helpers.CF("create-space-quota", quotaName)
				Eventually(session).Should(Exit(0))

				_, err := buffer.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("prompts for confirmation and deletes the space quota", func() {
				session := helpers.CFWithStdin(buffer, "delete-space-quota", quotaName)
				Eventually(session).Should(Say(`Really delete the space quota %s in org %s\? \[yN\]`, quotaName, orgName))
				Eventually(session).Should(Say(`Deleting space quota %s for org %s as %s\.\.\.`, quotaName, orgName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("space-quotas")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).NotTo(ContainSubstring(quotaName))
			})
		})

		Context("the -f flag is provided", func() {
			BeforeEach(func() {
				session := helpers.CF("create-space-quota", quotaName)
				Eventually(session).Should(Exit(0))
			})

			It("deletes the specified quota without prompting", func() {
				session := helpers.CF("delete-space-quota", quotaName, "-f")
				Eventually(session).Should(Say(`Deleting space quota %s for org %s as %s\.\.\.`, quotaName, orgName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the quota name is not provided", func() {
			It("displays an error and help", func() {
				session := helpers.CF("delete-space-quota")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `QUOTA` was not provided"))
				Eventually(session).Should(Say("USAGE"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota doesn't exist", func() {
			It("displays a warning and exits 0", func() {
				session := helpers.CF("delete-space-quota", "-f", "nonexistent-quota")
				Eventually(session).Should(Say(`Deleting space quota nonexistent-quota for org %s as %s\.\.\.`, orgName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session.Err).Should(Say(`Space quota with name 'nonexistent-quota' not found\.`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
