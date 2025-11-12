package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("KubernetesAuthActor", func() {
	var (
		k8sAuthActor    v7action.AuthActor
		k8sConfigGetter *v7actionfakes.FakeKubernetesConfigGetter
		whoAmIer        *v7actionfakes.FakeWhoAmIer
		config          *v7actionfakes.FakeConfig
		err             error
	)

	BeforeEach(func() {
		config = new(v7actionfakes.FakeConfig)
		k8sConfigGetter = new(v7actionfakes.FakeKubernetesConfigGetter)
		k8sConfigGetter.GetReturns(&clientcmdapi.Config{
			AuthInfos: map[string]*clientcmdapi.AuthInfo{"foo": {}, "bar": {}},
		}, nil)
		whoAmIer = new(v7actionfakes.FakeWhoAmIer)
		k8sAuthActor = v7action.NewKubernetesAuthActor(config, k8sConfigGetter, whoAmIer)
	})

	Describe("Authenticate", func() {
		var username string
		BeforeEach(func() {
			username = "bar"
		})

		JustBeforeEach(func() {
			err = k8sAuthActor.Authenticate(map[string]string{"username": username}, "", constant.GrantTypePassword)
		})

		It("sets the Kubernetes auth-info", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(config.SetKubernetesAuthInfoCallCount()).To(Equal(1))
			Expect(config.SetKubernetesAuthInfoArgsForCall(0)).To(Equal("bar"))
		})

		When("the given username is not in the k8s config", func() {
			BeforeEach(func() {
				username = "no-such-person"
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("kubernetes user not found in configuration: " + username))
			})
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

	Describe("GetLoginPrompts", func() {
		var authPrompts map[string]coreconfig.AuthPrompt

		JustBeforeEach(func() {
			authPrompts, err = k8sAuthActor.GetLoginPrompts()
		})

		It("returns an auth prompt menu", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(authPrompts).To(HaveLen(1))
			Expect(authPrompts).To(HaveKey("username"))

			authPrompt := authPrompts["username"]
			Expect(authPrompt.Type).To(Equal(coreconfig.AuthPromptTypeMenu))
			Expect(authPrompt.DisplayName).To(Equal("Choose your Kubernetes authentication info"))
			Expect(authPrompt.Entries).To(ConsistOf("foo", "bar"))
		})

		It("sorts the entries", func() {
			authPrompt := authPrompts["username"]
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

	Describe("Get Current User", func() {
		var (
			user configv3.User
			err  error
		)

		BeforeEach(func() {
			whoAmIer.WhoAmIReturns(resources.K8sUser{Name: "bob", Kind: "User"}, nil, nil)
		})

		JustBeforeEach(func() {
			user, err = k8sAuthActor.GetCurrentUser()
		})

		It("uses the WhoAmI function to get the real current user name", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("bob"))
		})

		When("calling the whoami endpoint fails", func() {
			BeforeEach(func() {
				whoAmIer.WhoAmIReturns(resources.K8sUser{}, nil, errors.New("boom!"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("boom!")))
			})
		})
	})
})
