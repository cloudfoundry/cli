package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("KubernetesAuthActor", func() {
	var (
		k8sAuthActor    v7action.AuthActor
		k8sConfigGetter *v7actionfakes.FakeKubernetesConfigGetter
		config          *v7actionfakes.FakeConfig
		err             error
	)

	BeforeEach(func() {
		config = new(v7actionfakes.FakeConfig)
		k8sConfigGetter = new(v7actionfakes.FakeKubernetesConfigGetter)
		k8sConfigGetter.GetReturns(&clientcmdapi.Config{
			AuthInfos: map[string]*clientcmdapi.AuthInfo{"foo": {}, "bar": {}},
		}, nil)
		k8sAuthActor = v7action.NewKubernetesAuthActor(config, k8sConfigGetter)
	})

	Describe("Authenticate", func() {
		JustBeforeEach(func() {
			err = k8sAuthActor.Authenticate(map[string]string{"k8s-auth-info": "bar"}, "", constant.GrantTypePassword)
		})

		It("sets the Kubernetes auth-info", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(config.SetKubernetesAuthInfoCallCount()).To(Equal(1))
			Expect(config.SetKubernetesAuthInfoArgsForCall(0)).To(Equal("bar"))
		})
	})

	Describe("GetLoginPrompts", func() {
		var authPrompts map[string]coreconfig.AuthPrompt

		JustBeforeEach(func() {
			authPrompts, err = k8sAuthActor.GetLoginPrompts()
		})

		It("returns an auth prompt menu", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(authPrompts).To(HaveLen(1))
			Expect(authPrompts).To(HaveKey("k8s-auth-info"))

			authPrompt := authPrompts["k8s-auth-info"]
			Expect(authPrompt.Type).To(Equal(coreconfig.AuthPromptTypeMenu))
			Expect(authPrompt.DisplayName).To(Equal("Choose your Kubernetes authentication info"))
			Expect(authPrompt.Entries).To(ConsistOf("foo", "bar"))
		})

		It("sorts the entries", func() {
			authPrompt := authPrompts["k8s-auth-info"]
			Expect(authPrompt.Entries).To(Equal([]string{"bar", "foo"}))
		})

		When("getting the k8s config fails", func() {
			BeforeEach(func() {
				k8sConfigGetter.GetReturns(nil, errors.New("oomph!"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("oomph!"))
			})
		})

		When("no auth infos are in the k8s config", func() {
			BeforeEach(func() {
				k8sConfigGetter.GetReturns(&clientcmdapi.Config{}, nil)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("no kubernetes authentication infos configured"))
			})
		})
	})
})
