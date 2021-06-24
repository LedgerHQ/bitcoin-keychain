package keystore

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ledgerhq/bitcoin-keychain/log"
	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"github.com/ledgerhq/bitcoin-keychain/pkg/chaincfg"
	"github.com/pkg/errors"
)

// Keystore is an interface that all keychain storage backends must implement.
// Currently, there are two Keystore implementations available:
//   InMemoryKeystore: useful for unit-tests
//   RedisKeystore:    for production
type Keystore interface {
	// Get returns a previously stored keychain information based on the provided
	// keychain UUID.
	//
	// Returns an error if the keychain UUID is missing in the keystore.
	Get(id uuid.UUID) (KeychainInfo, error)
	// Delete removes a keychain corresponding to a UUID from the keystore.
	Delete(id uuid.UUID) error
	// Reset removes derivations and addresses of a keychain corresponding to a UUID from the keystore.
	Reset(id uuid.UUID) error
	// Create parses a populates the keystore with the corresponding
	// keychain information, based on the provided extended public key, Scheme,
	// and Network information.
	//
	// Only initial state is populated, so no addresses will be inserted into the
	// keystore by this method.
	Create(extendedPublicKey string, fromChainCode *FromChainCode, scheme Scheme, net chaincfg.Network,
		lookaheadSize uint32, index uint32, metadata string) (KeychainInfo, error)
	// GetFreshAddress retrieves an unused address from the keystore at a
	// given Change index, for the keychain corresponding to the provided keychain
	// ID.
	//
	// See GetFreshAddresses for getting fresh addresses in bulk, and for further
	// details.
	GetFreshAddress(id uuid.UUID, change Change) (*AddressInfo, error)
	// GetFreshAddresses retrieves bulk fresh addresses from the keystore.
	//
	// In addition to ensuring that issued addresses are always fresh (unused), the
	// method also detects gaps in used addresses and includes it in fresh address
	// list.
	GetFreshAddresses(id uuid.UUID, change Change, size uint32) ([]AddressInfo, error)
	// MarkPathAsUsed sets a given derivation path as used. It records bookkeeping
	// information about gaps in the derivation.
	//
	// A derivation path is considered "used" if it has transaction history.
	//
	// If marking of a derivation path as used introduces any gaps, they are
	// detected and saved in the keystore. For this we rely on two main fields:
	//   MaxConsecutiveIndex   -> the largest consecutive index without any gaps
	//   NonConsecutiveIndexes -> list of used indexes that introduced gaps
	MarkPathAsUsed(id uuid.UUID, path DerivationPath) error
	// MarkAddressAsUsed is a helper to directly mark an address as used. It
	// internally fetches the derivation path of the address from the keystore,
	// and then marks this DerivationPath value as used.
	MarkAddressAsUsed(id uuid.UUID, address string) error
	GetAllObservableAddresses(id uuid.UUID, change Change,
		fromIndex uint32, toIndex uint32) ([]AddressInfo, error)
	// GetDerivationPath reads the address-to-derivations mapping in the keystore,
	// and returns the DerivationPath corresponding to the specified address.
	GetDerivationPath(id uuid.UUID, address string) (DerivationPath, error)
	// GetAddressesPublicKeys reads the derivation-to-publicKey mapping in the keystore,
	// and returns extendend public keys corresponding to given derivations.
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

// KeychainInfo models the global information related to an account registered
// in the keystore.
//
// Rather than using the associated gRPC message struct, it is defined here
// independently to avoid having gRPC dependency in this package.
type KeychainInfo struct {
	ID                            uuid.UUID        `json:"id"`                               // Keychain ID as a uuid.UUID type
	ExternalDescriptor            string           `json:"external_descriptor"`              // External chain output descriptor
	InternalDescriptor            string           `json:"internal_descriptor"`              // Internal chain output descriptor
	ExtendedPublicKey             string           `json:"extended_public_key"`              // Extended public key serialized with standard HD version bytes
	SLIP32ExtendedPublicKey       string           `json:"slip32_extended_public_key"`       // Extended public key serialized with SLIP-0132 HD version bytes
	ExternalXPub                  string           `json:"external_xpub"`                    // External chain extended public key at HD tree depth 4
	MaxConsecutiveExternalIndex   uint32           `json:"max_consecutive_external_index"`   // Max consecutive index (without any gap) on the external chain
	InternalXPub                  string           `json:"internal_xpub"`                    // Internal chain extended public key at HD tree depth 4
	MaxConsecutiveInternalIndex   uint32           `json:"max_consecutive_internal_index"`   // Max consecutive index (without any gap) on the internal chain
	LookaheadSize                 uint32           `json:"lookahead_size"`                   // Numerical size of the lookahead zone
	Scheme                        Scheme           `json:"scheme"`                           // String identifier for keychain scheme
	Network                       chaincfg.Network `json:"network"`                          // String denoting the network to use for encoding addresses
	NonConsecutiveExternalIndexes []uint32         `json:"non_consecutive_external_indexes"` // Used external indexes that are creating a gap in the derivation
	NonConsecutiveInternalIndexes []uint32         `json:"non_consecutive_internal_indexes"` // Used internal indexes that are creating a gap in the derivation
	AccountIndex                  uint32           `json:"account_index"`                    // Account index
	Metadata                      string           `json:"metadata"`                         // Additional info, unspecified
}

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

func keystoreCreate(
	extendedPublicKey string,
	fromChainCode *FromChainCode,
	scheme Scheme,
	net chaincfg.Network,
	lookaheadSize uint32,
	index uint32,
	metadata string,
	client bitcoin.CoinServiceClient,
) (Meta, error) {
	if fromChainCode != nil {
		res, err := GetAccountExtendedKey(client, net, fromChainCode)
		if err != nil {
			return Meta{}, errors.Wrapf(err,
				"failed to get extendend public key from chain code, request = %v", fromChainCode)
		}

		extendedPublicKey = res.ExtendedKey
	}

	internalDescriptor, err := MakeDescriptor(extendedPublicKey, Internal, scheme)
	if err != nil {
		return Meta{}, errors.Wrapf(err,
			"failed to make internal descriptor, xkey = %v", extendedPublicKey)
	}

	externalDescriptor, err := MakeDescriptor(extendedPublicKey, External, scheme)
	if err != nil {
		return Meta{}, errors.Wrapf(err,
			"failed to make internal descriptor, xkey = %v", extendedPublicKey)
	}

	externalChild, err := childKDF(client, extendedPublicKey, 0)
	if err != nil {
		return Meta{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", extendedPublicKey, 0)
	}

	internalChild, err := childKDF(client, extendedPublicKey, 1)
	if err != nil {
		return Meta{}, errors.Wrapf(
			err, "failed to derive xpub %v at index %v", extendedPublicKey, 1)
	}

	id := uuid.New()

	keychainInfo := KeychainInfo{
		ID:                          id,
		InternalDescriptor:          internalDescriptor,
		ExternalDescriptor:          externalDescriptor,
		ExtendedPublicKey:           extendedPublicKey,
		SLIP32ExtendedPublicKey:     extendedPublicKey, // TODO: Convert ExtendedPublicKey to SLIP-0132 form
		ExternalXPub:                externalChild.ExtendedKey,
		MaxConsecutiveExternalIndex: 0,
		InternalXPub:                internalChild.ExtendedKey,
		MaxConsecutiveInternalIndex: 0,
		LookaheadSize:               lookaheadSize,
		Scheme:                      scheme,
		Network:                     net,
		AccountIndex:                index,
		Metadata:                    metadata,
	}

	meta := Meta{
		Main:        keychainInfo,
		Derivations: map[DerivationPath]string{},
		Addresses:   map[string]DerivationPath{},
	}

	return meta, nil
}

func (m *Meta) keystoreGetFreshAddresses(
	client bitcoin.CoinServiceClient,
	change Change,
	size uint32,
) ([]AddressInfo, error) {
	addrs := []AddressInfo{}
	maxConsecutiveIndex, err := m.MaxConsecutiveIndex(change)
	if err != nil {
		return addrs, err
	}

	nonConsecutiveIndexes, err := m.NonConsecutiveIndexes(change)
	if err != nil {
		return nil, err
	}

	for i := uint32(0); uint32(len(addrs)) < size; i++ {
		index := maxConsecutiveIndex + i

		// Skip any index that exists in non-consecutive indexes, to prevent
		// address reuse.
		if !contains(nonConsecutiveIndexes, index) {
			path := DerivationPath{uint32(change), index}

			addr, err := deriveAddress(client, m, path)
			if err != nil {
				return addrs, err
			}

			addrInfo := AddressInfo{
				Address:    addr,
				Derivation: path,
				Change:     change,
			}

			addrs = append(addrs, addrInfo)
		}
	}
	return addrs, nil
}

func (m *Meta) keystoreMarkPathAsUsed(path DerivationPath) error {
	change := path.ChangeIndex()

	maxConsecutiveIndex, err := m.MaxConsecutiveIndex(change)
	if err != nil {
		return err
	}

	nonConsecutiveIndexes, err := m.NonConsecutiveIndexes(change)
	if err != nil {
		return err
	}

	switch {
	// CASE 1: Address index being marked as used already falls within the
	// range of consecutive indexes. This is typically when an address index
	// is marked as used twice.
	case path.AddressIndex() < maxConsecutiveIndex:
		// Nothing to do in this case.

	// CASE 2: Mark as used at the boundary of the consecutive used indexes, by
	// incrementing the max consecutive index.
	case path.AddressIndex() == maxConsecutiveIndex:
		maxConsecutiveIndex++

		// Handle case when the max consecutive index overreaches into the
		// non-consecutive indexes. This typically happens when a gap is filled
		// by marking the gap address index as used.
		//
		// Repeat this step until the max consecutive index is outside the
		// overlap of non-consecutive indexes, where it is safe to issue a
		// fresh address.
		for contains(nonConsecutiveIndexes, maxConsecutiveIndex) {
			maxConsecutiveIndex++
		}

		// Save the max consecutive index. It is important to perform this step
		// before saving the non-consecutive indexes.
		if err := m.SetMaxConsecutiveIndex(change, maxConsecutiveIndex); err != nil {
			return err
		}

		// Reconcile non-consecutive indexes depending on the updated state of
		// the max consecutive index.
		//
		// TODO: Implement a dedicated method for this reconciliation, since the
		//       non-consecutive indexes are never changed until this step.
		if err := m.SetNonConsecutiveIndexes(change, nonConsecutiveIndexes); err != nil {
			return err
		}

	// CASE 3: Attempt to introduce a gap after the max consecutive index.
	//
	// Consider the following list of address indexes as the state of the
	// keychain. * indicates that an index is used.
	//
	// Before:
	//   state                  : 0*  1  2  3  4  5
	//   max consecutive index  : 1
	//   non-consecutive indexes: []
	//
	// After: mark index 2 as used
	//   state                  : 0*  1  2*  3  4  5
	//   max consecutive index  : 1   <- no change
	//   non-consecutive indexes: [2] <- add the index that created a gap
	case path.AddressIndex() > maxConsecutiveIndex:
		// Add address index to list of non-consecutive indexes (if does not
		// exist already).
		if !contains(nonConsecutiveIndexes, path.AddressIndex()) {
			nonConsecutiveIndexes = append(
				nonConsecutiveIndexes, path.AddressIndex())

			err := m.SetNonConsecutiveIndexes(change, nonConsecutiveIndexes)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Meta) keystoreGetAllObservableAddresses(
	client bitcoin.CoinServiceClient,
	change Change,
	fromIndex uint32,
	toIndex uint32,
) ([]AddressInfo, error) {
	maxObservableIndex, err := m.MaxObservableIndex(change)
	if err != nil {
		return nil, err
	}

	length := minUint32(toIndex, maxObservableIndex) - fromIndex

	log.WithFields(log.Fields{
		"maxObservable": maxObservableIndex,
		"givenRange":    []uint32{fromIndex, toIndex},
		"computedRange": []uint32{fromIndex, fromIndex + length},
	}).Info("[keystore] GetAllObservableAddresses: compute range")

	addrs := []AddressInfo{}

	for i := fromIndex; i <= fromIndex+length; i++ {
		path := DerivationPath{uint32(change), i}

		addr, err := deriveAddress(client, m, path)
		if err != nil {
			return nil, err
		}

		addrInfo := AddressInfo{
			Address:    addr,
			Derivation: path,
			Change:     change,
		}

		addrs = append(addrs, addrInfo)
	}

	return addrs, nil
}

func (m *Meta) keystoreGetDerivationPath(address string) (DerivationPath, error) {
	path, ok := m.Addresses[address]
	if !ok {
		return DerivationPath{}, ErrAddressNotFound
	}
	return path, nil
}

func keystoreMarkAddressAsUsed(s Keystore, id uuid.UUID, address string) error {
	path, err := s.GetDerivationPath(id, address)
	if err != nil {
		return err
	}

	return s.MarkPathAsUsed(id, path)
}

func (m *Meta) keystoreGetAddressesPublicKeys(derivations []DerivationPath) ([]string, error) {
	publicKeys := make([]string, len(derivations))

	for idx, derivation := range derivations {
		publicKey, ok := m.Derivations[derivation]

		if !ok {
			return nil, ErrDerivationNotFound
		}

		publicKeys[idx] = publicKey
	}

	return publicKeys, nil
}
