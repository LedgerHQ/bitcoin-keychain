package main

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/go-redis/redis/v8"

	"github.com/ledgerhq/bitcoin-keychain/config"
	controllers "github.com/ledgerhq/bitcoin-keychain/grpc"
	"github.com/ledgerhq/bitcoin-keychain/log"
	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func serve(grpcAddr string, redisOpts *redis.Options) {
	conn, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"addr": grpcAddr,
		}).Fatal("cannot listen to address")
	}

	s := grpc.NewServer()

	keychainController, err := controllers.NewKeychainController(redisOpts)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to init controller")
	}

	pb.RegisterKeychainServiceServer(s, keychainController)

	healthCheckerController := controllers.NewHealthChecker()

	grpc_health_v1.RegisterHealthServer(s, healthCheckerController)

	reflection.Register(s)

	if err := s.Serve(conn); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("failed to serve")
	}
}

func main() {
	configProvider := config.LoadProvider("")

	var (
		host          string
		port          int32 = 50052
		redisHost     string
		redisPort     int32 = 6379
		redisDB       int   = 0
		redisPassword string
		tlsConfig     *tls.Config
	)

	host = configProvider.GetString("host")

	if val := configProvider.GetInt32("port"); val != 0 {
		port = val
	}

	grpcAddr := fmt.Sprintf("%s:%d", host, port)

	redisHost = configProvider.GetString("redis_host")

	if val := configProvider.GetInt32("redis_port"); val != 0 {
		redisPort = val
	}

	if val := configProvider.GetInt("redis_db"); val != 0 {
		redisDB = val
	}

	redisPassword = configProvider.GetString("redis_password")

	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	if configProvider.GetBool("redis_ssl") {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	serve(grpcAddr, &redis.Options{
		Addr:      redisAddr,
		Password:  redisPassword, // set password
		DB:        redisDB,       // use default DB
		TLSConfig: tlsConfig,
	})
}
