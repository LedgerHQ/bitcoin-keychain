module github.com/ledgerhq/bitcoin-keychain-svc

go 1.15

require (
	github.com/golang/protobuf v1.4.1
	github.com/google/uuid v1.1.2
	github.com/ledgerhq/bitcoin-keychain-svc/pb v0.1.0
	github.com/magefile/mage v1.10.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/viper v1.3.2
	google.golang.org/genproto v0.0.0-20200715011427-11fb19a81f2c // indirect
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
)

replace github.com/ledgerhq/bitcoin-keychain-svc/pb => ./pb
