package keystore

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
)

type baseRedisKeystore struct {
	db     *redis.Client
	client bitcoin.CoinServiceClient
}

func (s *baseRedisKeystore) Get(id uuid.UUID) (KeychainInfo, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return KeychainInfo{}, ErrKeychainNotFound
	}

	return meta.Main, nil
}

func (s *baseRedisKeystore) Delete(id uuid.UUID) error {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return ErrKeychainNotFound
	}

	s.db.Del(context.Background(), id.String())

	return nil
}

func (s *baseRedisKeystore) Reset(id uuid.UUID) error {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return ErrKeychainNotFound
	}

	meta.ResetKeychainMeta()

	return set(s.db, id.String(), meta)
}

func (s *baseRedisKeystore) Create(
	extendedPublicKey string, fromChainCode *FromChainCode, scheme Scheme,
	net chaincfg.Network, lookaheadSize uint32, index uint32, metadata string,
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

	if err := set(s.db, meta.Main.ID.String(), meta); err != nil {
		return KeychainInfo{}, err
	}

	return meta.Main, nil
}

func (s *baseRedisKeystore) GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error) {
	var meta Meta
	err := get(s.db, id.String(), &meta)
	if err != nil {
		return DerivationPath{}, ErrKeychainNotFound
	}

	return meta.keystoreGetDerivationPath(address)
}

func (s *baseRedisKeystore) GetAddressesPublicKeys(id uuid.UUID, derivations []DerivationPath) ([]string, error) {
	var meta Meta

	err := get(s.db, id.String(), &meta)
	if err != nil {
		return nil, ErrKeychainNotFound
	}

	return meta.keystoreGetAddressesPublicKeys(derivations)
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
