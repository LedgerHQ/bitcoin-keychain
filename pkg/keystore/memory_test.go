package keystore

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/pkg/errors"

	"github.com/ledgerhq/bitcoin-keychain-svc/bitcoin"
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
	return &bitcoin.EncodeAddressResponse{
		Address: "deadbeef",
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
		name       string
		descriptor string
		want       KeychainInfo
		wantErr    error
	}{
		{
			name:       "invalid descriptor",
			descriptor: "bad xpub",
			wantErr:    ErrUnrecognizedScheme,
		},
		{
			name:       "native segwit",
			descriptor: "wpkh(xpub1111)",
			want: KeychainInfo{
				Descriptor:              "wpkh(xpub1111)",
				XPub:                    "xpub1111",
				SLIP32ExtendedPublicKey: "xpub1111",
				ExternalXPub:            "xpub1111->0",
				MaxExternalIndex:        0,
				InternalXPub:            "xpub1111->1",
				MaxInternalIndex:        0,
				LookaheadSize:           20,
				Scheme:                  "BIP84",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keystore := NewMockInMemoryKeystore()

			got, err := keystore.Create(tt.descriptor)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("Create() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && errors.Cause(err) != tt.wantErr {
				t.Fatalf("Create() got error '%v', want '%v'",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Create() got = '%v', want = '%v'",
					got, tt.want)
			}

			gotDB, dbErr := keystore.Get(tt.descriptor)
			if dbErr != nil && err == nil {
				t.Fatalf("Get() unexpected error: %v", dbErr)
			}

			if dbErr == nil && err != nil {
				t.Fatalf("Get() got no error, want '%v'",
					ErrDescriptorNotFound)
			}

			if dbErr != nil && errors.Cause(dbErr) != ErrDescriptorNotFound {
				t.Fatalf("Get() got error '%v', want '%v'",
					dbErr, ErrDescriptorNotFound)
			}

			if !reflect.DeepEqual(gotDB, tt.want) {
				t.Fatalf("Get() got = '%v', want = '%v'",
					gotDB, tt.want)
			}
		})
	}
}
