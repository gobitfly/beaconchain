package utils

// SafeCopyBytes will copy and return a non-nil byte array, otherwise it returns nil.
func SafeCopyBytes(cp []byte) []byte {
	if cp != nil {
		copied := make([]byte, len(cp))
		copy(copied, cp)
		return copied
	}
	return nil
}

func IsByteArrayAllZero(bytes []byte) bool {
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}
	return true
}
