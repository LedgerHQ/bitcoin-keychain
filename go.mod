module github.com/ledgerhq/bitcoin-keychain

go 1.15

require (
	github.com/go-redis/redis/v8 v8.2.3
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2
	github.com/kr/pretty v0.1.0 // indirect
	github.com/ledgerhq/bitcoin-keychain/pb v0.1.0
	github.com/magefile/mage v1.10.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/viper v1.3.2
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20201005172224-997123666555 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20201006033701-bcad7cf615f2 // indirect
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/ledgerhq/bitcoin-keychain/pb => ./pb
