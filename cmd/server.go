package main

import (
	"fmt"
	"net"

	"github.com/ledgerhq/bitcoin-keychain-svc/config"
	"google.golang.org/grpc/reflection"

	controllers "github.com/ledgerhq/bitcoin-keychain-svc/grpc"
	"github.com/ledgerhq/bitcoin-keychain-svc/log"
	pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"
	"google.golang.org/grpc"
)

func serve(addr string) {
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Cannot listen to address %s", addr)
	}

	s := grpc.NewServer()
	keychainController := controllers.NewKeychainController()
	pb.RegisterKeychainServiceServer(s, keychainController)

	reflection.Register(s)

	if err := s.Serve(conn); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	configProvider := config.LoadProvider("bitcoin_keychain")

	var (
		host string
		port int32 = 50052
	)

	host = configProvider.GetString("host")

	if val := configProvider.GetInt32("port"); val != 0 {
		port = val
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	serve(addr)
}
