package integration

import (
	"context"
	"net"

	pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"

	controllers "github.com/ledgerhq/bitcoin-keychain-svc/grpc"
	"github.com/ledgerhq/bitcoin-keychain-svc/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// launch at package initialization
func init() {
	startKeychainSvc()
}

// startKeychainSvc launches a keychain gRPC server over a buffered connection.
// This allows us to use a full-blown server for tests, without reserving a
// TCP port for the same.
//
// gRPC client can use the same connection to dial to the server.
func startKeychainSvc() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	keychainController := controllers.NewKeychainController()
	pb.RegisterKeychainServiceServer(s, keychainController)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

// connect to buffered connection.
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// keychainSvcClient connects to the keychain gRPC service via a buffered
// connection.
func keychainSvcClient(ctx context.Context) (pb.KeychainServiceClient, *grpc.ClientConn) {
	conn, err := grpc.DialContext(
		ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewKeychainServiceClient(conn)

	return client, conn
}
