// +build !integration

package keystore

import (
	"context"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/pkg/errors"

	"github.com/ledgerhq/bitcoin-keychain/pb/bitcoin"
	"google.golang.org/grpc"
)

type mockBitcoinClient struct{}

func (c mockBitcoinClient) ValidateAddress(
	ctx context.Context,
	in *bitcoin.ValidateAddressRequest,
	opts ...grpc.CallOption,
) (*bitcoin.ValidateAddressResponse, error) {
	return &bitcoin.ValidateAddressResponse{
		Address: in.Address,
		IsValid: true,
	}, nil
}

func (c mockBitcoinClient) DeriveExtendedKey(
	ctx context.Context,
	in *bitcoin.DeriveExtendedKeyRequest,
	opts ...grpc.CallOption,
) (*bitcoin.DeriveExtendedKeyResponse, error) {
	extendedKey := in.ExtendedKey
	publicKey := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	chainCode := []byte{0xCA, 0xFE, 0xBA, 0xBE}

	for _, index := range in.Derivation {
		extendedKey += "->" + strconv.Itoa(int(index))
		publicKey = append(publicKey, byte(index))
		chainCode = append(chainCode, byte(index))
	}

	return &bitcoin.DeriveExtendedKeyResponse{
		ExtendedKey: extendedKey,
		PublicKey:   publicKey,
		ChainCode:   chainCode,
	}, nil
}

func (c mockBitcoinClient) EncodeAddress(
	ctx context.Context,
	in *bitcoin.EncodeAddressRequest,
	opts ...grpc.CallOption,
) (*bitcoin.EncodeAddressResponse, error) {
	net, err := networkFromChainParams(in.ChainParams)
	if err != nil {
		panic(err)
	}

	scheme, err := schemeFromProtoEncoding(in.Encoding)
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("%s-%s-%s",
		hex.EncodeToString(in.PublicKey), scheme, net)

	return &bitcoin.EncodeAddressResponse{
		Address: addr,
	}, nil
}

func NewMockInMemoryKeystore() Keystore {
	return &InMemoryKeystore{
		db:     Schema{},
		client: mockBitcoinClient{},
	}
}

func TestInMemoryKeystore_GetCreate(t *testing.T) {
	tests := []struct {
		name        string
		extendedKey string
		scheme      Scheme
		network     Network
		want        KeychainInfo
		wantErr     error
	}{
		{
			name:        "native segwit",
			extendedKey: "xpub1111",
			scheme:      BIP84,
			network:     Mainnet,
			want: KeychainInfo{
				ExternalDescriptor:          "wpkh(xpub1111/0/*)",
				InternalDescriptor:          "wpkh(xpub1111/1/*)",
				ExtendedPublicKey:           "xpub1111",
				SLIP32ExtendedPublicKey:     "xpub1111",
				ExternalXPub:                "xpub1111->0",
				MaxConsecutiveExternalIndex: 0,
				InternalXPub:                "xpub1111->1",
				MaxConsecutiveInternalIndex: 0,
				LookaheadSize:               20,
				Scheme:                      "BIP84",
				Network:                     Mainnet,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keystore := NewMockInMemoryKeystore()

			gotInfo, err := keystore.Create(
				tt.extendedKey, tt.scheme, tt.network, DefaultLookaheadSize)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("Create() gotInfo no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && errors.Cause(err) != tt.wantErr {
				t.Fatalf("Create() gotInfo error '%v', want '%v'",
					err, tt.wantErr)
			}

			// Do not compare UUIDs, since it random.
			tt.want.ID = gotInfo.ID

			if !reflect.DeepEqual(gotInfo, tt.want) {
				t.Fatalf("Create() gotInfo = '%v', want = '%v'",
					gotInfo, tt.want)
			}

			gotDB, dbErr := keystore.Get(gotInfo.ID)
			if dbErr != nil && err == nil {
				t.Fatalf("Get() unexpected error: %v", dbErr)
			}

			if dbErr == nil && err != nil {
				t.Fatalf("Get() gotInfo no error, want '%v'",
					ErrKeychainNotFound)
			}

			if dbErr != nil && errors.Cause(dbErr) != ErrKeychainNotFound {
				t.Fatalf("Get() gotInfo error '%v', want '%v'",
					dbErr, ErrKeychainNotFound)
			}

			if !reflect.DeepEqual(gotDB, tt.want) {
				t.Fatalf("Get() gotInfo = '%v', want = '%v'",
					gotDB, tt.want)
			}
		})
	}
}

