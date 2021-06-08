package keystore

import (
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
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

// Schema is a map between keychain ID and the keystore information.
type Schema map[uuid.UUID]*Meta

// NewInMemoryKeystore returns an instance of InMemoryKeystore which implements
// the Keystore interface.
func NewInMemoryKeystore() Keystore {
	return &InMemoryKeystore{
		db:     Schema{},
		client: bitcoin.NewBitcoinClient(),
	}
}

func (s *InMemoryKeystore) Get(id uuid.UUID) (KeychainInfo, error) {
	document, ok := s.db[id]
	if !ok {
		return KeychainInfo{}, ErrKeychainNotFound
	}

	return document.Main, nil
}

func (s *InMemoryKeystore) Delete(id uuid.UUID) error {
	_, ok := s.db[id]
	if !ok {
		return ErrKeychainNotFound
	}

	delete(s.db, id)

	return nil
}

func (s *InMemoryKeystore) Reset(id uuid.UUID) error {
	meta, ok := s.db[id]
	if !ok {
		return ErrKeychainNotFound
	}

	keystoreReset(meta)

	return nil
}

func (s *InMemoryKeystore) Create(
	extendedPublicKey string, fromChainCode *FromChainCode, scheme Scheme, net chaincfg.Network, lookaheadSize uint32,
) (KeychainInfo, error) {

	meta, err := keystoreCreate(
		extendedPublicKey,
		fromChainCode,
		scheme,
		net,
		lookaheadSize,
		s.client,
	)

	if err != nil {
		return KeychainInfo{}, err
	}

	s.db[meta.Main.ID] = &meta

	return meta.Main, nil
}

func (s InMemoryKeystore) GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error) {
	return keystoreGetFreshAddress(&s, id, change)
}

func (s InMemoryKeystore) GetFreshAddresses(
	id uuid.UUID, change Change, size uint32,
) ([]AddressInfo, error) {
	addrs := []AddressInfo{}

	meta, ok := s.db[id]
	if !ok {
		return addrs, ErrKeychainNotFound
	}
	addrs, _, err := keystoreGetFreshAddresses(*meta, s.client, id, change, size)
	if err != nil {
		return addrs, err
	}

	return addrs, nil
}

func (s *InMemoryKeystore) MarkPathAsUsed(id uuid.UUID, path DerivationPath) error {
	// Get keychain by ID
	meta, ok := s.db[id]
	if !ok {
		return ErrKeychainNotFound
	}

	err := keystoreMarkPathAsUsed(meta, id, path) // meta is changed
	if err != nil {
		return err
	}

	return nil
}

func (s *InMemoryKeystore) GetAllObservableAddresses(
	id uuid.UUID, change Change, fromIndex uint32, toIndex uint32,
) ([]AddressInfo, error) {
	meta, ok := s.db[id]
	if !ok {
		return nil, ErrKeychainNotFound
	}
	return keystoreGetAllObservableAddresses(
		meta, s.client, id, change, fromIndex, toIndex,
	)
}

func (s InMemoryKeystore) GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error) {
	meta, ok := s.db[id]
	if !ok {
		return DerivationPath{}, ErrKeychainNotFound
	}

	return keystoreGetDerivationPath(*meta, id, address)
}

func (s *InMemoryKeystore) MarkAddressAsUsed(id uuid.UUID, address string) error {
	return keystoreMarkAddressAsUsed(s, id, address)
}

// GetAddressesPublicKeys reads the derivation-to-publicKey mapping in the keystore,
// and returns extendend public keys corresponding to given derivations.
func (s *InMemoryKeystore) GetAddressesPublicKeys(id uuid.UUID, derivations []DerivationPath) ([]string, error) {
	meta, ok := s.db[id]
	if !ok {
		return nil, ErrKeychainNotFound
	}

	return keystoreGetAddressesPublicKeys(*meta, id, derivations)
}
