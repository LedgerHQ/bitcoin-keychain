package grpc

import "github.com/pkg/errors"

var (
	// ErrUnrecognizedNetwork indicates that an unrecognized Bitcoin network,
	// defined in the keychain service, was encountered.
	ErrUnrecognizedNetwork = errors.New("unrecognized keychain network")
)