func TestInMemoryKeystore_GetFreshAddress(t *testing.T) {
	tests := []struct {
		name        string
		extendedKey string
		scheme      Scheme
		change      Change
		network     Network
		want        string
		wantErr     error
	}{
		{
			name:        "p2pkh mainnet",
			extendedKey: "xpub1111",
			scheme:      BIP84,
			change:      External,
			network:     Mainnet,
			want:        "deadbeef00-BIP84-mainnet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keystore := NewMockInMemoryKeystore()
			info, err := keystore.Create(
				tt.extendedKey, tt.scheme, tt.network, DefaultLookaheadSize)
			if err != nil {
				panic(err)
			}

			got, err := keystore.GetFreshAddress(info.ID, tt.change)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("GetFreshAddress() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("GetFreshAddress() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && errors.Cause(err) != tt.wantErr {
				t.Fatalf("GetFreshAddress() got error '%v', want '%v'",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetFreshAddress() got = '%v', want = '%v'",
					got, tt.want)
			}
		})
	}
}

func TestInMemoryKeystore_GetFreshAddresses(t *testing.T) {
	tests := []struct {
		name        string
		extendedKey string
		change      Change
		scheme      Scheme
		network     Network
		size        uint32
		want        []string
		wantErr     error
	}{
		{
			name:        "empty",
			extendedKey: "xpub1111",
			scheme:      BIP84,
			change:      External,
			network:     Mainnet,
			size:        0,
			want:        []string{},
		},
		{
			name:        "p2pkh mainnet multi",
			extendedKey: "xpub1111",
			scheme:      BIP84,
			change:      External,
			network:     Mainnet,
			size:        5,
			want: []string{
				"deadbeef00-BIP84-mainnet",
				"deadbeef01-BIP84-mainnet",
				"deadbeef02-BIP84-mainnet",
				"deadbeef03-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keystore := NewMockInMemoryKeystore()
			info, err := keystore.Create(
				tt.extendedKey, tt.scheme, tt.network, DefaultLookaheadSize)
			if err != nil {
				panic(err)
			}

			got, err := keystore.GetFreshAddresses(info.ID, tt.change, tt.size)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("GetFreshAddresses() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("GetFreshAddresses() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && errors.Cause(err) != tt.wantErr {
				t.Fatalf("GetFreshAddresses() got error '%v', want '%v'",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
					got, tt.want)
			}
		})
	}
}

func TestInMemoryKeystore_MarkPathAsUsed(t *testing.T) {
	keystore := NewMockInMemoryKeystore()

	info, err := keystore.Create(
		"xpub1111", BIP84, Mainnet, DefaultLookaheadSize)
	if err != nil {
		panic(err)
	}

	workflow := []struct {
		name               string
		path               DerivationPath
		change             Change
		size               uint32
		wantFreshAddresses []string
		wantFreshAddress   string
		wantErr            error
	}{
		{
			name:   "mark 0/0 as used",
			path:   DerivationPath{0, 0},
			change: External,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef01-BIP84-mainnet", // should have no gaps
				"deadbeef02-BIP84-mainnet",
				"deadbeef03-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef01-BIP84-mainnet",
		},
		{
			name:   "mark 0/2 as used",
			path:   DerivationPath{0, 2}, // introduce a gap
			change: External,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef01-BIP84-mainnet", // should detect the gap
				"deadbeef03-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
				"deadbeef06-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef01-BIP84-mainnet",
		},
		{
			name:   "mark 0/1 as used",
			path:   DerivationPath{0, 1}, // fill the gap
			change: External,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef03-BIP84-mainnet", // should have no gaps
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
				"deadbeef06-BIP84-mainnet",
				"deadbeef07-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef03-BIP84-mainnet",
		},
		{
			// internal chain should be unaffected by previous mutations
			name:   "mark 1/0 as used",
			path:   DerivationPath{1, 0},
			change: Internal,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef01-BIP84-mainnet",
				"deadbeef02-BIP84-mainnet",
				"deadbeef03-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef01-BIP84-mainnet",
		},
		{
			name:   "mark 1/3 as used",
			path:   DerivationPath{1, 3},
			change: Internal,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef01-BIP84-mainnet",
				"deadbeef02-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
				"deadbeef06-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef01-BIP84-mainnet",
		},
		{
			name:   "mark 1/6 as used",
			path:   DerivationPath{1, 6},
			change: Internal,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef01-BIP84-mainnet",
				"deadbeef02-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
				"deadbeef07-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef01-BIP84-mainnet",
		},
		{
			name:   "mark 1/1 as used",
			path:   DerivationPath{1, 1},
			change: Internal,
			size:   5,
			wantFreshAddresses: []string{
				"deadbeef02-BIP84-mainnet",
				"deadbeef04-BIP84-mainnet",
				"deadbeef05-BIP84-mainnet",
				"deadbeef07-BIP84-mainnet",
				"deadbeef08-BIP84-mainnet",
			},
			wantFreshAddress: "deadbeef02-BIP84-mainnet",
		},
	}

	for _, tt := range workflow {
		t.Run(tt.name, func(t *testing.T) {
			if err := keystore.MarkPathAsUsed(info.ID, tt.path); err != nil {
				t.Fatalf("MarkPathAsUsed() unexpected error: %v", err)
			}

			gotBulk, err := keystore.GetFreshAddresses(info.ID, tt.change, tt.size)
			if err != nil {
				t.Fatalf("GetFreshAddresses() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(gotBulk, tt.wantFreshAddresses) {
				t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
					gotBulk, tt.wantFreshAddresses)
			}

			got, err := keystore.GetFreshAddress(info.ID, tt.change)
			if err != nil {
				t.Fatalf("GetFreshAddress() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.wantFreshAddress) {
				t.Fatalf("GetFreshAddress() got = '%v', want = '%v'",
					got, tt.wantFreshAddress)
			}
		})
	}
}

func TestInMemoryKeystore_GetAddressesPublicKeys(t *testing.T) {
	tests := []struct {
		name        string
		extendedKey string
		change      Change
		scheme      Scheme
		network     Network
		size        uint32
		derivations []DerivationPath
		want        []string
		wantErr     error
	}{
		{
			name:        "p2pkh mainnet multi (change: external)",
			extendedKey: "xpub1111",
			change:      External,
			scheme:      BIP84,
			network:     Mainnet,
			size:        5,
			derivations: []DerivationPath{
				{0, 0},
				{0, 1},
				{0, 2},
				{0, 3},
				{0, 4},
			},
			want: []string{
				"deadbeef00",
				"deadbeef01",
				"deadbeef02",
				"deadbeef03",
				"deadbeef04",
			},
		},
		// internal chain should return the same public keys
		{
			name:        "p2pkh mainnet multi (change: internal)",
			extendedKey: "xpub1111",
			change:      Internal,
			scheme:      BIP84,
			network:     Mainnet,
			size:        5,
			derivations: []DerivationPath{
				{1, 0},
				{1, 1},
				{1, 2},
				{1, 3},
				{1, 4},
			},
			want: []string{
				"deadbeef00",
				"deadbeef01",
				"deadbeef02",
				"deadbeef03",
				"deadbeef04",
			},
		},
		{
			name:        "p2pkh mainnet multi (wrong given derivations)",
			extendedKey: "xpub1111",
			change:      Internal,
			scheme:      BIP84,
			network:     Mainnet,
			size:        5,
			derivations: []DerivationPath{
				{1, 0},
				{1, 6},
				{1, 7},
				{1, 8},
				{1, 9},
			},
			wantErr: ErrDerivationNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keystore := NewMockInMemoryKeystore()
			info, err := keystore.Create(
				tt.extendedKey, tt.scheme, tt.network, DefaultLookaheadSize)
			if err != nil {
				panic(err)
			}

			// Firstly, call get fresh addresses to derive addresses
			_, err = keystore.GetFreshAddresses(info.ID, tt.change, tt.size)
			if err != nil {
				t.Fatalf("GetFreshAddresses() unexpected error: %v", err)
			}

			got, err := keystore.GetAddressesPublicKeys(info.ID, tt.derivations)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("GetAddressesPublicKeys() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("GetAddressesPublicKeys() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && errors.Cause(err) != tt.wantErr {
				t.Fatalf("GetAddressesPublicKeys() got error '%v', want '%v'",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetAddressesPublicKeys() got = '%v', want = '%v'",
					got, tt.want)
			}
		})
	}
}
