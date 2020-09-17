package grpc

import (
	"context"

	pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"
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
	net, err := Network(request.Network)
	if err != nil {
		return nil, err
	}

	r, err := c.store.Create(request.AccountDescriptor, net)
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

func (c Controller) MarkPathAsUsed(
	ctx context.Context, request *pb.MarkPathAsUsedRequest,
) (*emptypb.Empty, error) {
	path, err := DerivationPath(request.Derivation)
	if err != nil {
		return nil, err
	}

	if err := c.store.MarkPathAsUsed(request.AccountDescriptor, path); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c Controller) GetFreshAddresses(
	ctx context.Context, request *pb.GetFreshAddressesRequest,
) (*pb.GetFreshAddressesResponse, error) {
	change, err := Change(request.Change)
	if err != nil {
		return nil, err
	}

	addrs, err := c.store.GetFreshAddresses(
		request.AccountDescriptor, change, request.BatchSize)
	if err != nil {
		return nil, err
	}

	return &pb.GetFreshAddressesResponse{Addresses: addrs}, nil
}

// NewKeychainController returns a new instance of a Controller struct that
// implements the pb.KeychainServiceServer interface.
func NewKeychainController() *Controller {
	return &Controller{
		store: keystore.NewInMemoryKeystore(),
	}
}
