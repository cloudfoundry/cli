package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("table alignment", func() {
	BeforeEach(func() {
		helpers.LoginCF()
	})

	Context("when output is in English", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		// Developer note: The spacing in this test is significant and explicit. Do
		// not replace with a regex.
		It("aligns the table correctly", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("target")
			Eventually(session.Out).Should(Say("api endpoint:   %s", apiURL))
			Eventually(session.Out).Should(Say(`api version:    [\d.]+`))
			Eventually(session.Out).Should(Say("user:           %s", username))
			Eventually(session.Out).Should(Say("org:            %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("space:          %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when output is in language with multibyte characters", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		// Developer note: The spacing in this test is significant and explicit. Do
		// not replace with a regex.
		// TODO: add these strings to the i18n/resources
		// See: https://www.pivotaltracker.com/story/show/151737497
		XIt("aligns the table correctly", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CFWithEnv(map[string]string{"LANG": "ja-JP.utf8"}, "target")
			Eventually(session.Out).Should(Say("API エンドポイント:   %s", apiURL))
			Eventually(session.Out).Should(Say("api version:          [\\d.]+"))
			Eventually(session.Out).Should(Say("ユーザー:             %s", username))
			Eventually(session.Out).Should(Say("組織:                 %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("スペース:             %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})
})
