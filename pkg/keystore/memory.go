package keystore

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/ledgerhq/bitcoin-keychain-svc/bitcoin"
)

// InMemoryKeystore implements the Keystore interface where the storage
// is an in-memory map. Useful for unit-tests.
//
// It also includes a client to communicate with a bitcoin-svc gRPC server
// for protocol-level operations.
type InMemoryKeystore struct {
	db     Schema
	client bitcoin.CoinServiceClient
}

// NewInMemoryKeystore returns an instance of InMemoryKeystore which implements
// the Keystore interface.
func NewInMemoryKeystore() Keystore {
	return &InMemoryKeystore{
		db:     Schema{},
		client: bitcoin.NewBitcoinClient(),
	}
}

// Get returns a previously stored keychain information based on the provided
// descriptor string.
//
// Returns an error if the descriptor is missing in the keystore.
func (s *InMemoryKeystore) Get(descriptor string) (KeychainInfo, error) {
	document, ok := s.db[descriptor]
	if !ok {
		return KeychainInfo{}, errors.New("does not exist")
	}

	return document.Main, nil
}

// Create parses a descriptor string and populars the in-memory keystore with
// the corresponding keychain information.
//
// Only initial state is populated, so no addresses will be inserted into the
// keystore by this method.
func (s *InMemoryKeystore) Create(descriptor string) (KeychainInfo, error) {
	tokens, err := ParseDescriptor(descriptor)
	if err != nil {
		return KeychainInfo{}, err
	}

	ckdFunc := func(childIndex uint32) (string, string, error) {
		child, err := s.client.DeriveExtendedKey(
			context.Background(), &bitcoin.DeriveExtendedKeyRequest{
				ExtendedKey: tokens.XPub,
				Derivation:  []uint32{childIndex},
			})
		if err != nil {
			return "", "", nil
		}

		return hex.EncodeToString(child.PublicKey), hex.EncodeToString(child.ChainCode), nil
	}

	externalPublicKey, externalChainCode, err := ckdFunc(0)
	if err != nil {
		return KeychainInfo{}, err
	}

	internalPublicKey, internalChainCode, err := ckdFunc(1)
	if err != nil {
		return KeychainInfo{}, err
	}

	keychainInfo := KeychainInfo{
		Descriptor:              descriptor,
		XPub:                    tokens.XPub,
		SLIP32ExtendedPublicKey: tokens.XPub, // TODO: Convert XPub to SLIP-0132 form
		ExternalPublicKey:       externalPublicKey,
		ExternalChainCode:       externalChainCode,
		MaxExternalIndex:        0,
		InternalPublicKey:       internalPublicKey,
		InternalChainCode:       internalChainCode,
		MaxInternalIndex:        0,
		LookaheadSize:           20,
		Scheme:                  tokens.Scheme,
	}

	s.db[descriptor] = Meta{
		Main:        keychainInfo,
		Derivations: nil,
		Addresses:   nil,
	}

	return keychainInfo, nil
}
