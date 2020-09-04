package keystore

import "github.com/pkg/errors"

var (
	// ErrInvalidDescriptor indicates that a descriptor string does not
	// have the expected structure/format.
	ErrInvalidDescriptor = errors.New("invalid descriptor")

	// ErrUnrecognizedScheme indicates that the parsed scheme of a descriptor
	// is invalid or missing.
	ErrUnrecognizedScheme = errors.New("unrecognized scheme")

	// ErrUnrecognizedChange indicates that the Change index encountered was
	// non-standard, and cannot be handled properly.
	ErrUnrecognizedChange = errors.New("unrecognized change")

	// ErrDescriptorNotFound indicates an attempt to get a descriptor from a
	// keystore that has not been registered.
	ErrDescriptorNotFound = errors.New("descriptor not found")
)
