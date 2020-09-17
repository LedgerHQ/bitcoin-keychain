package grpc

import "github.com/pkg/errors"

var (
	// ErrUnrecognizedNetwork indicates that an unrecognized Bitcoin network,
	// defined in the keychain service, was encountered.
	ErrUnrecognizedNetwork = errors.New("unrecognized keychain network")

	// ErrUnrecognizedChange indicates that an unrecognized Change path was
	// encountered.
	ErrUnrecognizedChange = errors.New("unrecognized change")

	// ErrInvalidDerivationPath indicates that a derivation path is invalid or
	// malformed.
	ErrInvalidDerivationPath = errors.New("invalid derivation path")
)
