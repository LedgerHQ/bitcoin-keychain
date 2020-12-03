package keystore

import (
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/pkg/errors"
)

// InMemoryKeystore implements the Keystore interface where the storage
// is an in-memory map. Useful for unit-tests.
//
// It also includes a client to communicate with a bitcoin-lib-grpc gRPC server
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
// keychain UUID.
//
// Returns an error if the keychain UUID is missing in the keystore.
func (s *InMemoryKeystore) Get(id uuid.UUID) (KeychainInfo, error) {
	document, ok := s.db[id]
	if !ok {
		return KeychainInfo{}, ErrKeychainNotFound
	}

	return document.Main, nil
}

// Delete removes a keychain corresponding to a UUID from the keystore.
func (s *InMemoryKeystore) Delete(id uuid.UUID) error {
	_, ok := s.db[id]
	if !ok {
		return ErrKeychainNotFound
	}

	delete(s.db, id)

	return nil
}

// Create parses a populates the in-memory keystore with the corresponding
// keychain information, based on the provided extended public key, Scheme,
// and Network information.
//
// Only initial state is populated, so no addresses will be inserted into the
// keystore by this method.
func (s *InMemoryKeystore) Create(
	extendedPublicKey string, scheme Scheme, net Network, lookaheadSize uint32,
) (KeychainInfo, error) {
	internalDescriptor, err := MakeDescriptor(extendedPublicKey, Internal, scheme)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(err,
			"failed to make internal descriptor, xkey = %v", extendedPublicKey)
	}

	externalDescriptor, err := MakeDescriptor(extendedPublicKey, External, scheme)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(err,
			"failed to make internal descriptor, xkey = %v", extendedPublicKey)
	}

	externalChild, err := childKDF(s.client, extendedPublicKey, 0)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", extendedPublicKey, 0)
	}

	internalChild, err := childKDF(s.client, extendedPublicKey, 1)
	if err != nil {
		return KeychainInfo{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", extendedPublicKey, 1)
	}

	id := uuid.New()

	keychainInfo := KeychainInfo{
		ID:                          id,
		InternalDescriptor:          internalDescriptor,
		ExternalDescriptor:          externalDescriptor,
		ExtendedPublicKey:           extendedPublicKey,
		SLIP32ExtendedPublicKey:     extendedPublicKey, // TODO: Convert ExtendedPublicKey to SLIP-0132 form
		ExternalXPub:                externalChild.ExtendedKey,
		MaxConsecutiveExternalIndex: 0,
		InternalXPub:                internalChild.ExtendedKey,
		MaxConsecutiveInternalIndex: 0,
		LookaheadSize:               lookaheadSize,
		Scheme:                      scheme,
		Network:                     net,
	}

	s.db[id] = &Meta{
		Main:        keychainInfo,
		Derivations: map[DerivationPath]string{},
		Addresses:   map[string]DerivationPath{},
	}

	return keychainInfo, nil
}

// GetFreshAddress retrieves an unused address from the in-memory keystore at a
// given Change index, for the keychain corresponding to the provided keychain
// ID.
//
// See GetFreshAddresses for getting fresh addresses in bulk, and for further
// details.
func (s InMemoryKeystore) GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error) {
	addrs, err := s.GetFreshAddresses(id, change, 1)
	if err != nil {
		return nil, err
	}

	return &addrs[0], nil
}

// GetFreshAddresses retrieves bulk fresh addresses from the in-memory keystore.
//
// In addition to ensuring that issued addresses are always fresh (unused), the
// method also detects gaps in used addresses and includes it in fresh address
// list.
func (s InMemoryKeystore) GetFreshAddresses(
	id uuid.UUID, change Change, size uint32,
) ([]AddressInfo, error) {
	addrs := []AddressInfo{}

	k, ok := s.db[id]
	if !ok {
		return addrs, ErrKeychainNotFound
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
			path := DerivationPath{uint32(change), index}

			addr, err := deriveAddress(s.client, k, path)
			if err != nil {
				return addrs, err
			}

			addrInfo := AddressInfo{
				Address:    addr,
				Derivation: path,
				Change:     change,
			}

			addrs = append(addrs, addrInfo)
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
func (s *InMemoryKeystore) MarkPathAsUsed(id uuid.UUID, path DerivationPath) error {
	// Get keychain by ID
	k, ok := s.db[id]
	if !ok {
		return ErrKeychainNotFound
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

func (s *InMemoryKeystore) GetAllObservableAddresses(
	id uuid.UUID, change Change, fromIndex uint32, toIndex uint32,
) ([]AddressInfo, error) {
	k, ok := s.db[id]
	if !ok {
		return nil, ErrKeychainNotFound
	}

	maxObservableIndex, err := k.MaxObservableIndex(change)
	if err != nil {
		return nil, err
	}

	length := minUint32(toIndex, maxObservableIndex) - fromIndex

	var result []uint32

	for i := uint32(0); i <= length; i++ {
		result = append(result, fromIndex+i)
	}

	addrs := make([]AddressInfo, 0, len(result))

	for _, i := range result {
		path := DerivationPath{uint32(change), i}

		addr, err := deriveAddress(s.client, k, path)
		if err != nil {
			return nil, err
		}

		addrInfo := AddressInfo{
			Address:    addr,
			Derivation: path,
			Change:     change,
		}

		addrs = append(addrs, addrInfo)
	}

	return addrs, nil
}

// GetDerivationPath reads the address-to-derivations mapping in the keystore,
// and returns the DerivationPath corresponding to the specified address.
func (s InMemoryKeystore) GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error) {
	k, ok := s.db[id]
	if !ok {
		return DerivationPath{}, ErrKeychainNotFound
	}

	path, ok := k.Addresses[address]
	if !ok {
		return DerivationPath{}, ErrAddressNotFound
	}

	return path, nil
}

// MarkAddressAsUsed is a helper to directly mark an address as used. It
// internally fetches the derivation path of the address from the keystore,
// and then marks this DerivationPath value as used.
func (s *InMemoryKeystore) MarkAddressAsUsed(id uuid.UUID, address string) error {
	path, err := s.GetDerivationPath(id, address)
	if err != nil {
		return err
	}

	return s.MarkPathAsUsed(id, path)
}

// GetAddressesPublicKeys reads the derivation-to-publicKey mapping in the keystore,
// and returns extendend public keys corresponding to given derivations.
func (s *InMemoryKeystore) GetAddressesPublicKeys(id uuid.UUID, derivations []DerivationPath) ([]string, error) {
	k, ok := s.db[id]
	if !ok {
		return nil, ErrKeychainNotFound
	}

	publicKeys := make([]string, len(derivations))

	for idx, derivation := range derivations {
		publicKey, ok := k.Derivations[derivation]

		if !ok {
			return nil, ErrDerivationNotFound
		}

		publicKeys[idx] = publicKey
	}

	return publicKeys, nil
}
