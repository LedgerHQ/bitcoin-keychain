package keystore

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
)

// match https://github.com/LedgerHQ/lib-ledger-core/blob/4.2.0/core/src/wallet/bitcoin/keychains/CommonBitcoinLikeKeychains.hpp#L42
type WDKeychainState struct {
	maxConsecutiveChangeIndex    uint32
	maxConsecutiveReceiveIndex   uint32
	nonConsecutiveChangeIndexes  map[uint32]bool
	nonConsecutiveReceiveIndexes map[uint32]bool
	empty                        bool
}

func ParseKeychainState(b64pref string) (WDKeychainState, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64pref)
	if err != nil {
		return WDKeychainState{}, err
	}

	reader := bytes.NewReader(decoded)

	var version uint32
	var size uint64

	state := WDKeychainState{
		maxConsecutiveChangeIndex:    0,
		maxConsecutiveReceiveIndex:   0,
		nonConsecutiveChangeIndexes:  make(map[uint32]bool),
		nonConsecutiveReceiveIndexes: make(map[uint32]bool),
		empty:                        true,
	}

	err = binary.Read(reader, binary.LittleEndian, &version)
	if err != nil {
		return WDKeychainState{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &state.maxConsecutiveChangeIndex)
	if err != nil {
		return WDKeychainState{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &state.maxConsecutiveReceiveIndex)
	if err != nil {
		return WDKeychainState{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &size)
	if err != nil {
		return WDKeychainState{}, err
	}

	state.nonConsecutiveChangeIndexes, err = readSet(reader, size)
	if err != nil {
		return WDKeychainState{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &size)
	if err != nil {
		return WDKeychainState{}, err
	}

	state.nonConsecutiveReceiveIndexes, err = readSet(reader, size)
	if err != nil {
		return WDKeychainState{}, err
	}

	err = binary.Read(reader, binary.LittleEndian, &state.empty)
	if err != nil {
		return WDKeychainState{}, err
	}

	return state, nil
}

func EncodeKeychainState(state WDKeychainState) (string, error) {
	var buffer bytes.Buffer

	var version uint32 = 0
	err := binary.Write(&buffer, binary.LittleEndian, version)
	if err != nil {
		return "", err
	}

	err = binary.Write(&buffer, binary.LittleEndian, state.maxConsecutiveChangeIndex)
	if err != nil {
		return "", err
	}

	err = binary.Write(&buffer, binary.LittleEndian, state.maxConsecutiveReceiveIndex)
	if err != nil {
		return "", err
	}

	err = writeSet(&buffer, state.nonConsecutiveChangeIndexes)
	if err != nil {
		return "", err
	}

	err = writeSet(&buffer, state.nonConsecutiveReceiveIndexes)
	if err != nil {
		return "", err
	}

	empty := false
	err = binary.Write(&buffer, binary.LittleEndian, empty)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())

	return encoded, nil
}

func readSet(reader io.Reader, size uint64) (map[uint32]bool, error) {
	var i uint64
	ret := make(map[uint32]bool)
	for i = 0; i < size; i++ {
		var index uint32
		err := binary.Read(reader, binary.LittleEndian, &index)
		if err != nil {
			return ret, err
		}
		ret[index] = true
	}
	return ret, nil
}

func writeSet(buffer io.Writer, set map[uint32]bool) error {
	size := uint64(len(set))
	err := binary.Write(buffer, binary.LittleEndian, size)
	if err != nil {
		return err
	}

	for key := range set {
		err = binary.Write(buffer, binary.LittleEndian, key)
		if err != nil {
			return err
		}
	}
	return nil
}
