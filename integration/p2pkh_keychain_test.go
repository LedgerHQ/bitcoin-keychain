// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/ledgerhq/bitcoin-keychain-svc/pb/keychain"
)

func TestP2PKHKeychainTest(t *testing.T) {
	ctx := context.Background()
	client, conn := keychainSvcClient(ctx)
	defer conn.Close()

	if _, err := client.CreateKeychain(ctx, &pb.CreateKeychainRequest{
		AccountDescriptor: BitcoinMainnetP2PKH.Descriptor,
		LookaheadSize:     20,
		Network:           BitcoinMainnetP2PKH.Network,
	}); err != nil {
		t.Fatalf("failed to create keychain - error = %v", err)
	}

	gotObsAddrs, err := client.GetAllObservableAddresses(ctx, &pb.GetAllObservableAddressesRequest{
		AccountDescriptor: BitcoinMainnetP2PKH.Descriptor,
		Change:            pb.Change_CHANGE_EXTERNAL,
		FromIndex:         0,
		ToIndex:           10,
	})
	if err != nil {
		t.Fatalf("failed to get addresses in observable range [1 10] - error = %v", err)
	}

	wantObsAddrs := &pb.GetAllObservableAddressesResponse{Addresses: []string{
		"151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR",
		"18tMkbibtxJPQoTPUv8s3mSXqYzEsrbeRb",
		"1GJr9FHZ1pbR4hjhX24M4L1BDUd2QogYYA",
		"1KZB7aFfuZE2skJQPHH56VhSxUpUBjouwQ",
		"1FyjDvDFcXLMmhMWD6u8bFovLgkhZabhTQ",
		"1NGp18iPyWfSZz4AWnwT6HptDdVJfTjxnF",
		"1L36ug5kWFLbMysfkAexh9LeicyMAteuEg",
		"169V9snkmcdzpEDhRyLMnEuhLKyWdjzhfd",
		"14K3JxsLwhpLiECaoJMsZYyk9peYP1Gtty",
		"1GEix38AknUMWH8DYSn43HqodoB7RjyBAJ",
		"1918hHSQNsNMRkDCUMy7DUmJ8GJzwfRkUV",
	}}

	if !proto.Equal(gotObsAddrs, wantObsAddrs) {
		t.Fatalf("GetAllObservableAddresses() got = '%v', want = '%v'",
			gotObsAddrs.Addresses, wantObsAddrs.Addresses)
	}

	gotReceiveAddr, err := client.GetFreshAddresses(
		ctx, &pb.GetFreshAddressesRequest{
			AccountDescriptor: BitcoinMainnetP2PKH.Descriptor,
			Change:            pb.Change_CHANGE_EXTERNAL,
			BatchSize:         1,
		})
	if err != nil {
		t.Fatalf("failed to get fresh external addr - error = %v", err)
	}

	if gotReceiveAddr.Addresses[0] != "151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR" {
		t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
			gotReceiveAddr.Addresses, []string{"151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR"})
	}

	if _, err := client.MarkAddressesAsUsed(
		ctx, &pb.MarkAddressesAsUsedRequest{
			AccountDescriptor: BitcoinMainnetP2PKH.Descriptor,
			Addresses:         []string{"151krzHgfkNoH3XHBzEVi6tSn4db7pVjmR"},
		}); err != nil {
		t.Fatalf("MarkAddressesAsUsed() - error = %v", err)
	}

	gotNextReceiveAddr, err := client.GetFreshAddresses(
		ctx, &pb.GetFreshAddressesRequest{
			AccountDescriptor: BitcoinMainnetP2PKH.Descriptor,
			Change:            pb.Change_CHANGE_EXTERNAL,
			BatchSize:         1,
		})
	if err != nil {
		t.Fatalf("failed to get fresh external addr - error = %v", err)
	}

	if gotNextReceiveAddr.Addresses[0] != "18tMkbibtxJPQoTPUv8s3mSXqYzEsrbeRb" {
		t.Fatalf("GetFreshAddresses() got = '%v', want = '%v'",
			gotNextReceiveAddr.Addresses, []string{"18tMkbibtxJPQoTPUv8s3mSXqYzEsrbeRb"})
	}
}
