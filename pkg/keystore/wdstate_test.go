package keystore

import (
	"reflect"
	"testing"
)

func TestWDState(t *testing.T) {
	tests := []struct {
		base64 string
		state  WDKeychainState
	}{
		{
			// staging cc bitcoin0
			"AAAAAAQAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			WDKeychainState{
				maxConsecutiveChangeIndex:    4,
				maxConsecutiveReceiveIndex:   2,
				nonConsecutiveChangeIndexes:  make(map[uint32]bool),
				nonConsecutiveReceiveIndexes: make(map[uint32]bool),
				empty:                        false,
			},
		},
		{
			// handcraft complex object
			"AAAAAAIAAQAqAAAAAwAAAAAAAAABAAAAAgAAAAMAAAAEAAAAAAAAAAoAAAALAAAADAAAAA0AAAAA",
			WDKeychainState{
				maxConsecutiveChangeIndex:    65538,
				maxConsecutiveReceiveIndex:   42,
				nonConsecutiveChangeIndexes:  map[uint32]bool{1: true, 2: true, 3: true},
				nonConsecutiveReceiveIndexes: map[uint32]bool{10: true, 11: true, 12: true, 13: true},
				empty:                        false,
			},
		},
	}
	for _, test := range tests {
		state, err := ParseKeychainState(test.base64)
		if err != nil {
			t.Fatal("cannot parse")
		}

		if !reflect.DeepEqual(state, test.state) {
			t.Fatal("unexpected result for parse", state)
		}

		// we would like to compare encoded value to original input, but as map
		// is unordered, the base64 value is not stable
		// We encode and decode again
		base64, err := EncodeKeychainState(state)
		if err != nil {
			t.Fatal("cannot encode")
		}
		state, err = ParseKeychainState(base64)
		if err != nil {
			t.Fatal("cannot parse")
		}
		if !reflect.DeepEqual(state, test.state) {
			t.Fatal("unexpected result for parse", state)
		}
	}
}
