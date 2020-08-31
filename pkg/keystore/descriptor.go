package keystore

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// DescriptorTokens models the result of parsing an output descriptor.
type DescriptorTokens struct {
	XPub   string
	Scheme Scheme
}

// Match the base key using regex, and ignore everything else.
// FIXME: make the regex match more robust.
var descriptorRegex = regexp.MustCompile(`.*\((?:\[.*])?([\w+]*).*\).*`)

// ParseDescriptor is a very basic tokenizer of output descriptor strings.
//
// References:
//   https://github.com/bitcoin/bitcoin/blob/master/doc/descriptors.md
//   https://github.com/bitcoin-core/HWI/blob/master/hwilib/descriptor.py
//   https://github.com/bitcoin/bitcoin/blob/master/src/script/descriptor.cpp
//
// TODO: Upstream this to btcsuite/btcutil, and access it through bitcoin-svc.
func ParseDescriptor(descriptor string) (DescriptorTokens, error) {
	var scheme Scheme

	switch {
	case strings.HasPrefix(descriptor, "sh(wpkh("):
		scheme = BIP49
	case strings.HasPrefix(descriptor, "wpkh("):
		scheme = BIP84
	case strings.HasPrefix(descriptor, "pkh("):
		scheme = BIP44
	default:
		return DescriptorTokens{}, errors.Wrapf(ErrUnrecognizedScheme,
			"failed to parse descriptor %v", descriptor)
	}

	groups := descriptorRegex.FindStringSubmatch(descriptor)
	if len(groups) != 2 {
		return DescriptorTokens{}, errors.Wrapf(
			ErrInvalidDescriptor, "failed to parse descriptor %v", descriptor)
	}

	xpub := groups[1]
	if xpub == "" {
		return DescriptorTokens{}, errors.Wrapf(ErrInvalidDescriptor,
			"empty xpub in descriptor %v", descriptor)
	}

	return DescriptorTokens{
		XPub:   xpub,
		Scheme: scheme,
	}, nil
}
