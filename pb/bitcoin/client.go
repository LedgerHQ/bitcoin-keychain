package bitcoin

import (
	"fmt"

	"github.com/ledgerhq/bitcoin-keychain-svc/config"
	"github.com/ledgerhq/bitcoin-keychain-svc/log"
	"google.golang.org/grpc"
)

// NewBitcoinSvcClient creates a new CoinService client by dialing the
// external bitcoin-svc gRPC service.
func NewBitcoinClient() CoinServiceClient {
	configProvider := config.LoadProvider("bitcoin")

	var (
		host string = ""
		port int32  = 50051
	)

	host = configProvider.GetString("host")

	if val := configProvider.GetInt32("port"); val != 0 {
		port = val
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return NewCoinServiceClient(conn)
}
