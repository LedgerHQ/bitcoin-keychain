package keystore

import (
	"fmt"

	"github.com/pkg/errors"
)

// MakeDescriptor builds output descriptor strings for a given Change path
// and Scheme.
//
// References:
//   https://github.com/bitcoin/bitcoin/blob/master/doc/descriptors.md
//   https://github.com/bitcoin-core/HWI/blob/master/hwilib/descriptor.py
//   https://github.com/bitcoin/bitcoin/blob/master/src/script/descriptor.cpp
func MakeDescriptor(extendedPublicKey string, change Change, scheme Scheme) (string, error) {
	switch scheme {
	case BIP44:
		return fmt.Sprintf("pkh(%s/%d/*)", extendedPublicKey, change), nil
	case BIP49:
		return fmt.Sprintf("sh(wpkh(%s/%d/*))", extendedPublicKey, change), nil
	case BIP84:
		return fmt.Sprintf("wpkh(%s/%d/*)", extendedPublicKey, change), nil
	default:
		return "", errors.Wrapf(ErrUnrecognizedScheme, "%v", scheme)
	}
}
