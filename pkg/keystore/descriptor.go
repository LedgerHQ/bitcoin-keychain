package keystore

import (
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// DescriptorTokens models the result of parsing an output descriptor.
type DescriptorTokens struct {
	XPub   string
	Scheme Scheme
}

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

	if strings.HasPrefix(descriptor, "sh(wpkh(") {
		scheme = BIP49
	} else if strings.HasPrefix(descriptor, "wpkh(") {
		scheme = BIP84
	} else if strings.HasPrefix(descriptor, "pkh(") {
		scheme = BIP44
	} else {
		return DescriptorTokens{}, errors.Wrapf(ErrUnrecognizedScheme,
			"failed to parse descriptor %v", descriptor)
	}

	// Match the base key using regex, and ignore everything else.
	// FIXME: make the regex match more robust.
	r := regexp.MustCompile(`.*\((?:\[.*])?([\w+]*).*\).*`)
	groups := r.FindStringSubmatch(descriptor)
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
