package bitcoin

import (
	"github.com/ledgerhq/bitcoin-keychain-svc/log"
	"google.golang.org/grpc"
)

// NewBitcoinSvcClient creates a new CoinService client by dialing the
// external bitcoin-svc gRPC service.
func NewBitcoinClient() CoinServiceClient {
	// TODO: use env vars (via config package) here.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return NewCoinServiceClient(conn)
}
