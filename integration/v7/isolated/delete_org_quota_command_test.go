package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-org-quota command", func() {
	var (
		quotaName string
		userName  string
	)

	BeforeEach(func() {
		userName = helpers.LoginCF()
		quotaName = helpers.QuotaName()
	})

	When("the --help flag is passed", func() {
		It("Displays the appropriate help text", func() {
			session := helpers.CF("delete-org-quota", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("delete-org-quota - Delete an organization quota"))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf delete-org-quota QUOTA \[-f]`))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--force, -f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("org-quotas"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the quota name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-org-quota")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `QUOTA` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the quota doesn't exist", func() {
		It("displays a warning and exits 0", func() {
			session := helpers.CF("delete-org-quota", "-f", "nonexistent-quota")
			Eventually(session).Should(Say(`Deleting org quota nonexistent-quota as %s\.\.\.`, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session.Err).Should(Say(`Organization quota with name 'nonexistent-quota' not found\.`))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("the -f flag is provided", func() {
		BeforeEach(func() {
			session := helpers.CF("create-org-quota", quotaName)
			Eventually(session).Should(Exit(0))
		})

		It("deletes the specified quota", func() {
			session := helpers.CF("delete-org-quota", quotaName, "-f")
			Eventually(session).Should(Say(`Deleting org quota %s as %s\.\.\.`, quotaName, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the -f flag not is provided", func() {
		var buffer *Buffer

		BeforeEach(func() {
			buffer = NewBuffer()
			session := helpers.CF("create-org-quota", quotaName)
			Eventually(session).Should(Exit(0))
		})

		When("the user enters 'y'", func() {
			BeforeEach(func() {
				_, err := buffer.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the quota", func() {
				session := helpers.CFWithStdin(buffer, "delete-org-quota", quotaName)
				Eventually(session).Should(Say(`Deleting org quota %s as %s\.\.\.`, quotaName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org-quotas")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).NotTo(ContainSubstring(quotaName))
			})
		})

		When("the user enters 'n'", func() {
			BeforeEach(func() {
				_, err := buffer.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not delete the quota", func() {
				session := helpers.CFWithStdin(buffer, "delete-org-quota", quotaName)
				Eventually(session).Should(Say(`Organization quota '%s' has not been deleted\.`, quotaName))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org-quotas")
				Eventually(session).Should(Say(quotaName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user enters the default input (hits return)", func() {
			BeforeEach(func() {
				_, err := buffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not delete the quota", func() {
				session := helpers.CFWithStdin(buffer, "delete-org-quota", quotaName)
				Eventually(session).Should(Say(`Organization quota '%s' has not been deleted\.`, quotaName))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org-quotas")
				Eventually(session).Should(Say(quotaName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
