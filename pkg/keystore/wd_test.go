// +build !integration

package keystore

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestWd_keychainInfoToWDKey(t *testing.T) {
	tests := []struct {
		input KeychainInfo
		want  WdKey
		err   error
	}{
		{
			input: KeychainInfo{Metadata: ""},
			want:  WdKey{},
			err:   errors.New(""),
		},
		{
			input: KeychainInfo{
				Metadata:     "libcore_prefix:ledger1",
				Scheme:       "BIP44",
				Network:      "bitcoin_mainnet",
				AccountIndex: 42,
			},
			want: WdKey{
				Prefix:     "libcore_prefix",
				Workspace:  "ledger1",
				WalletType: "bitcoin",
				Index:      42,
			},
			err: nil,
		},
		{
			input: KeychainInfo{
				Metadata:     "libcore_prefix:ledger1",
				Scheme:       "BIP44",
				Network:      "visa",
				AccountIndex: 42,
			},
			want: WdKey{},
			err:  errors.New(""),
		},
	}
	for _, tt := range tests {
		got, err := keychainInfoToWDKey(tt.input)
		if tt.err != nil && err == nil {
			t.Fatalf("error expected")
		}
		if tt.err == nil && err != nil {
			t.Fatalf("unexpected error")
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("get %v, want %v", got, tt.want)
		}
	}
}
