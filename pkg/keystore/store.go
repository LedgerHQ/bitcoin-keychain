package keystore

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Keystore is an interface that all keychain storage backends must implement.
// Currently, there are two Keystore implementations available:
//   InMemoryKeystore: useful for unit-tests
//   RedisKeystore:    TBA
type Keystore interface {
	Get(id uuid.UUID) (KeychainInfo, error)
	Delete(id uuid.UUID) error
	Reset(id uuid.UUID) error
	Create(extendedPublicKey string, fromChainCode *FromChainCode, scheme Scheme, net Network,
		lookaheadSize uint32) (KeychainInfo, error)
	GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error)
	GetFreshAddresses(id uuid.UUID, change Change, size uint32) ([]AddressInfo, error)
	MarkPathAsUsed(id uuid.UUID, path DerivationPath) error
	MarkAddressAsUsed(id uuid.UUID, address string) error
	GetAllObservableAddresses(id uuid.UUID, change Change,
		fromIndex uint32, toIndex uint32) ([]AddressInfo, error)
	GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error)
	GetAddressesPublicKeys(id uuid.UUID, derivations []DerivationPath) ([]string, error)
}

// DefaultLookaheadSize defines the zone of addresses that the keychain must
// observe.
const DefaultLookaheadSize = 20

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

// Network defines the network (and therefore the chain parameters)
// that a keychain is associated to.
type Network string

const (
	// Mainnet indicates the main Bitcoin network
	Mainnet Network = "mainnet"

	// Testnet3 indicates the current Bitcoin test network
	Testnet3 Network = "testnet3"

	// Regtest indicates the Bitcoin regression test network
	Regtest Network = "regtest"
)

// KeychainInfo models the global information related to an account registered
// in the keystore.
//
// Rather than using the associated gRPC message struct, it is defined here
// independently to avoid having gRPC dependency in this package.
type KeychainInfo struct {
	ID                            uuid.UUID `json:"id"`                               // Keychain ID as a uuid.UUID type
	ExternalDescriptor            string    `json:"external_descriptor"`              // External chain output descriptor
	InternalDescriptor            string    `json:"internal_descriptor"`              // Internal chain output descriptor
	ExtendedPublicKey             string    `json:"extended_public_key"`              // Extended public key serialized with standard HD version bytes
	SLIP32ExtendedPublicKey       string    `json:"slip32_extended_public_key"`       // Extended public key serialized with SLIP-0132 HD version bytes
	ExternalXPub                  string    `json:"external_xpub"`                    // External chain extended public key at HD tree depth 4
	MaxConsecutiveExternalIndex   uint32    `json:"max_consecutive_external_index"`   // Max consecutive index (without any gap) on the external chain
	InternalXPub                  string    `json:"internal_xpub"`                    // Internal chain extended public key at HD tree depth 4
	MaxConsecutiveInternalIndex   uint32    `json:"max_consecutive_internal_index"`   // Max consecutive index (without any gap) on the internal chain
	LookaheadSize                 uint32    `json:"lookahead_size"`                   // Numerical size of the lookahead zone
	Scheme                        Scheme    `json:"scheme"`                           // String identifier for keychain scheme
	Network                       Network   `json:"network"`                          // String denoting the network to use for encoding addresses
	NonConsecutiveExternalIndexes []uint32  `json:"non_consecutive_external_indexes"` // Used external indexes that are creating a gap in the derivation
	NonConsecutiveInternalIndexes []uint32  `json:"non_consecutive_internal_indexes"` // Used internal indexes that are creating a gap in the derivation
}

// Schema is a map between keychain ID and the keystore information.
type Schema map[uuid.UUID]*Meta

// Meta is a struct containing account details corresponding to a keychain ID,
// such as derivations, addresses, etc.
type Meta struct {
	Main        KeychainInfo              `json:"main"`
	Derivations map[DerivationPath]string `json:"derivations"` // public key at HD tree depth 5
	Addresses   map[string]DerivationPath `json:"addresses"`   // derivation path at HD tree depth 5
}

type FromChainCode struct {
	// Serialized public key associated with the extended key derived
	// at the account-level derivation path.
	//
	// Both compressed as well as uncompressed public keys are accepted.
	PublicKey []byte
	// Serialized chain code associated with the extended key derived at the
	// account-level derivation path.
	//
	// This field is 32 bytes long.
	ChainCode []byte
	// Index at BIP32 level 3.
	AccountIndex uint32
}

