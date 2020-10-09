package keystore

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ledgerhq/bitcoin-keychain/log"
	"github.com/pkg/errors"

	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
)

// protoEncodingFromScheme is a helper to convert a Scheme from keystore
// package to the corresponding type from bitcoin-lib-grpc.
func protoEncodingFromScheme(scheme Scheme) (bitcoin.AddressEncoding, error) {
	switch scheme {
	case BIP44:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2PKH, nil
	case BIP49:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2SH_P2WPKH, nil
	case BIP84:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2WPKH, nil
	default:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_UNSPECIFIED,
			errors.Wrap(ErrUnrecognizedScheme, fmt.Sprint(scheme))
	}
}

// schemeFromProtoEncoding is a helper to convert an address encoding from
// bitcoin-lib-grpc to the corresponding Scheme in the keystore package.
func schemeFromProtoEncoding(encoding bitcoin.AddressEncoding) (Scheme, error) {
	switch encoding {
	case bitcoin.AddressEncoding_ADDRESS_ENCODING_P2PKH:
		return BIP44, nil
	case bitcoin.AddressEncoding_ADDRESS_ENCODING_P2SH_P2WPKH:
		return BIP49, nil
	case bitcoin.AddressEncoding_ADDRESS_ENCODING_P2WPKH:
		return BIP84, nil
	default:
		return "", errors.Wrap(bitcoin.ErrUnrecognizedAddressEncoding,
			fmt.Sprint(encoding))
	}
}

// bitcoinChainParams is a helper to convert a Network in keystore package to
// the corresponding *bitcoin.ChainParams value in bitcoin-lib-grpc.
func bitcoinChainParams(net Network) (*bitcoin.ChainParams, error) {
	var network bitcoin.BitcoinNetwork

	switch net {
	case Mainnet:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_MAINNET
	case Testnet3:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3
	case Regtest:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_REGTEST
	default:
		return nil, errors.Wrap(ErrUnrecognizedNetwork, fmt.Sprint(net))
	}

	return &bitcoin.ChainParams{
		Network: &bitcoin.ChainParams_BitcoinNetwork{BitcoinNetwork: network},
	}, nil
}

// networkFromChainParams is a helper to convert chain params from bitcoin-lib-grpc
// to the corresponding Network in keystore package.
func networkFromChainParams(params *bitcoin.ChainParams) (Network, error) {
	switch net := params.GetBitcoinNetwork(); net {
	case bitcoin.BitcoinNetwork_BITCOIN_NETWORK_MAINNET:
		return Mainnet, nil
	case bitcoin.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3:
		return Testnet3, nil
	case bitcoin.BitcoinNetwork_BITCOIN_NETWORK_REGTEST:
		return Regtest, nil
	default:
		return "", errors.Wrap(bitcoin.ErrUnrecognizedNetwork, fmt.Sprint(net))
	}
}

// encodeAddress is a helper to serialize a public key to an address, based on
// the Scheme and Network.
func encodeAddress(
	client bitcoin.CoinServiceClient,
	publicKey []byte,
	scheme Scheme,
	net Network,
) (string, error) {
	encoding, err := protoEncodingFromScheme(scheme)
	if err != nil {
		return "", err
	}

	network, err := bitcoinChainParams(net)
	if err != nil {
		return "", err
	}

	addr, err := client.EncodeAddress(
		context.Background(), &bitcoin.EncodeAddressRequest{
			PublicKey:   publicKey,
			Encoding:    encoding,
			ChainParams: network,
		})
	if err != nil {
		return "", nil
	}

	return addr.Address, nil
}

// deriveAddressAtIndex is a helper to derive a child for a registered keychain
// at the given DerivationPath, and encode the corresponding public key to an
// address based on the given Scheme and Network.
func deriveAddress(
	client bitcoin.CoinServiceClient,
	keychain *Meta,
	path DerivationPath,
) (string, error) {
	xPub, err := keychain.ChangeXPub(path.ChangeIndex())
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to get xPub for change index %d", path.ChangeIndex())
	}

	child, err := childKDF(client, xPub, path.AddressIndex())
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to derive extended key %s at child index %d",
			xPub, path.AddressIndex())
	}

	addr, err := encodeAddress(
		client, child.PublicKey, keychain.Main.Scheme, keychain.Main.Network)
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to encode public key %s to %s address on %s",
			hex.EncodeToString(child.PublicKey), keychain.Main.Scheme,
			keychain.Main.Network)
	}

	log.WithFields(log.Fields{
		"id":   keychain.Main.ID.String(),
		"addr": addr,
		"path": path,
	}).Info("[keystore] derive address")

	// Feed address -> derivation path mapping
	keychain.Addresses[addr] = path

	// Feed derivation path -> public key mapping
	keychain.Derivations[path] = hex.EncodeToString(child.PublicKey)

	return addr, nil
}
