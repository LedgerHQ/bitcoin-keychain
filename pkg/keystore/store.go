package keystore

import (
	"fmt"

	"github.com/pkg/errors"
)

// Keystore is an interface that all keychain storage backends must implement.
// Currently, there are two Keystore implementations available:
//   InMemoryKeystore: useful for unit-tests
//   RedisKeystore:    TBA
type Keystore interface {
	Get(descriptor string) (KeychainInfo, error)
	Create(descriptor string, net Network) (KeychainInfo, error)
	GetFreshAddress(descriptor string, change Change) (string, error)
	GetFreshAddresses(descriptor string, change Change, size uint32) ([]string, error)
}

// Scheme defines the scheme on which a keychain entry is based.
type Scheme string

const (
	// BIP44 indicates that the keychain scheme is legacy.
	BIP44 Scheme = "BIP44"

	// BIP49 indicates that the keychain scheme is segwit.
	BIP49 Scheme = "BIP49"

	// BIP84 indicates that the keychain scheme is native segwit.
	BIP84 Scheme = "BIP84"
)

// Network defines the network (and therefore the chain parameters)
// that a keychain is associated to.
type Network string

const (
	// Mainnet indicates the main Bitcoin network
	Mainnet Network = "mainnet"

	// Testnet3 indicates the current Bitcoin test network
	Testnet3 Network = "testnet3"

	// Regtest indicates the Bitcoin regression test network
	Regtest Network = "regtest"
)

const lookaheadSize = 20

// KeychainInfo models the global information related to an account registered
// in the keystore.
//
// Rather than using the associated gRPC message struct, it is defined here
// independently to avoid having gRPC dependency in this package.
type KeychainInfo struct {
	Descriptor                string  `json:"descriptor"`
	XPub                      string  `json:"xpub"`                         // Extended public key serialized with standard HD version bytes
	SLIP32ExtendedPublicKey   string  `json:"slip32_extended_public_key"`   // Extended public key serialized with SLIP-0132 HD version bytes
	ExternalXPub              string  `json:"external_xpub"`                // External chain extended public key at HD tree depth 4
	ExternalFreshAddressIndex uint32  `json:"external_fresh_address_index"` // Index of the next fresh address on the external chain
	InternalXPub              string  `json:"internal_xpub"`                // Internal chain extended public key at HD tree depth 4
	InternalFreshAddressIndex uint32  `json:"internal_fresh_address_index"` // Index of the next fresh address on the internal chain
	LookaheadSize             uint32  `json:"lookahead_size"`               // Numerical size of the lookahead zone
	Scheme                    Scheme  `json:"scheme"`                       // String identifier for keychain scheme
	Network                   Network `json:"network"`                      // String denoting the network to use for encoding addresses
}

type derivationToPublicKeyMap map[DerivationPath]struct {
	PublicKey string `json:"public_key"` // Public key at HD tree depth 5
	Used      bool   `json:"used"`       // Whether any txn history at derivation
}

// Schema is a map between account descriptors and account information.
type Schema map[string]Meta

// Meta is a struct containing account details corresponding to a descriptor,
// such as derivations, addresses, etc.
type Meta struct {
	Main        KeychainInfo              `json:"main"`
	Derivations derivationToPublicKeyMap  `json:"derivations"`
	Addresses   map[string]DerivationPath `json:"addresses"` // derivation path at HD tree depth 5
}

func (m Meta) ChangeXPub(change Change) (string, error) {
	switch change {
	case External:
		return m.Main.ExternalXPub, nil
	case Internal:
		return m.Main.InternalXPub, nil
	default:
		return "", errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

func (m Meta) FreshAddressIndex(change Change) (uint32, error) {
	switch change {
	case External:
		return m.Main.ExternalFreshAddressIndex, nil
	case Internal:
		return m.Main.InternalFreshAddressIndex, nil
	default:
		return 0, errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}
