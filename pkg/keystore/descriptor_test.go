package keystore

import (
	"reflect"
	"testing"
)

func TestParseDescriptor(t *testing.T) {
	tests := []struct {
		name       string
		descriptor string
		want       DescriptorTokens
		err        string
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
			err:        "unrecognized scheme",
		},
		{
			name:       "invalid descriptor",
			descriptor: "wpkh(deadbeef",
			err:        "invalid descriptor",
		},
		{
			name:       "empty descriptor",
			descriptor: "wpkh()",
			err:        "invalid descriptor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDescriptor(tt.descriptor)
			if err != nil {
				if tt.err == "" {
					t.Fatalf("ParseDescriptor: unexpected error - %v", err)
				}

				if tt.err != err.Error() {
					t.Fatalf("ParseDescriptor: expected error '%v', got '%v'", tt.err, err)
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDescriptor() = %v, want %v", got, tt.want)
			}
		})
	}
}
