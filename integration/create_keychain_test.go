// +build integration

package integration

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"

	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
)

func TestKeychainRegistration(t *testing.T) {
	ctx := context.Background()
	client, conn := keychainClient(ctx)
	defer conn.Close()

	tests := []struct {
		name            string
		fixture         Fixture
		externalAddress *pb.GetFreshAddressesResponse
		internalAddress *pb.GetFreshAddressesResponse
		wantErr         error
	}{
		{
			name:    "bitcoin mainnet p2pkh",
			fixture: BitcoinMainnetP2PKH,
			externalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR", Derivation: []uint32{0, 0}, Change: pb.Change_CHANGE_EXTERNAL},
				},
			},
			internalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "13hSrTAvfRzyEcjRcGS5gLEcNVNDhPvvUv", Derivation: []uint32{1, 0}, Change: pb.Change_CHANGE_INTERNAL},
				},
			},
		},
		{
			name:    "bitcoin testnet3 p2pkh",
			fixture: BitcoinTestnet3P2PKH,
			externalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "mkpZhYtJu2r87Js3pDiWJDmPte2NRZ8bJV", Derivation: []uint32{0, 0}, Change: pb.Change_CHANGE_EXTERNAL},
				},
			},
			internalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "mi8nhzZgGZQthq6DQHbru9crMDerUdTKva", Derivation: []uint32{1, 0}, Change: pb.Change_CHANGE_INTERNAL},
				},
			},
		},
		{
			name:    "bitcoin testnet3 p2sh-p2wpkh",
			fixture: BitcoinTestnet3P2SHP2WPKH,
			externalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "2MvuUMAG1NFQmmM69Writ6zTsYCnQHFG9BF", Derivation: []uint32{0, 0}, Change: pb.Change_CHANGE_EXTERNAL},
				},
			},
			internalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "2MsMvWTbPMg4eiSudDa5i7y8XNC8fLCok3c", Derivation: []uint32{1, 0}, Change: pb.Change_CHANGE_INTERNAL},
				},
			},
		},
		{
			name:    "bitcoin testnet3 p2wpkh",
			fixture: BitcoinMainnetP2WPKH,
			externalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "bc1qh4kl0a0a3d7su8udc2rn62f8w939prqpl34z86", Derivation: []uint32{0, 0}, Change: pb.Change_CHANGE_EXTERNAL},
				},
			},
			internalAddress: &pb.GetFreshAddressesResponse{
				Addresses: []*pb.AddressInfo{
					{Address: "bc1qry3crfssh8w6guajms7upclgqsfac4fs4g7nwj", Derivation: []uint32{1, 0}, Change: pb.Change_CHANGE_INTERNAL},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.CreateKeychain(ctx, &pb.CreateKeychainRequest{
				Account:       &pb.CreateKeychainRequest_ExtendedPublicKey{ExtendedPublicKey: tt.fixture.ExtendedPublicKey},
				LookaheadSize: 20,
				Network:       tt.fixture.Network,
				Scheme:        tt.fixture.Scheme,
			})
			if err != nil {
				t.Fatalf("failed to create keychain - error = %v", err)
			}

			wantKeychainInfo := &pb.KeychainInfo{
				KeychainId:              info.KeychainId,
				InternalDescriptor:      tt.fixture.InternalDescriptor,
				ExternalDescriptor:      tt.fixture.ExternalDescriptor,
				ExtendedPublicKey:       tt.fixture.ExtendedPublicKey,
				Slip32ExtendedPublicKey: tt.fixture.ExtendedPublicKey,
				LookaheadSize:           20,
				Scheme:                  tt.fixture.Scheme,
				Network:                 tt.fixture.Network,
			}

			if !proto.Equal(info, wantKeychainInfo) {
				t.Fatalf("CreateKeychain() info = '%v', want = '%v'",
					info, wantKeychainInfo)
			}

			gotExtAddr, err := client.GetFreshAddresses(
				ctx, &pb.GetFreshAddressesRequest{
					KeychainId: info.KeychainId,
					Change:     pb.Change_CHANGE_EXTERNAL,
					BatchSize:  1,
				})
			if err != nil {
				t.Fatalf("failed to get fresh external addr - error = %v", err)
			}

			if !proto.Equal(gotExtAddr, tt.externalAddress) {
				t.Fatalf("GetFreshAddresses() info = '%v', want = '%v'",
					gotExtAddr.Addresses, tt.externalAddress.Addresses)
			}

			gotIntAddr, err := client.GetFreshAddresses(
				ctx, &pb.GetFreshAddressesRequest{
					KeychainId: info.KeychainId,
					Change:     pb.Change_CHANGE_INTERNAL,
					BatchSize:  1,
				})
			if err != nil {
				t.Fatalf("failed to get fresh internal addr - error = %v", err)
			}

			if !proto.Equal(gotIntAddr, tt.internalAddress) {
				t.Fatalf("GetFreshAddresses() info = '%v', want = '%v'",
					gotIntAddr.Addresses, tt.internalAddress.Addresses)
			}

			// Mark first internal address as used
			_, err = client.MarkAddressesAsUsed(ctx, &pb.MarkAddressesAsUsedRequest{
				KeychainId: info.KeychainId,
				Addresses:  []string{gotIntAddr.Addresses[0].Address},
			})

			if err != nil {
				t.Fatalf("failed to mark first interal addr as used - error = %v", err)
			}

			// Check fresh addresses not return the mark as used one
			gotIntAddrAfterMark, err := client.GetFreshAddresses(
				ctx, &pb.GetFreshAddressesRequest{
					KeychainId: info.KeychainId,
					Change:     pb.Change_CHANGE_INTERNAL,
					BatchSize:  1,
				})

			if err != nil {
				t.Fatalf("failed to get fresh internal addr - error = %v", err)
			}

			nextIntFreshDerivationPath := gotIntAddrAfterMark.Addresses[0].Derivation
			expectedNextIntFreshDerivationPath := []uint32{1, 1}

			if !reflect.DeepEqual(nextIntFreshDerivationPath, expectedNextIntFreshDerivationPath) {
				t.Fatalf("Next fresh internal index info = '%v', want = '%v'",
					nextIntFreshDerivationPath, expectedNextIntFreshDerivationPath)
			}

			// Reset the keychain for this id
			_, err = client.ResetKeychain(ctx, &pb.ResetKeychainRequest{KeychainId: info.KeychainId})

			if err != nil {
				t.Fatalf("failed to reset keychain = %v", err)
			}

			// Check that fresh addresses after reset are good
			gotIntAddr, err = client.GetFreshAddresses(
				ctx, &pb.GetFreshAddressesRequest{
					KeychainId: info.KeychainId,
					Change:     pb.Change_CHANGE_INTERNAL,
					BatchSize:  1,
				})

			if err != nil {
				t.Fatalf("failed to get fresh internal addr - error = %v", err)
			}

			if !proto.Equal(gotIntAddr, tt.internalAddress) {
				t.Fatalf("GetFreshAddresses() info = '%v', want = '%v'",
					gotIntAddr.Addresses, tt.internalAddress.Addresses)
			}
		})
	}
}
