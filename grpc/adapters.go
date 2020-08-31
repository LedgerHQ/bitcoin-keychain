package grpc

import (
	"github.com/ledgerhq/bitcoin-keychain-svc/pb/v1"
	"github.com/ledgerhq/bitcoin-keychain-svc/pkg/keystore"
)

// KeychainInfo is an adapter function to convert a keystore.KeychainInfo
// instance to the corresponding protobuf message format.
func KeychainInfo(value keystore.KeychainInfo) *pb.GetKeychainInfoResponse {
	var scheme pb.GetKeychainInfoResponse_Scheme
	switch value.Scheme {
	case keystore.BIP44:
		scheme = pb.GetKeychainInfoResponse_SCHEME_BIP44
	case keystore.BIP49:
		scheme = pb.GetKeychainInfoResponse_SCHEME_BIP49
	case keystore.BIP84:
		scheme = pb.GetKeychainInfoResponse_SCHEME_BIP84
	default:
		scheme = pb.GetKeychainInfoResponse_SCHEME_UNSPECIFIED
	}

	return &pb.GetKeychainInfoResponse{
		AccountDescriptor:       value.Descriptor,
		Xpub:                    value.XPub,
		Slip32ExtendedPublicKey: value.SLIP32ExtendedPublicKey,
		ExternalPublicKey:       value.ExternalPublicKey,
		ExternalChainCode:       value.ExternalChainCode,
		MaxExternalIndex:        value.MaxExternalIndex,
		InternalPublicKey:       value.InternalPublicKey,
		InternalChainCode:       value.ExternalChainCode,
		MaxInternalIndex:        value.MaxInternalIndex,
		LookaheadSize:           value.LookaheadSize,
		Scheme:                  scheme,
	}
}
