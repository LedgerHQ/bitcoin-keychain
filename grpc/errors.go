package grpc

import "github.com/pkg/errors"

var (
	// ErrUnrecognizedNetwork indicates that an unrecognized Bitcoin network,
	// defined in the keychain service, was encountered.
	ErrUnrecognizedNetwork = errors.New("unrecognized keychain network")

	// ErrUnrecognizedChange indicates that an unrecognized Change path was
	// encountered.
	ErrUnrecognizedChange = errors.New("unrecognized change")

	// ErrUnrecognizedScheme indicates that an unrecognized derivation Scheme
	// was encountered.
	ErrUnrecognizedScheme = errors.New("unrecognized scheme")

	// ErrInvalidDerivationPath indicates that a derivation path is invalid or
	// malformed.
	ErrInvalidDerivationPath = errors.New("invalid derivation path")

	// ErrInvalidKeychainID indicates that the UUID representing the keychain
	// could not be serialized / deserialized.
	ErrInvalidKeychainID = errors.New("invalid keychain id")
)
