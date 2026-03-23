package cachering

import (
	"hash/fnv"
	"slices"
)

// have list of cache ips

type CacheRing struct {
	cache_hashes []int
	cache_ips    []int
}

// ghetto ass bin search
func (ring *CacheRing) binSearch(hashed_key int) int {
	//edge case if hashed_key is smallest
	if hashed_key < slices.Min(ring.cache_hashes) {
		return len(ring.cache_hashes) - 1
	}
	l := 0
	r := len(ring.cache_hashes) - 1
	idx := l
	for l <= r {
		mid := l + (r-l)/2
		if ring.cache_hashes[mid] == hashed_key {
			idx = mid
			return idx
		}
		if ring.cache_hashes[mid] < hashed_key {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	idx = l - 1
	return max(idx, 0)
}

func (ring *CacheRing) hashKey(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s)) // write to buffer
	return h.Sum32()
}

func (ring *CacheRing) FindCache(key string) int {
	hashed_key := ring.hashKey(key)

	cache_idx := ring.binSearch(int(hashed_key))

	return ring.cache_ips[cache_idx]
}
