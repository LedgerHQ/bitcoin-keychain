package keystore

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ledgerhq/bitcoin-keychain-svc/bitcoin"
)

// encodingFromScheme is a helper to convert a Scheme from keystore package
// to the corresponding type from bitcoin-svc.
func encodingFromScheme(scheme Scheme) bitcoin.AddressEncoding {
	switch scheme {
	case BIP44:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2PKH
	case BIP49:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2SH_P2WPKH
	case BIP84:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_P2WPKH
	default:
		return bitcoin.AddressEncoding_ADDRESS_ENCODING_UNSPECIFIED
	}
}

// schemeFromEncoding is a helper to convert an address encoding from
// bitcoin-svc to the corresponding Scheme in the keystore package.
func schemeFromEncoding(encoding bitcoin.AddressEncoding) (Scheme, error) {
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
// the corresponding *bitcoin.ChainParams value in bitcoin-svc.
func bitcoinChainParams(net Network) *bitcoin.ChainParams {
	var network bitcoin.BitcoinNetwork

	switch net {
	case Mainnet:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_MAINNET
	case Testnet3:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_TESTNET3
	case Regtest:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_REGTEST

	default:
		network = bitcoin.BitcoinNetwork_BITCOIN_NETWORK_UNSPECIFIED
	}

	return &bitcoin.ChainParams{
		Network: &bitcoin.ChainParams_BitcoinNetwork{BitcoinNetwork: network},
	}
}

// networkFromChainParams is a helper to convert chain params from bitcoin-svc
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
	encoding := encodingFromScheme(scheme)
	network := bitcoinChainParams(net)

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

// deriveAddressAtIndex is a helper to derive a child from an xPub at the given
// index, and encode the corresponding public key to an address based on the
// given Scheme and Network.
func deriveAddressAtIndex(
	client bitcoin.CoinServiceClient,
	xPub string,
	index uint32,
	scheme Scheme,
	net Network,
) (string, error) {
	child, err := childKDF(client, xPub, index)
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to derive extended key %s at child index %d", xPub, index)
	}

	addr, err := encodeAddress(client, child.PublicKey, scheme, net)
	if err != nil {
		return "", errors.Wrapf(err,
			"failed to encode public key %s to %s address on %s",
			hex.EncodeToString(child.PublicKey), scheme, net)
	}

	return addr, nil
}
