package v7action

import (
	"errors"
	"sort"

	"code.cloudfoundry.org/cli/api/uaa/constant"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KubernetesConfigGetter

type KubernetesConfigGetter interface {
	Get() (*clientcmdapi.Config, error)
}

type DefaultKubernetesConfigGetter struct{}

func NewDefaultKubernetesConfigGetter() DefaultKubernetesConfigGetter {
	return DefaultKubernetesConfigGetter{}
}

func (c DefaultKubernetesConfigGetter) Get() (*clientcmdapi.Config, error) {
	pathOpts := clientcmd.NewDefaultPathOptions()
	return pathOpts.GetStartingConfig()
}

type kubernetesAuthActor struct {
	config          Config
	k8sConfigGetter KubernetesConfigGetter
}

func NewKubernetesAuthActor(config Config, k8sConfigGetter KubernetesConfigGetter) AuthActor {
	return &kubernetesAuthActor{
		config:          config,
		k8sConfigGetter: k8sConfigGetter,
	}
}

func (actor kubernetesAuthActor) Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error {
	actor.config.SetKubernetesAuthInfo(credentials["k8s-auth-info"])
	return nil
}

func (actor kubernetesAuthActor) GetLoginPrompts() (map[string]coreconfig.AuthPrompt, error) {
	conf, err := actor.k8sConfigGetter.Get()
	if err != nil {
		return nil, err
	}

	if len(conf.AuthInfos) == 0 {
		return nil, errors.New("no kubernetes authentication infos configured")
	}

	var prompts []string
	for authInfo := range conf.AuthInfos {
		prompts = append(prompts, authInfo)
	}
	sort.Strings(prompts)

	return map[string]coreconfig.AuthPrompt{"k8s-auth-info": {
		Type:        coreconfig.AuthPromptTypeMenu,
		Entries:     prompts,
		DisplayName: "Choose your Kubernetes authentication info",
	}}, nil
}
