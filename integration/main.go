package integration

// This file only exists to prevent warnings due to no buildable source files
// when the build tag for enabling the tests is not specified.

// Suppress unused/deadcode warnings when package side-effects are not invoked,
// for example while running linters.
var (
	_ = keychainSvcClient
)
