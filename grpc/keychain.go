package grpc

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"github.com/ledgerhq/bitcoin-keychain/log"

	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
	"github.com/ledgerhq/bitcoin-keychain/pkg/keystore"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Controller is a type that implements the pb.KeychainServiceServer
// interface.
type Controller struct{}

var store *keystore.RedisKeystore

func (c Controller) CreateKeychain(
	ctx context.Context, request *pb.CreateKeychainRequest,
) (*pb.KeychainInfo, error) {
	net, err := Network(request.Network)
	if err != nil {
		return nil, err
	}

	scheme, err := Scheme(request.Scheme)
	if err != nil {
		return nil, err
	}

	lookaheadSize := uint32(keystore.DefaultLookaheadSize) // default lookahead size
	if s := request.GetLookaheadSize(); s != 0 {
		lookaheadSize = s
	}

	extendedKey := request.GetExtendedPublicKey()
	fromChainCode := FromChainCode(request.GetFromChainCode())

	r, err := store.Create(
		extendedKey, fromChainCode, scheme, net, lookaheadSize)
	if err != nil {
		return nil, err
	}

	return KeychainInfo(r)
}

func (c Controller) DeleteKeychain(
	ctx context.Context, request *pb.DeleteKeychainRequest,
) (*emptypb.Empty, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, store.Delete(id)
}

func (c Controller) GetKeychainInfo(
	ctx context.Context, request *pb.GetKeychainInfoRequest,
) (*pb.KeychainInfo, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		return nil, err
	}

	r, err := store.Get(id)
	if err != nil {
		return nil, err
	}

	return KeychainInfo(r)
}

func (c Controller) GetFreshAddresses(
	ctx context.Context, request *pb.GetFreshAddressesRequest,
) (*pb.GetFreshAddressesResponse, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		return nil, err
	}

	change, err := Change(request.Change)
	if err != nil {
		return nil, err
	}

	addrs, err := store.GetFreshAddresses(id, change, request.BatchSize)
	if err != nil {
		return nil, err
	}

	var addrInfoList []*pb.AddressInfo

	for _, addrInfo := range addrs {
		addrInfoProto, err := AddressInfoProto(addrInfo)
		if err != nil {
			log.WithFields(log.Fields{
				"id":    id.String(),
				"addr":  addrInfo.Address,
				"error": err,
			}).Error("[grpc] GetAllObservableAddresses: invalid AddressInfo")

			return nil, err
		}

		addrInfoList = append(addrInfoList, addrInfoProto)
	}

	return &pb.GetFreshAddressesResponse{Addresses: addrInfoList}, nil
}

func (c Controller) MarkAddressesAsUsed(
	ctx context.Context, request *pb.MarkAddressesAsUsedRequest,
) (*emptypb.Empty, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		log.WithFields(log.Fields{
			"id":    request.KeychainId,
			"error": err,
		}).Error("[grpc] MarkAddressesAsUsed: invalid KeychainID")

		return nil, err
	}

	for _, addr := range request.Addresses {
		if err := store.MarkAddressAsUsed(id, addr); err != nil {
			log.WithFields(log.Fields{
				"id":    id.String(),
				"addr":  addr,
				"error": err,
			}).Error("[grpc] MarkAddressesAsUsed: failed")

			return nil, err
		}
	}

	log.WithFields(log.Fields{
		"id":    id.String(),
		"addrs": request.Addresses,
	}).Info("[grpc] MarkAddressesAsUsed: successful")

	return &emptypb.Empty{}, nil
}

