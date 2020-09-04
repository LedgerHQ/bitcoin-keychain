package keystore

import (
	"github.com/ledgerhq/bitcoin-keychain-svc/bitcoin"
	"github.com/pkg/errors"
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
		return KeychainInfo{}, ErrDescriptorNotFound
	}

	return document.Main, nil
}

// Create parses a descriptor string and populars the in-memory keystore with
// the corresponding keychain information.
//
// Only initial state is populated, so no addresses will be inserted into the
// keystore by this method.
func (s *InMemoryKeystore) Create(descriptor string, net Network) (KeychainInfo, error) {
	tokens, err := ParseDescriptor(descriptor)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(
			err, "failed to parse descriptor %v", descriptor)
	}

	externalChild, err := childKDF(s.client, tokens.XPub, 0)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", tokens.XPub, 0)
	}

	internalChild, err := childKDF(s.client, tokens.XPub, 1)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", tokens.XPub, 1)
	}

	keychainInfo := KeychainInfo{
		Descriptor:                descriptor,
		XPub:                      tokens.XPub,
		SLIP32ExtendedPublicKey:   tokens.XPub, // TODO: Convert XPub to SLIP-0132 form
		ExternalXPub:              externalChild.ExtendedKey,
		ExternalFreshAddressIndex: 0,
		InternalXPub:              internalChild.ExtendedKey,
		InternalFreshAddressIndex: 0,
		LookaheadSize:             lookaheadSize,
		Scheme:                    tokens.Scheme,
		Network:                   net,
	}

	s.db[descriptor] = Meta{
		Main:        keychainInfo,
		Derivations: nil,
		Addresses:   nil,
	}

	return keychainInfo, nil
}

// GetFreshAddress retrieves an unused address from the in-memory keystore at a
// given Change index, for the keychain corresponding to the provided
// descriptor.
//
// See GetFreshAddresses for getting fresh addresses in bulk.
//
// This is a read-only operation.
func (s InMemoryKeystore) GetFreshAddress(descriptor string, change Change) (string, error) {
	k, ok := s.db[descriptor]
	if !ok {
		return "", ErrDescriptorNotFound
	}

	changeXPub, err := k.ChangeXPub(change)
	if err != nil {
		return "", err
	}

	freshAddressIndex, err := k.FreshAddressIndex(change)
	if err != nil {
		return "", err
	}

	addr, err := deriveAddressAtIndex(s.client, changeXPub, freshAddressIndex,
		k.Main.Scheme, k.Main.Network)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// GetFreshAddresses retrieves bulk fresh addresses from the in-memory keystore.
//
// See GetFreshAddress for further details.
//
// This is a read-only operation.
func (s InMemoryKeystore) GetFreshAddresses(descriptor string, change Change, size uint32) ([]string, error) {
	addrs := []string{}

	k, ok := s.db[descriptor]
	if !ok {
		return addrs, ErrDescriptorNotFound
	}

	changeXPub, err := k.ChangeXPub(change)
	if err != nil {
		return addrs, err
	}

	freshAddressIndex, err := k.FreshAddressIndex(change)
	if err != nil {
		return addrs, err
	}

	for i := uint32(0); i < size; i++ {
		addr, err := deriveAddressAtIndex(s.client, changeXPub, freshAddressIndex+i,
			k.Main.Scheme, k.Main.Network)
		if err != nil {
			return addrs, err
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}
