package grpc

import (
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/pkg/errors"

	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
	"github.com/ledgerhq/bitcoin-keychain/pkg/keystore"
)

// KeychainInfo is an adapter function to convert a keystore.KeychainInfo
// instance to the corresponding protobuf message format.
func KeychainInfo(value keystore.KeychainInfo) (*pb.KeychainInfo, error) {
	var scheme pb.Scheme

	switch value.Scheme {
	case keystore.BIP44:
		scheme = pb.Scheme_SCHEME_BIP44
	case keystore.BIP49:
		scheme = pb.Scheme_SCHEME_BIP49
	case keystore.BIP84:
		scheme = pb.Scheme_SCHEME_BIP84
	default:
		scheme = pb.Scheme_SCHEME_UNSPECIFIED
	}

	var network pb.BitcoinNetwork

	switch value.Network {
	case keystore.Mainnet:
		network = pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET
	case keystore.Testnet3:
		network = pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3
	case keystore.Regtest:
		network = pb.BitcoinNetwork_BITCOIN_NETWORK_REGTEST
	default:
		network = pb.BitcoinNetwork_BITCOIN_NETWORK_UNSPECIFIED
	}

	id, err := value.ID.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(
			ErrInvalidKeychainID, fmt.Sprintf("%v", value.ID))
	}

	return &pb.KeychainInfo{
		KeychainId:              id,
		ExternalDescriptor:      value.ExternalDescriptor,
		InternalDescriptor:      value.InternalDescriptor,
		ExtendedPublicKey:       value.ExtendedPublicKey,
		Slip32ExtendedPublicKey: value.SLIP32ExtendedPublicKey,
		LookaheadSize:           value.LookaheadSize,
		Scheme:                  scheme,
		Network:                 network,
	}, nil
}

// Network is an adapter function to convert a gRPC pb.BitcoinNetwork
// to keystore.Network instance.
func Network(params pb.BitcoinNetwork) (keystore.Network, error) {
	switch params {
	case pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET:
		return keystore.Mainnet, nil
	case pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3:
		return keystore.Testnet3, nil
	case pb.BitcoinNetwork_BITCOIN_NETWORK_REGTEST:
		return keystore.Regtest, nil
	default:
		return "", errors.Wrap(ErrUnrecognizedNetwork, fmt.Sprint(params))
	}
}

// DerivationPath is an adapter function to convert a derivation path (slice)
// to a keystore.DerivationPath instance.
func DerivationPath(path []uint32) (keystore.DerivationPath, error) {
	if len(path) != 2 {
		return keystore.DerivationPath{}, errors.Wrap(
			ErrInvalidDerivationPath, fmt.Sprintf("%v", path))
	}

	return keystore.DerivationPath{path[0], path[1]}, nil
}

// Change is an adapter function to convert a gRPC pb.Change to an instance of
// keystore.Change.
func Change(change pb.Change) (keystore.Change, error) {
	switch change {
	case pb.Change_CHANGE_EXTERNAL:
		return keystore.External, nil
	case pb.Change_CHANGE_INTERNAL:
		return keystore.Internal, nil
	default:
		return -1, errors.Wrap(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

// KeychainID is an adapter function to convert raw bytes to a uuid.UUID
// instance.
func KeychainID(id []byte) (uuid.UUID, error) {
	keychainID, err := uuid.FromBytes(id)
	if err != nil {
		return [16]byte{}, errors.Wrap(
			ErrInvalidKeychainID, hex.EncodeToString(id))
	}

	return keychainID, nil
}

// Scheme is an adapter function to convert pb.Scheme to a keystore.Scheme
// instance.
func Scheme(scheme pb.Scheme) (keystore.Scheme, error) {
	switch scheme {
	case pb.Scheme_SCHEME_BIP44:
		return keystore.BIP44, nil
	case pb.Scheme_SCHEME_BIP49:
		return keystore.BIP49, nil
	case pb.Scheme_SCHEME_BIP84:
		return keystore.BIP84, nil
	default:
		return "", ErrUnrecognizedScheme
	}
}

// AddressInfoProto is an adapter function to convert a keystore.AddressInfo
// to a pb.AddressInfo object.
func AddressInfoProto(info keystore.AddressInfo) (*pb.AddressInfo, error) {
	var change pb.Change

	switch info.Change {
	case keystore.External:
		change = pb.Change_CHANGE_EXTERNAL
	case keystore.Internal:
		change = pb.Change_CHANGE_INTERNAL
	default:
		return nil, errors.Wrap(ErrUnrecognizedChange, fmt.Sprint(change))
	}

	return &pb.AddressInfo{
		Address:    info.Address,
		Derivation: info.Derivation.ToSlice(),
		Change:     change,
	}, nil
}

// FromChainCode is an adapter function to convert pb.FromChainCode to a keystore.FromChainCode instance.
func FromChainCode(chainCodeInfo *pb.FromChainCode) *keystore.FromChainCode {
	if chainCodeInfo != nil {
		return &keystore.FromChainCode{
			PublicKey:    chainCodeInfo.PublicKey,
			ChainCode:    chainCodeInfo.ChainCode,
			AccountIndex: chainCodeInfo.AccountIndex,
		}
	}

	return nil
}
