package keystore

// Keystore is an interface that all keychain storage backends must implement.
// Currently, there are two Keystore implementations available:
//   InMemoryKeystore: useful for unit-tests
//   RedisKeystore:    TBA
type Keystore interface {
	Get(descriptor string) (KeychainInfo, error)
	Create(descriptor string) (KeychainInfo, error)
}

// Scheme defines the scheme on which a keychain entry is based.
type Scheme string

const (
	// BIP44 indicates that the keychain scheme is legacy.
	BIP44 Scheme = "BIP44"

	// BIP49 indicates that the keychain scheme is segwit.
	BIP49 Scheme = "BIP49"

	// BIP84 indicates that the keychain scheme is native segwit.
	BIP84 Scheme = "BIP84"
)

const lookaheadSize = 20

// KeychainInfo models the global information related to an account registered
// in the keystore.
//
// Rather than using the associated gRPC message struct, it is defined here
// independently to avoid having gRPC dependency in this package.
type KeychainInfo struct {
	Descriptor              string `json:"descriptor"`
	XPub                    string `json:"xpub"`                       // Extended public key serialized with standard HD version bytes
	SLIP32ExtendedPublicKey string `json:"slip32_extended_public_key"` // Extended public key serialized with SLIP-0132 HD version bytes
	ExternalXPub            string `json:"external_xpub"`              // External chain extended public key at HD tree depth 4
	MaxExternalIndex        uint32 `json:"max_external_index"`         // Number of external chain addresses in keychain
	InternalXPub            string `json:"internal_xpub"`              // Internal chain extended public key at HD tree depth 4
	MaxInternalIndex        uint32 `json:"max_internal_index"`         // Number of external chain addresses in keychain
	LookaheadSize           uint32 `json:"lookahead_size"`             // Numerical size of the lookahead zone
	Scheme                  Scheme `json:"scheme"`                     // String identifier for keychain scheme
}

type derivationToPublicKeyMap map[string]struct {
	PublicKey string `json:"public_key"` // Public key at HD tree depth 5
	Used      bool   `json:"used"`       // Whether any txn history at derivation
}

// Schema is a map between account descriptors and account information.
type Schema map[string]Meta

// Meta is a struct containing account details corresponding to a descriptor,
// such as derivations, addresses, etc.
type Meta struct {
	Main        KeychainInfo             `json:"main"`
	Derivations derivationToPublicKeyMap `json:"derivations"`
	Addresses   map[string][2]uint32     `json:"addresses"` // derivation path at HD tree depth 5
}
