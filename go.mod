module code.cloudfoundry.org/cli

go 1.16

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20170428003108-f4415fafc561
	code.cloudfoundry.org/cfnetworking-cli-api v0.0.0-20190103195135-4b04f26287a6
	code.cloudfoundry.org/cli-plugin-repo v0.0.0-20200304195157-af98c4be9b85
	code.cloudfoundry.org/cli/integration/assets/hydrabroker v0.0.0-20201002233634-81722a1144e4
	code.cloudfoundry.org/clock v1.0.0
	code.cloudfoundry.org/diego-ssh v0.0.0-20170109142818-18cdb3586e7f
	code.cloudfoundry.org/go-log-cache v1.0.1-0.20211011162012-ede82a99d3cc
	code.cloudfoundry.org/go-loggregator/v8 v8.0.5
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f
	code.cloudfoundry.org/jsonry v1.1.3
	code.cloudfoundry.org/lager v1.1.1-0.20191008172124-a9afc05ee5be
	code.cloudfoundry.org/tlsconfig v0.0.0-20210615191307-5d92ef3894a7
	code.cloudfoundry.org/ykk v0.0.0-20170424192843-e4df4ce2fd4d
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2
	github.com/blang/semver v3.5.1+incompatible
	github.com/bmatcuk/doublestar v1.3.1 // indirect
	github.com/charlievieth/fs v0.0.0-20170613215519-7dc373669fa1 // indirect
	github.com/cloudfoundry/bosh-cli v5.5.1+incompatible
	github.com/cloudfoundry/bosh-utils v0.0.0-20180315210917-c6a922e299b8 // indirect
	github.com/cppforlife/go-patch v0.1.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.1
	github.com/docker/distribution v2.6.0-rc.1.0.20171109224904-e5b5e44386f7+incompatible
	github.com/docker/docker v1.4.2-0.20171120205147-9de84a78d76e // indirect
	github.com/fatih/color v1.5.1-0.20170926111411-5df930a27be2
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
	github.com/kr/pty v1.1.1
	github.com/lunixbochs/vtclean v1.0.0
	github.com/mattn/go-colorable v0.1.0
	github.com/mattn/go-isatty v0.0.3 // indirect
	github.com/mattn/go-runewidth v0.0.5-0.20181218000649-703b5e6b11ae
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.2
	github.com/moby/moby v1.4.2-0.20171120205147-9de84a78d76e
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sabhiram/go-gitignore v0.0.0-20171017070213-362f9845770f
	github.com/sajari/fuzzy v1.0.0
	github.com/sirupsen/logrus v1.2.0
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	github.com/tedsuo/rata v1.0.1-0.20170830210128-07d200713958
	github.com/vito/go-interact v0.0.0-20171111012221-fa338ed9e9ec
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/net v0.0.0-20211013171255-e13a2654a71e
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	golang.org/x/text v0.3.7
	golang.org/x/tools v0.1.6-0.20210908190839-cf92b39a962c // indirect
	google.golang.org/genproto v0.0.0-20211013025323-ce878158c4d4 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/client-go v0.22.2
)

replace gopkg.in/fsnotify.v1 v1.4.7 => github.com/fsnotify/fsnotify v1.4.7
