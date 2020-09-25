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

func (c Controller) MarkAddressesAsUsed(
	ctx context.Context, request *pb.MarkAddressesAsUsedRequest,
) (*emptypb.Empty, error) {
	for _, addr := range request.Addresses {
		if err := c.store.MarkAddressAsUsed(request.AccountDescriptor, addr); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func (c Controller) GetAllObservableAddresses(
	ctx context.Context, request *pb.GetAllObservableAddressesRequest,
) (*pb.GetAllObservableAddressesResponse, error) {
	var changeList []keystore.Change
	if request.Change == pb.Change_CHANGE_UNSPECIFIED {
		changeList = []keystore.Change{keystore.External, keystore.Internal}
	} else {
		change, err := Change(request.Change)
		if err != nil {
			return nil, err
		}

		changeList = []keystore.Change{change}
	}

	// If the toIndex field is left out in the request payload, we substitute
	// it with a large value so that the max observable range is used instead.
	var to uint32
	if request.GetToIndex() == 0 {
		to = (1 << 31) - 1 // uint32 max
	} else {
		to = request.ToIndex
	}

	var addrs []string

	for _, change := range changeList {
		changeAddrs, err := c.store.GetAllObservableAddresses(
			request.AccountDescriptor, change, request.FromIndex, to)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, changeAddrs...)
	}

	return &pb.GetAllObservableAddressesResponse{Addresses: addrs}, nil
}

// NewKeychainController returns a new instance of a Controller struct that
// implements the pb.KeychainServiceServer interface.
func NewKeychainController() *Controller {
	return &Controller{
		store: keystore.NewInMemoryKeystore(),
	}
}
