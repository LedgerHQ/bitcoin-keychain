// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/ledgerhq/bitcoin-keychain/pb/keychain"
)

func TestP2PKHKeychainTest(t *testing.T) {
	ctx := context.Background()
	client, conn := keychainClient(ctx)
	defer conn.Close()

	info, err := client.CreateKeychain(ctx, &pb.CreateKeychainRequest{
		Account:       &pb.CreateKeychainRequest_ExtendedPublicKey{ExtendedPublicKey: BitcoinMainnetP2PKH.ExtendedPublicKey},
		LookaheadSize: 20,
		Network:       BitcoinMainnetP2PKH.Network,
		Scheme:        BitcoinMainnetP2PKH.Scheme,
	})

	if err != nil {
		t.Fatalf("failed to create keychain - error = %v", err)
	}

	gotObsAddrs, err := client.GetAllObservableAddresses(ctx, &pb.GetAllObservableAddressesRequest{
		KeychainId: info.KeychainId,
		Change:     pb.Change_CHANGE_EXTERNAL,
		FromIndex:  0,
		ToIndex:    10,
	})
	if err != nil {
		t.Fatalf("failed to get addresses in observable range [1 10] - error = %v", err)
	}

	wantObsAddrs := &pb.GetAllObservableAddressesResponse{Addresses: []*pb.AddressInfo{
		{
			Address:    "151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR",
			Derivation: []uint32{0, 0},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "18tMkbibtxJPQoTPUv8s3mSXqYzEsrbeRb",
			Derivation: []uint32{0, 1},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1GJr9FHZ1pbR4hjhX24M4L1BDUd2QogYYA",
			Derivation: []uint32{0, 2},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1KZB7aFfuZE2skJQPHH56VhSxUpUBjouwQ",
			Derivation: []uint32{0, 3},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1FyjDvDFcXLMmhMWD6u8bFovLgkhZabhTQ",
			Derivation: []uint32{0, 4},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1NGp18iPyWfSZz4AWnwT6HptDdVJfTjxnF",
			Derivation: []uint32{0, 5},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1L36ug5kWFLbMysfkAexh9LeicyMAteuEg",
			Derivation: []uint32{0, 6},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "169V9snkmcdzpEDhRyLMnEuhLKyWdjzhfd",
			Derivation: []uint32{0, 7},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "14K3JxsLwhpLiECaoJMsZYyk9peYP1Gtty",
			Derivation: []uint32{0, 8},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1GEix38AknUMWH8DYSn43HqodoB7RjyBAJ",
			Derivation: []uint32{0, 9},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
		{
			Address:    "1918hHSQNsNMRkDCUMy7DUmJ8GJzwfRkUV",
			Derivation: []uint32{0, 10},
			Change:     pb.Change_CHANGE_EXTERNAL,
		},
	}}

	if !proto.Equal(gotObsAddrs, wantObsAddrs) {
		t.Fatalf("GetAllObservableAddresses() got = '%v', want = '%v'",
			gotObsAddrs.Addresses, wantObsAddrs.Addresses)
	}

	gotReceiveAddr, err := client.GetFreshAddresses(
		ctx, &pb.GetFreshAddressesRequest{
			KeychainId: info.KeychainId,
			Change:     pb.Change_CHANGE_EXTERNAL,
			BatchSize:  1,
		})
	if err != nil {
		t.Fatalf("failed to get fresh external addr - error = %v", err)
	}

	expectedFreshAddress := "151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR"

	if gotReceiveAddr.Addresses[0].Address != expectedFreshAddress {
		t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
			gotReceiveAddr.Addresses[0].Address, expectedFreshAddress)
	}

	if _, err := client.MarkAddressesAsUsed(
		ctx, &pb.MarkAddressesAsUsedRequest{
			KeychainId: info.KeychainId,
			Addresses:  []string{"151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR"},
		}); err != nil {
		t.Fatalf("MarkAddressesAsUsed() - error = %v", err)
	}

	gotNextReceiveAddr, err := client.GetFreshAddresses(
		ctx, &pb.GetFreshAddressesRequest{
			KeychainId: info.KeychainId,
			Change:     pb.Change_CHANGE_EXTERNAL,
			BatchSize:  1,
		})
	if err != nil {
		t.Fatalf("failed to get fresh external addr - error = %v", err)
	}

	expectedNextFreshAddress := "18tMkbibtxJPQoTPUv8s3mSXqYzEsrbeRb"

	if gotNextReceiveAddr.Addresses[0].Address != expectedNextFreshAddress {
		t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
			gotNextReceiveAddr.Addresses[0].Address, expectedNextFreshAddress)
	}
}
