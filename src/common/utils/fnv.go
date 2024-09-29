package utils

func Fnv32a(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}

	return hash
}