func (c Controller) GetAllObservableAddresses(
	ctx context.Context, request *pb.GetAllObservableAddressesRequest,
) (*pb.GetAllObservableAddressesResponse, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		log.WithFields(log.Fields{
			"id":    request.KeychainId,
			"error": err,
		}).Error("[grpc] GetAllObservableAddresses: invalid KeychainID")

		return nil, err
	}

	var changeList []keystore.Change
	if request.Change == pb.Change_CHANGE_UNSPECIFIED {
		changeList = []keystore.Change{keystore.External, keystore.Internal}
	} else {
		change, err := Change(request.Change)
		if err != nil {
			log.WithFields(log.Fields{
				"id":     id.String(),
				"change": request.Change.String(),
				"error":  err,
			}).Error("[grpc] GetAllObservableAddresses: invalid Change")

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

	var addrs []keystore.AddressInfo

	for _, change := range changeList {
		log.WithFields(log.Fields{
			"id":     id,
			"change": change,
			"range":  []uint32{request.FromIndex, to},
		}).Info("[grpc] GetAllObservableAddresses: get from keystore")

		changeAddrs, err := store.GetAllObservableAddresses(
			id, change, request.FromIndex, to)
		if err != nil {
			log.WithFields(log.Fields{
				"id":     id.String(),
				"change": request.Change.String(),
				"error":  err,
			}).Error("[grpc] GetAllObservableAddresses: failed to fetch from keystore")

			return nil, err
		}

		addrs = append(addrs, changeAddrs...)
	}

	var addrInfoList []*pb.AddressInfo

	for _, addrInfo := range addrs {
		addrInfoProto, err := AddressInfoProto(addrInfo)
		if err != nil {
			log.WithFields(log.Fields{
				"id":    id.String(),
				"addr":  addrInfo.Address,
				"error": err,
			}).Error("[grpc] GetAllObservableAddresses: invalid AddressInfo")

			return nil, err
		}

		addrInfoList = append(addrInfoList, addrInfoProto)
	}

	log.WithFields(log.Fields{
		"id":     id.String(),
		"num":    len(addrInfoList),
		"change": changeList,
	}).Info("[grpc] GetAllObservableAddresses: successful")

	return &pb.GetAllObservableAddressesResponse{Addresses: addrInfoList}, nil
}

func (c Controller) GetAddressesPublicKeys(
	ctx context.Context, request *pb.GetAddressesPublicKeysRequest,
) (*pb.GetAddressesPublicKeysResponse, error) {
	id, err := KeychainID(request.KeychainId)
	if err != nil {
		log.WithFields(log.Fields{
			"id":    request.KeychainId,
			"error": err,
		}).Error("[grpc] GetAddressesPublicKeys: invalid KeychainID")

		return nil, err
	}

	derivations := make([]keystore.DerivationPath, len(request.Derivations))

	for idx, path := range request.Derivations {
		derivationPath, err := DerivationPath(path.Derivation)

		if err != nil {
			log.WithFields(log.Fields{
				"id":    request.KeychainId,
				"error": err,
			}).Error("[grpc] GetAddressesPublicKeys: invalid derivation path from request")

			return nil, err
		}

		derivations[idx] = derivationPath
	}

	publicKeys, err := store.GetAddressesPublicKeys(id, derivations)
	if err != nil {
		log.WithFields(log.Fields{
			"id":    request.KeychainId,
			"error": err,
		}).Error("[grpc] GetAddressesPublicKeys: failed to fetch from keystore")

		return nil, err
	}

	response := &pb.GetAddressesPublicKeysResponse{PublicKeys: publicKeys}

	log.WithFields(log.Fields{
		"id":          id.String(),
		"derivations": request.Derivations,
		"publicKeys":  publicKeys,
	}).Info("[grpc] GetAddressesPublicKeys: successful")

	return response, nil
}

// NewKeychainController returns a new instance of a Controller struct that
// implements the pb.KeychainServiceServer interface.
func NewKeychainController(redisOpts *redis.Options) (*Controller, error) {
	var err error

	store, err = keystore.NewRedisKeystore(redisOpts)
	if err != nil {
		return nil, fmt.Errorf("Creating redis client failed: %w", err)
	}

	return &Controller{}, nil
}
