package keystore

import (
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
)

// Schema is a map between keychain ID and the keystore information.
type schema map[uuid.UUID]*Meta

// InMemoryKeystore implements the Keystore interface where the storage
// is an in-memory map. Useful for unit-tests.
//
// It also includes a client to communicate with a bitcoin-lib-grpc gRPC server
// for protocol-level operations.
type InMemoryKeystore struct {
	db     schema
	client bitcoin.CoinServiceClient
}

// NewInMemoryKeystore returns an instance of InMemoryKeystore which implements
// the Keystore interface.
func NewInMemoryKeystore() Keystore {
	return &InMemoryKeystore{
		db:     schema{},
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

	meta.ResetKeychainMeta()

	return nil
}

func (s *InMemoryKeystore) Create(
	extendedPublicKey string, fromChainCode *FromChainCode, scheme Scheme, net chaincfg.Network, lookaheadSize uint32, index uint32, metadata string,
) (KeychainInfo, error) {
	meta, err := keystoreCreate(
		extendedPublicKey,
		fromChainCode,
		scheme,
		net,
		lookaheadSize,
		index,
		metadata,
		s.client,
	)

	if err != nil {
		return KeychainInfo{}, err
	}

	s.db[meta.Main.ID] = &meta

	return meta.Main, nil
}

func (s InMemoryKeystore) GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error) {
	addrs, err := s.GetFreshAddresses(id, change, 1)
	if err != nil {
		return nil, err
	}
	return &addrs[0], err
}

func (s InMemoryKeystore) GetFreshAddresses(
	id uuid.UUID, change Change, size uint32,
) ([]AddressInfo, error) {
	addrs := []AddressInfo{}

	meta, ok := s.db[id]
	if !ok {
		return addrs, ErrKeychainNotFound
	}
	addrs, err := meta.keystoreGetFreshAddresses(s.client, change, size)
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

	err := meta.keystoreMarkPathAsUsed(path)
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
	return meta.keystoreGetAllObservableAddresses(
		s.client, change, fromIndex, toIndex,
	)
}

func (s InMemoryKeystore) GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error) {
	meta, ok := s.db[id]
	if !ok {
		return DerivationPath{}, ErrKeychainNotFound
	}

	return meta.keystoreGetDerivationPath(address)
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

	return meta.keystoreGetAddressesPublicKeys(derivations)
}
