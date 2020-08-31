package grpc

import (
	"context"

	"github.com/ledgerhq/bitcoin-keychain-svc/pb/v1"
	"github.com/ledgerhq/bitcoin-keychain-svc/pkg/keystore"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Controller is a type that implements the pb.KeychainServiceServer
// interface.
type Controller struct {
	store keystore.Keystore
}

func (c Controller) CreateKeychain(
	ctx context.Context, request *pb.CreateKeychainRequest,
) (*pb.KeychainInfo, error) {
	r, err := c.store.Create(request.AccountDescriptor)
	if err != nil {
		return nil, err
	}

	return KeychainInfo(r), nil
}

func (c Controller) DeleteKeychain(
	ctx context.Context, request *pb.DeleteKeychainRequest,
) (*emptypb.Empty, error) {
	panic("implement me")
}

func (c Controller) GetKeychainInfo(
	ctx context.Context, request *pb.GetKeychainInfoRequest,
) (*pb.KeychainInfo, error) {
	r, err := c.store.Get(request.AccountDescriptor)
	if err != nil {
		return nil, err
	}

	return KeychainInfo(r), nil
}

// NewKeychainController returns a new instance of a Controller struct that
// implements the pb.KeychainServiceServer interface.
func NewKeychainController() *Controller {
	return &Controller{
		store: keystore.NewInMemoryKeystore(),
	}
}
