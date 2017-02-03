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

		It("aligns the table correctly", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("target")
			Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
			Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
			Eventually(session.Out).Should(Say("User:           %s", username))
			Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when output is in language with multibyte characters", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		It("aligns the table correctly", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CFWithEnv(map[string]string{"LANG": "ja-JP.utf8"}, "target")
			Eventually(session.Out).Should(Say("API エンドポイント:   %s", apiURL))
			Eventually(session.Out).Should(Say(`API version:          [\d.]+`))
			Eventually(session.Out).Should(Say("ユーザー:             %s", username))
			Eventually(session.Out).Should(Say("組織:                 %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("スペース:             %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})
})
