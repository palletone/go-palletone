module github.com/palletone/go-palletone

go 1.12

require (
	github.com/aristanetworks/goarista v0.0.0-20190712234253-ed1100a1c015
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/btcsuite/btcd v0.0.0-20190807005414-4063feeff79a
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cloudflare/cfssl v0.0.0-20190726000631-633726f6bcb7 // indirect
	github.com/containerd/continuity v0.0.0-20190426062206-aaeac12a7ffc // indirect
	github.com/coocood/freecache v0.0.0-20180304015925-036298587d3a
	github.com/coreos/etcd v3.3.10+incompatible // indirect
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/coreos/go-semver v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/docker v1.4.2-0.20190710153559-aa8249ae1b8b
	github.com/elastic/gosigar v0.10.4
	github.com/ethereum/go-ethereum v1.9.0
	github.com/fatih/color v1.6.0
	github.com/fjl/memsize v0.0.0-20180418122429-ca190fb6ffbc
	github.com/fsouza/go-dockerclient v1.4.2
	github.com/gizak/termui v0.0.0-20170117222342-991cd3d38091
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1
	github.com/golangci/golangci-lint v1.17.1 // indirect
	github.com/gomodule/redigo v0.0.0-20180627144507-2cd21d9966bf
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huin/goupnp v1.0.0
	github.com/influxdata/influxdb v0.0.0-20180221223340-01288bdb0883
	github.com/jackpal/go-nat-pmp v1.0.1
	github.com/julienschmidt/httprouter v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/looplab/fsm v0.0.0-20180515091235-f980bdb68a89
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/martinlindhe/base36 v0.0.0-20180729042928-5cda0030da17
	github.com/maruel/panicparse v0.0.0-20160720141634-ad661195ed0e // indirect
	github.com/maruel/ut v1.0.0 // indirect
	github.com/mattn/go-colorable v0.0.9
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/naoina/toml v0.0.0-20170410220130-ac014c6b6502
	github.com/nsf/termbox-go v0.0.0-20170211012700-3540b76b9c77 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/palletone/adaptor v0.6.1-0.20190823165629-94ce197415e9
	github.com/palletone/btc-adaptor v0.6.1-0.20190718062616-d31b41f123f7
	github.com/palletone/digital-identity v0.6.1-0.20190729063546-3dca665105bb
	github.com/palletone/eth-adaptor v0.6.1-0.20190823172410-3ebb6f741360
	github.com/pborman/uuid v1.2.0
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/peterh/liner v0.0.0-20170902204657-a37ad3984311
	github.com/pkg/errors v0.8.1
	github.com/prometheus/tsdb v0.10.0
	github.com/rjeczalik/notify v0.9.2
	github.com/robertkrimen/otto v0.0.0-20170205013659-6a77b7cbc37d
	github.com/rs/cors v1.7.0
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.0
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	go.dedis.ch/kyber/v3 v3.0.3
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v0.0.0-20180122172545-ddea229ff1df // indirect
	go.uber.org/zap v0.0.0-20180531205250-88c71ae3d702
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/sys v0.0.0-20190726091711-fc99dfbffb4e // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610 // indirect
	google.golang.org/grpc v1.22.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/errgo.v1 v1.0.1
	gopkg.in/karalabe/cookiejar.v2 v2.0.0-20150724131613-8dcd6a7f4951
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
)
