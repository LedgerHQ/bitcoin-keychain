package keystore

func contains(s []uint32, e uint32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func minUint32(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}

	return b
}
