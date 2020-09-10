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
		Descriptor:                  descriptor,
		XPub:                        tokens.XPub,
		SLIP32ExtendedPublicKey:     tokens.XPub, // TODO: Convert XPub to SLIP-0132 form
		ExternalXPub:                externalChild.ExtendedKey,
		MaxConsecutiveExternalIndex: 0,
		InternalXPub:                internalChild.ExtendedKey,
		MaxConsecutiveInternalIndex: 0,
		LookaheadSize:               lookaheadSize,
		Scheme:                      tokens.Scheme,
		Network:                     net,
	}

	s.db[descriptor] = &Meta{
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
// See GetFreshAddresses for getting fresh addresses in bulk, and for further
// details.
func (s InMemoryKeystore) GetFreshAddress(descriptor string, change Change) (string, error) {
	addrs, err := s.GetFreshAddresses(descriptor, change, 1)
	if err != nil {
		return "", err
	}

	return addrs[0], nil
}

// GetFreshAddresses retrieves bulk fresh addresses from the in-memory keystore.
//
// In addition to ensuring that issued addresses are always fresh (unused), the
// method also detects gaps in used addresses and includes it in fresh address
// list.
func (s InMemoryKeystore) GetFreshAddresses(
	descriptor string, change Change, size uint32,
) ([]string, error) {
	addrs := []string{}

	k, ok := s.db[descriptor]
	if !ok {
		return addrs, ErrDescriptorNotFound
	}

	changeXPub, err := k.ChangeXPub(change)
	if err != nil {
		return addrs, err
	}

	maxConsecutiveIndex, err := k.MaxConsecutiveIndex(change)
	if err != nil {
		return addrs, err
	}

	nonConsecutiveIndexes, err := k.NonConsecutiveIndexes(change)
	if err != nil {
		return nil, err
	}

	for i := uint32(0); uint32(len(addrs)) < size; i++ {
		index := maxConsecutiveIndex + i

		// Skip any index that exists in non-consecutive indexes, to prevent
		// address reuse.
		if !contains(nonConsecutiveIndexes, index) {
			addr, err := deriveAddressAtIndex(s.client, changeXPub, index,
				k.Main.Scheme, k.Main.Network)
			if err != nil {
				return addrs, err
			}

			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

// MarkPathAsUsed sets a given derivation path as used. It records bookkeeping
// information about gaps in the derivation.
//
// A derivation path is considered "used" if it has transaction history.
//
// If marking of a derivation path as used introduces any gaps, they are
// detected and saved in the keystore. For this we rely on two main fields:
//   MaxConsecutiveIndex   -> the largest consecutive index without any gaps
//   NonConsecutiveIndexes -> list of used indexes that introduced gaps
func (s *InMemoryKeystore) MarkPathAsUsed(descriptor string, path DerivationPath) error {
	// Get keychain by descriptor
	k, ok := s.db[descriptor]
	if !ok {
		return ErrDescriptorNotFound
	}

	change := path.ChangeIndex()

	maxConsecutiveIndex, err := k.MaxConsecutiveIndex(change)
	if err != nil {
		return err
	}

	nonConsecutiveIndexes, err := k.NonConsecutiveIndexes(change)
	if err != nil {
		return err
	}

	switch {
	// CASE 1: Address index being marked as used already falls within the
	// range of consecutive indexes. This is typically when an address index
	// is marked as used twice.
	case path.AddressIndex() < maxConsecutiveIndex:
		// Nothing to do in this case.

	// CASE 2: Mark as used at the boundary of the consecutive used indexes, by
	// incrementing the max consecutive index.
	case path.AddressIndex() == maxConsecutiveIndex:
		maxConsecutiveIndex++

		// Handle case when the max consecutive index overreaches into the
		// non-consecutive indexes. This typically happens when a gap is filled
		// by marking the gap address index as used.
		//
		// Repeat this step until the max consecutive index is outside the
		// overlap of non-consecutive indexes, where it is safe to issue a
		// fresh address.
		for contains(nonConsecutiveIndexes, maxConsecutiveIndex) {
			maxConsecutiveIndex++
		}

		// Save the max consecutive index. It is important to perform this step
		// before saving the non-consecutive indexes.
		if err := k.SetMaxConsecutiveIndex(change, maxConsecutiveIndex); err != nil {
			return err
		}

		// Reconcile non-consecutive indexes depending on the updated state of
		// the max consecutive index.
		//
		// TODO: Implement a dedicated method for this reconciliation, since the
		//       non-consecutive indexes are never changed until this step.
		if err := k.SetNonConsecutiveIndexes(change, nonConsecutiveIndexes); err != nil {
			return err
		}

	// CASE 3: Attempt to introduce a gap after the max consecutive index.
	//
	// Consider the following list of address indexes as the state of the
	// keychain. * indicates that an index is used.
	//
	// Before:
	//   state                  : 0*  1  2  3  4  5
	//   max consecutive index  : 1
	//   non-consecutive indexes: []
	//
	// After: mark index 2 as used
	//   state                  : 0*  1  2*  3  4  5
	//   max consecutive index  : 1   <- no change
	//   non-consecutive indexes: [2] <- add the index that created a gap
	case path.AddressIndex() > maxConsecutiveIndex:
		// Add address index to list of non-consecutive indexes (if does not
		// exist already).
		if !contains(nonConsecutiveIndexes, path.AddressIndex()) {
			nonConsecutiveIndexes = append(
				nonConsecutiveIndexes, path.AddressIndex())

			err := k.SetNonConsecutiveIndexes(change, nonConsecutiveIndexes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *InMemoryKeystore) GetAllObservableIndexes(
	descriptor string, change Change, from uint32, to uint32,
) ([]uint32, error) {
	k, ok := s.db[descriptor]
	if !ok {
		return nil, ErrDescriptorNotFound
	}

	maxObservableIndex, err := k.MaxObservableIndex(change)
	if err != nil {
		return nil, err
	}

	length := minUint32(to-from, maxObservableIndex-from)

	var result []uint32

	for i := uint32(0); i <= length; i++ {
		result = append(result, from+i)
	}

	return result, nil
}
