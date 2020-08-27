package keystore

import (
	"github.com/pkg/errors"
	"reflect"
	"testing"
)

func TestParseDescriptor(t *testing.T) {
	tests := []struct {
		name       string
		descriptor string
		want       DescriptorTokens
		wantErr    error
	}{
		{
			name:       "legacy",
			descriptor: "pkh(deadbeef)",
			want: DescriptorTokens{
				XPub:   "deadbeef",
				Scheme: "BIP44",
			},
		},
		{
			name:       "wrapped segwit",
			descriptor: "sh(wpkh(deadbeef))",
			want: DescriptorTokens{
				XPub:   "deadbeef",
				Scheme: "BIP49",
			},
		},
		{
			name:       "native segwit",
			descriptor: "wpkh(deadbeef)",
			want: DescriptorTokens{
				XPub:   "deadbeef",
				Scheme: "BIP84",
			},
		},
		{
			name:       "verbose wrapped segwit",
			descriptor: "sh(wpkh([d34db33f/44'/0'/0']deadbeef/1/*))",
			want: DescriptorTokens{
				XPub:   "deadbeef",
				Scheme: "BIP49",
			},
		},
		{
			name:       "invalid scheme",
			descriptor: "abcd(deadbeef)",
			wantErr:    ErrUnrecognizedScheme,
		},
		{
			name:       "invalid descriptor",
			descriptor: "wpkh(deadbeef",
			wantErr:    ErrInvalidDescriptor,
		},
		{
			name:       "empty descriptor",
			descriptor: "wpkh()",
			wantErr:    ErrInvalidDescriptor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDescriptor(tt.descriptor)
			if err != nil && tt.wantErr == nil {
				t.Fatalf("ParseDescriptor() unexpected error: %v", err)
			}

			if err == nil && tt.wantErr != nil {
				t.Fatalf("ParseDescriptor() got no error, want '%v'",
					tt.wantErr)
			}

			if err != nil && tt.wantErr != errors.Cause(err) {
				t.Fatalf("ParseDescriptor() got error = %v, want = %v",
					err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDescriptor() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