func (m *Meta) MarshalJSON() ([]byte, error) {
	// Step 1: Create type aliases of the original struct, including the
	// embedded one.

	type Alias Meta

	type KeychainInfoAlias KeychainInfo

	// Step 2: Create an anonymous struct with raw replacements for the special
	// fields.
	aux := &struct {
		Main struct {
			ID string `json:"id"`
			*KeychainInfoAlias
		} `json:"main"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	aux.Main.ID = m.Main.ID.String()

	// Step 3: Unmarshal the data into the anonymous struct.
	return json.Marshal(aux)
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	// Step 1: Create type aliases of the original struct, including the
	// embedded one.

	type Alias Meta

	type KeychainInfoAlias KeychainInfo

	// Step 2: Create an anonymous struct with raw replacements for the special
	// fields.
	aux := &struct {
		Derivations map[string]string `json:"derivations"`
		Addresses   map[string]string `json:"addresses"`
		Main        struct {
			ID string `json:"id"`
			*KeychainInfoAlias
		} `json:"main"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	// Step 3: Unmarshal the data into the anonymous struct.
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Step 4: Convert the raw fields to the desired types
	id, err := uuid.Parse(aux.Main.ID)
	if err != nil {
		return err
	}

	m.Main = *(*KeychainInfo)(aux.Main.KeychainInfoAlias)
	m.Main.ID = id

	m.Addresses = map[string]DerivationPath{}

	for k, v := range aux.Addresses {
		path := strings.Split(v, "/")

		changeIndex, err := strconv.Atoi(path[0])
		if err != nil {
			return err
		}

		addressIndex, err := strconv.Atoi(path[1])
		if err != nil {
			return err
		}

		m.Addresses[k] = DerivationPath{
			uint32(changeIndex), uint32(addressIndex),
		}
	}

	m.Derivations = map[DerivationPath]string{}

	for k, v := range aux.Derivations {
		path := strings.Split(k, "/")

		changeIndex, err := strconv.Atoi(path[0])
		if err != nil {
			return err
		}

		addressIndex, err := strconv.Atoi(path[1])
		if err != nil {
			return err
		}

		derivation := DerivationPath{
			uint32(changeIndex), uint32(addressIndex),
		}

		m.Derivations[derivation] = v
	}

	return nil
}

// AddressInfo encapsulates an address along with useful information associated
// to the address.
type AddressInfo struct {
	Address    string
	Derivation DerivationPath
	Change     Change
}

// ChangeXPub returns the ExtendedPublicKey of the keychain for the specified Change
// (Internal or External).
func (m Meta) ChangeXPub(change Change) (string, error) {
	switch change {
	case External:
		return m.Main.ExternalXPub, nil
	case Internal:
		return m.Main.InternalXPub, nil
	default:
		return "", errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

// MaxConsecutiveIndex returns the max consecutive index without any gap,
// for the specified Change (Internal or External).
func (m Meta) MaxConsecutiveIndex(change Change) (uint32, error) {
	switch change {
	case External:
		return m.Main.MaxConsecutiveExternalIndex, nil
	case Internal:
		return m.Main.MaxConsecutiveInternalIndex, nil
	default:
		return 0, errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

// SetMaxConsecutiveIndex updates the max consecutive index value for the
// specified Change (Internal or External).
func (m *Meta) SetMaxConsecutiveIndex(change Change, index uint32) error {
	switch change {
	case External:
		m.Main.MaxConsecutiveExternalIndex = index
	case Internal:
		m.Main.MaxConsecutiveInternalIndex = index
	default:
		return errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}

	return nil
}

// NonConsecutiveIndexes returns the non-consecutive indexes introduced due to
// gaps in derived addresses, for the specified Change (Internal or External).
func (m Meta) NonConsecutiveIndexes(change Change) ([]uint32, error) {
	switch change {
	case External:
		return m.Main.NonConsecutiveExternalIndexes, nil
	case Internal:
		return m.Main.NonConsecutiveInternalIndexes, nil
	default:
		return nil, errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

// SetNonConsecutiveIndexes updates the non-consecutive indexes for the
// specified Change (Internal or External).
//
// Any index less than the max consecutive index will be filtered out to handle
// the case when a previously introduced gap is filled.
func (m *Meta) SetNonConsecutiveIndexes(change Change, indexes []uint32) error {
	maxConsecutiveIndex, err := m.MaxConsecutiveIndex(change)
	if err != nil {
		return err
	}

	var result []uint32

	// Filter out all non-consecutive indexes less than the max consecutive
	// index.
	for _, i := range indexes {
		if i >= maxConsecutiveIndex {
			result = append(result, i)
		}
	}

	switch change {
	case External:
		m.Main.NonConsecutiveExternalIndexes = result
	case Internal:
		m.Main.NonConsecutiveInternalIndexes = result
	default:
		return errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}

	return nil
}

// MaxObservableIndex returns the maximum index inclusive of used and unused
// address indexes, for a given Change. It is therefore the maximum index
// that is currently observed by the keychain.
func (m Meta) MaxObservableIndex(change Change) (uint32, error) {
	switch change {
	case External:
		n := uint32(len(m.Main.NonConsecutiveExternalIndexes))
		return m.Main.MaxConsecutiveExternalIndex + n + m.Main.LookaheadSize, nil
	case Internal:
		n := uint32(len(m.Main.NonConsecutiveInternalIndexes))
		return m.Main.MaxConsecutiveInternalIndex + n + m.Main.LookaheadSize, nil
	default:
		return 0, errors.Wrapf(ErrUnrecognizedChange, fmt.Sprint(change))
	}
}

// ResetKeychainMeta resets the max consecutive indexes (external and interal), the derivations and addresses maps.
func (m *Meta) ResetKeychainMeta() {
	m.Main.MaxConsecutiveExternalIndex = 0
	m.Main.MaxConsecutiveInternalIndex = 0
	m.Derivations = map[DerivationPath]string{}
	m.Addresses = map[string]DerivationPath{}
}
