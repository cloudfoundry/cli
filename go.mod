module code.cloudfoundry.org/cli

go 1.22

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20230612151507-41ef4d1f67a4
	code.cloudfoundry.org/cfnetworking-cli-api v0.0.0-20190103195135-4b04f26287a6
	code.cloudfoundry.org/cli-plugin-repo v0.0.0-20200304195157-af98c4be9b85
	code.cloudfoundry.org/cli/integration/assets/hydrabroker v0.0.0-20201002233634-81722a1144e4
	code.cloudfoundry.org/clock v1.1.0
	code.cloudfoundry.org/diego-ssh v0.0.0-20230810200140-af9d79fe9c82
	code.cloudfoundry.org/go-log-cache/v2 v2.0.7
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f
	code.cloudfoundry.org/jsonry v1.1.4
	code.cloudfoundry.org/lager/v3 v3.0.3
	code.cloudfoundry.org/tlsconfig v0.0.0-20240510172918-c1e19801fe80
	code.cloudfoundry.org/ykk v0.0.0-20170424192843-e4df4ce2fd4d
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudfoundry/bosh-cli v6.4.1+incompatible
	github.com/creack/pty v1.1.21
	github.com/cyphar/filepath-securejoin v0.2.5
	github.com/distribution/reference v0.6.0
	github.com/fatih/color v1.17.0
	github.com/google/go-querystring v1.1.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/lunixbochs/vtclean v1.0.0
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-runewidth v0.0.15
	github.com/maxbrunsfeld/counterfeiter/v6 v6.8.1
	github.com/moby/term v0.5.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo/v2 v2.19.0
	github.com/onsi/gomega v1.33.1
	github.com/pkg/errors v0.9.1
	github.com/sabhiram/go-gitignore v0.0.0-20171017070213-362f9845770f
	github.com/sajari/fuzzy v1.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/tedsuo/rata v1.0.1-0.20170830210128-07d200713958
	github.com/vito/go-interact v0.0.0-20171111012221-fa338ed9e9ec
	golang.org/x/crypto v0.23.0
	golang.org/x/net v0.25.0
	golang.org/x/term v0.20.0
	golang.org/x/text v0.16.0
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.29.3
	k8s.io/client-go v0.29.3
)

require (
	code.cloudfoundry.org/inigo v0.0.0-20230612153013-b300679e6ed6 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/bmizerany/pat v0.0.0-20170815010413-6226ea591a40 // indirect
	github.com/charlievieth/fs v0.0.3 // indirect
	github.com/cloudfoundry/bosh-utils v0.0.397 // indirect
	github.com/cppforlife/go-patch v0.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.2 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/oauth2 v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
