package integration

import (
	"context"
	"net"

	"github.com/go-redis/redis/v8"

	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"

	controllers "github.com/ledgerhq/bitcoin-keychain/grpc"
	"github.com/ledgerhq/bitcoin-keychain/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// launch at package initialization
func init() {
	startKeychain()
}

// startKeychain launches a keychain gRPC server over a buffered connection.
// This allows us to use a full-blown server for tests, without reserving a
// TCP port for the same.
//
// gRPC client can use the same connection to dial to the server.
func startKeychain() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	keychainController, err := controllers.NewKeychainController(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to init controller")
	}

	pb.RegisterKeychainServiceServer(s, keychainController)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("failed to serve")
		}
	}()
}

// connect to buffered connection.
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// keychainClient connects to the keychain gRPC service via a buffered
// connection.
func keychainClient(ctx context.Context) (pb.KeychainServiceClient, *grpc.ClientConn) {
	conn, err := grpc.DialContext(
		ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to dial bufnet")
	}

	client := pb.NewKeychainServiceClient(conn)

	return client, conn
}
