package grpc

import (
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
	"github.com/ledgerhq/bitcoin-keychain/pkg/keystore"
	"github.com/pkg/errors"
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

	chainParams, err := ChainParams(value.Network)
	if err != nil {
		return nil, err
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
		ChainParams:             chainParams,
	}, nil
}

// Network is an adapter function to convert a gRPC pb.BitcoinNetwork
// to keystore.Network instance.
func Network(params *pb.ChainParams) (chaincfg.Network, error) {
	switch net := params.GetBitcoinNetwork(); net {
	case pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET:
		return chaincfg.BitcoinMainnet, nil
	case pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3:
		return chaincfg.BitcoinTestnet3, nil
	case pb.BitcoinNetwork_BITCOIN_NETWORK_REGTEST:
		return chaincfg.BitcoinRegtest, nil
	}

	switch net := params.GetLitecoinNetwork(); net {
	case pb.LitecoinNetwork_LITECOIN_NETWORK_MAINNET:
		return chaincfg.LitecoinMainnet, nil
	default:
		return "", errors.Wrap(ErrUnrecognizedNetwork, fmt.Sprint(net))
	}
}

// ChainParams is a helper to convert a Network in keystore package to
// the corresponding gRPC *pb.ChainParams value.
func ChainParams(net chaincfg.Network) (*pb.ChainParams, error) {
	switch net {
	case chaincfg.BitcoinMainnet:
		return &pb.ChainParams{
			Network: &pb.ChainParams_BitcoinNetwork{
				BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_MAINNET,
			},
		}, nil
	case chaincfg.BitcoinTestnet3:
		return &pb.ChainParams{
			Network: &pb.ChainParams_BitcoinNetwork{
				BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3,
			},
		}, nil
	case chaincfg.BitcoinRegtest:
		return &pb.ChainParams{
			Network: &pb.ChainParams_BitcoinNetwork{
				BitcoinNetwork: pb.BitcoinNetwork_BITCOIN_NETWORK_REGTEST,
			},
		}, nil
	case chaincfg.LitecoinMainnet:
		return &pb.ChainParams{
			Network: &pb.ChainParams_LitecoinNetwork{
				LitecoinNetwork: pb.LitecoinNetwork_LITECOIN_NETWORK_MAINNET,
			},
		}, nil
	default:
		return nil, errors.Wrap(ErrUnrecognizedNetwork, fmt.Sprint(net))
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
