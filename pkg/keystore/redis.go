package keystore

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
)

// RedisKeystore implements the Keystore interface where the storage
// is an in-memory map. Useful for unit-tests.
//
// It also includes a client to communicate with a bitcoin-lib-grpc server
// for protocol-level operations.
type RedisKeystore struct {
	baseRedisKeystore
}

// RedisKeystore returns an instance of RedisKeystore which implements
// the Keystore interface.
func NewRedisKeystore(redisOpts *redis.Options) (*RedisKeystore, error) {
	rdb := redis.NewClient(redisOpts)

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("Pinging redis failed: %w", err)
	}

	baseKeystore := baseRedisKeystore{
		db:     rdb,
		client: bitcoin.NewBitcoinClient(),
	}

	return &RedisKeystore{baseKeystore}, nil
}

func (s *RedisKeystore) GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error) {
	addrs, err := s.GetFreshAddresses(id, change, 1)
	if err != nil {
		return nil, err
	}
	return &addrs[0], err
}

func (s *RedisKeystore) GetFreshAddresses(
	id uuid.UUID, change Change, size uint32,
) ([]AddressInfo, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return []AddressInfo{}, ErrKeychainNotFound
	}

	addrs, err := meta.keystoreGetFreshAddresses(s.client, change, size)
	if err != nil {
		return addrs, err
	}

	if err := set(s.db, id.String(), meta); err != nil {
		return nil, err
	}

	return addrs, nil
}

func (s *RedisKeystore) MarkPathAsUsed(id uuid.UUID, path DerivationPath) error {
	// Get keychain by ID
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return ErrKeychainNotFound
	}

	err = meta.keystoreMarkPathAsUsed(path)
	if err != nil {
		return err
	}

	return set(s.db, id.String(), meta)
}

func (s *RedisKeystore) GetAllObservableAddresses(
	id uuid.UUID, change Change, fromIndex uint32, toIndex uint32,
) ([]AddressInfo, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return nil, ErrKeychainNotFound
	}

	addrs, err := meta.keystoreGetAllObservableAddresses(
		s.client, change, fromIndex, toIndex,
	)
	if err != nil {
		return addrs, err
	}

	err = set(s.db, id.String(), meta)
	if err != nil {
		return addrs, err
	}

	return addrs, nil
}

func (s *RedisKeystore) MarkAddressAsUsed(id uuid.UUID, address string) error {
	return keystoreMarkAddressAsUsed(s, id, address)
}
