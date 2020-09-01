package keystore

import (
	"context"

	"github.com/ledgerhq/bitcoin-keychain-svc/bitcoin"
)

// childKDF is a helper to derive an child extended public key on a child
// index, from a parent extended public key.
//
// This helper can only be used to derive one BIP32 level at a time.
func childKDF(client bitcoin.CoinServiceClient, xPub string, childIndex uint32) (*bitcoin.DeriveExtendedKeyResponse, error) {
	return client.DeriveExtendedKey(
		context.Background(), &bitcoin.DeriveExtendedKeyRequest{
			ExtendedKey: xPub,
			Derivation:  []uint32{childIndex},
		})
}
