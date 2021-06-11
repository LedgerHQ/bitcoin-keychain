package keystore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
)

// RedisKeystore implements the Keystore interface where the storage
// is an in-memory map. Useful for unit-tests.
//
// It also includes a client to communicate with a bitcoin-lib-grpc server
// for protocol-level operations.
type RedisKeystore struct {
	db     *redis.Client
	client bitcoin.CoinServiceClient
}

// RedisKeystore returns an instance of RedisKeystore which implements
// the Keystore interface.
func NewRedisKeystore(redisOpts *redis.Options) (*RedisKeystore, error) {
	rdb := redis.NewClient(redisOpts)

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("Pinging redis failed: %w", err)
	}

	return &RedisKeystore{
		db:     rdb,
		client: bitcoin.NewBitcoinClient(),
	}, nil
}

func (s *RedisKeystore) Get(id uuid.UUID) (KeychainInfo, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return KeychainInfo{}, ErrKeychainNotFound
	}

	return meta.Main, nil
}

func (s *RedisKeystore) Delete(id uuid.UUID) error {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return ErrKeychainNotFound
	}

	s.db.Del(context.Background(), id.String())

	return nil
}

func (s *RedisKeystore) Reset(id uuid.UUID) error {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return ErrKeychainNotFound
	}

	meta.ResetKeychainMeta()

	if err := set(s.db, id.String(), meta); err != nil {
		return err
	}

	return nil
}

func (s *RedisKeystore) Create(
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

	if err := set(s.db, meta.Main.ID.String(), meta); err != nil {
		return KeychainInfo{}, err
	}

	return meta.Main, nil
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

func (s *RedisKeystore) GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error) {
	var meta Meta
	err := get(s.db, id.String(), &meta)
	if err != nil {
		return DerivationPath{}, ErrKeychainNotFound
	}

	return meta.keystoreGetDerivationPath(address)
}

func (s *RedisKeystore) MarkAddressAsUsed(id uuid.UUID, address string) error {
	return keystoreMarkAddressAsUsed(s, id, address)
}

func set(c *redis.Client, key string, value interface{}) error {
	p, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Set(context.Background(), key, string(p), 0).Err()
}

func get(c *redis.Client, key string, dest interface{}) error {
	p, err := c.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(p), dest)
}

func (s *RedisKeystore) GetAddressesPublicKeys(id uuid.UUID, derivations []DerivationPath) ([]string, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return nil, ErrKeychainNotFound
	}

	return meta.keystoreGetAddressesPublicKeys(derivations)
}
