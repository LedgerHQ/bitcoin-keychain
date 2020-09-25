// +build !integration

package keystore

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestMakeDescriptor(t *testing.T) {
	tests := []struct {
		name              string
		extendedPublicKey string
		scheme            Scheme
		change            Change
		want              string
		wantErr           error
	}{
		{
			name:              "legacy",
			extendedPublicKey: "deadbeef",
			scheme:            BIP44,
			change:            External,
			want:              "pkh(deadbeef/0/*)",
		},
		{
			name:              "wrapped segwit",
			extendedPublicKey: "deadbeef",
			scheme:            BIP49,
			change:            External,
			want:              "sh(wpkh(deadbeef/0/*))",
		},
		{
			name:              "native segwit",
			extendedPublicKey: "deadbeef",
			scheme:            BIP84,
			change:            External,
			want:              "wpkh(deadbeef/0/*)",
		},
		{
			name:              "native segwit",
			extendedPublicKey: "deadbeef",
			scheme:            BIP84,
			change:            Internal,
			want:              "wpkh(deadbeef/1/*)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeDescriptor(tt.extendedPublicKey, tt.change, tt.scheme)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("MakeDescriptor() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("MakeDescriptor() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && tt.wantErr != errors.Cause(err) {
				t.Fatalf("MakeDescriptor() got error = %v, want = %v",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeDescriptor() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
