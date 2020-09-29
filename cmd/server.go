package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc/reflection"

	controllers "github.com/ledgerhq/bitcoin-keychain-svc/grpc"
	"github.com/ledgerhq/bitcoin-keychain-svc/log"
	pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"
	"google.golang.org/grpc"
)

func serve() {
	addr := fmt.Sprintf(":%d", 50052)

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
	serve()
}
