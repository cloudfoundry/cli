module code.cloudfoundry.org/cli

go 1.20

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
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2
	github.com/blang/semver v3.5.1+incompatible
	github.com/cloudfoundry/bosh-cli v5.5.1+incompatible
	github.com/cloudfoundry/noaa/v2 v2.2.0
	github.com/cloudfoundry/sonde-go v0.0.0-20220627221915-ff36de9c3435
	github.com/cyphar/filepath-securejoin v0.2.1
	github.com/docker/distribution v2.8.2+incompatible
	github.com/fatih/color v1.15.0
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
	github.com/kr/pty v1.1.8
	github.com/lunixbochs/vtclean v1.0.0
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-runewidth v0.0.5-0.20181218000649-703b5e6b11ae
	github.com/maxbrunsfeld/counterfeiter/v6 v6.7.0
	github.com/moby/term v0.0.0-20221120202655-abb19827d345
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.10
	github.com/pkg/errors v0.9.1
	github.com/sabhiram/go-gitignore v0.0.0-20171017070213-362f9845770f
	github.com/sajari/fuzzy v1.0.0
	github.com/sirupsen/logrus v1.2.0
	github.com/tedsuo/rata v1.0.1-0.20170830210128-07d200713958
	github.com/vito/go-interact v0.0.0-20171111012221-fa338ed9e9ec
	golang.org/x/crypto v0.12.0
	golang.org/x/net v0.14.0
	golang.org/x/text v0.12.0
	gopkg.in/cheggaaa/pb.v1 v1.0.27
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/bmatcuk/doublestar v1.3.1 // indirect
	github.com/bmizerany/pat v0.0.0-20170815010413-6226ea591a40 // indirect
	github.com/charlievieth/fs v0.0.0-20170613215519-7dc373669fa1 // indirect
	github.com/cloudfoundry/bosh-utils v0.0.0-20180315210917-c6a922e299b8 // indirect
	github.com/cppforlife/go-patch v0.1.0 // indirect
	github.com/creack/pty v1.1.11 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/term v0.11.0 // indirect
	golang.org/x/tools v0.12.0 // indirect
	google.golang.org/genproto v0.0.0-20230526161137-0005af68ea54 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230525234035-dd9d682886f9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/fsnotify.v1 v1.4.7 => github.com/fsnotify/fsnotify v1.4.7
