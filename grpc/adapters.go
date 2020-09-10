package grpc

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ledgerhq/bitcoin-keychain-svc/pb/v1"
	"github.com/ledgerhq/bitcoin-keychain-svc/pkg/keystore"
)

// KeychainInfo is an adapter function to convert a keystore.KeychainInfo
// instance to the corresponding protobuf message format.
func KeychainInfo(value keystore.KeychainInfo) *pb.KeychainInfo {
	var scheme pb.KeychainInfo_Scheme

	switch value.Scheme {
	case keystore.BIP44:
		scheme = pb.KeychainInfo_SCHEME_BIP44
	case keystore.BIP49:
		scheme = pb.KeychainInfo_SCHEME_BIP49
	case keystore.BIP84:
		scheme = pb.KeychainInfo_SCHEME_BIP84
	default:
		scheme = pb.KeychainInfo_SCHEME_UNSPECIFIED
	}

	return &pb.KeychainInfo{
		AccountDescriptor:         value.Descriptor,
		Xpub:                      value.XPub,
		Slip32ExtendedPublicKey:   value.SLIP32ExtendedPublicKey,
		ExternalXpub:              value.ExternalXPub,
		ExternalFreshAddressIndex: value.MaxConsecutiveExternalIndex,
		InternalXpub:              value.InternalXPub,
		InternalFreshAddressIndex: value.MaxConsecutiveInternalIndex,
		LookaheadSize:             value.LookaheadSize,
		Scheme:                    scheme,
	}
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
