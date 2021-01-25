package keystore

import (
	"context"
	"fmt"

	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
)

// DerivationPath represents the BIP32 derivation path as an array, relative
// to the BIP32 path-level 3 (account level).
//
// Hardened indexes are NOT supported, which is enforced by the uint32 type.
//
//   ┌──────────────┬───────────────┐
//   │ change index │ address index │
//   │ 4 bytes      │ 4 bytes       │
//   └──────────────┴───────────────┘
//
// For example, if the full derivation path is m/44'/0'/0'/1/2, the
// representation in DerivationPath would be DerivationPath{1, 2}.
type DerivationPath [2]uint32

func (path DerivationPath) MarshalText() (text []byte, err error) {
	p := fmt.Sprintf("%d/%d", path.ChangeIndex(), path.AddressIndex())
	return []byte(p), nil
}

// ChangeIndex returns the change index at BIP32 path-level 4, from a given
// DerivationPath.
func (path DerivationPath) ChangeIndex() Change {
	if path[0] == 0 {
		return External
	}

	return Internal // path[0] is 1
}

// ChangeIndex returns the address index at BIP32 path-level 5, from a given
// DerivationPath.
func (path DerivationPath) AddressIndex() uint32 {
	return path[1]
}

// ToSlice returns the raw derivation path as a uint32 slice. The derivation
// is relative to the BIP-32 account-level.
func (path DerivationPath) ToSlice() []uint32 {
	return []uint32{path[0], path[1]}
}

// Change is an enum type to indicate whether an address belongs to the
// external chain (receive) or the internal chain (change).
//
//   m / purpose' / coin_type' / account' / change / address_index
//                       This is Change ------^
//
// Change values should never be BIP32 hardened.
type Change int

const (
	// External indicates that an address belongs to the external chain.
	// The value of the enum indicates the value in BIP32 path-level 4.
	External Change = iota

	// Internal indicates that an address belongs to the internal chain.
	// The value of the enum indicates the value in BIP32 path-level 4.
	Internal
)

// childKDF is a helper to derive an child extended public key on a child
// index, from a parent extended public key.
//
// This helper can only be used to derive one BIP32 level at a time.
func childKDF(client bitcoin.CoinServiceClient, xPub string, childIndex uint32) (*bitcoin.DeriveExtendedKeyResponse, error) {
	return client.DeriveExtendedKey(
		context.Background(), &bitcoin.DeriveExtendedKeyRequest{
			ExtendedKey: xPub,
			Derivation:  []uint32{childIndex},
		})
}

// GetAccountExtendedKey is a helper to get extendend key from
// a public key, a chain code, an account index and a chain params.
func GetAccountExtendedKey(client bitcoin.CoinServiceClient, net Network, request *FromChainCode) (*bitcoin.GetAccountExtendedKeyResponse, error) {
	chainParams, err := bitcoinChainParams(net)

	if err != nil {
		return nil, err
	}

	return client.GetAccountExtendedKey(
		context.Background(), &bitcoin.GetAccountExtendedKeyRequest{
			PublicKey:    request.PublicKey,
			ChainCode:    request.ChainCode,
			AccountIndex: request.AccountIndex,
			ChainParams:  chainParams,
		})
}
