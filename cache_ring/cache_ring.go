package cachering

import (
	"hash/fnv"
	"slices"
)

type CacheRing struct {
	cache_hashes []uint32
	cache_ips    []string
}

// create empty ring
func NewRing() *CacheRing {
	return &CacheRing{
		cache_hashes: []uint32{},
		cache_ips:    []string{},
	}
}

// TODO IF TIME: CREATE CONSTRUCTOR THAT TAKES A LIST

// add cache to ring
func (ring *CacheRing) AddIP(ip string) {
	hash := hashIP(ip)
	i := 0
	for ; i < len(ring.cache_hashes); i++ {
		if hash < ring.cache_hashes[i] {
			break
		}
	}
	ring.cache_hashes = slices.Insert(ring.cache_hashes, i, hash)
	ring.cache_ips = slices.Insert(ring.cache_ips, i, ip)
}

// remove specified cache from ring
func (ring *CacheRing) RemoveIP(ip string) {
	hash := hashIP(ip)
	i := 0
	for ; i < len(ring.cache_hashes); i++ {
		if hash == ring.cache_hashes[i] {
			break
		}
	}
	ring.cache_hashes = slices.Delete(ring.cache_hashes, i, i+1)
	ring.cache_ips = slices.Delete(ring.cache_ips, i, i+1)
}

// figure out which cache to put it in
func (ring *CacheRing) FindCache(key string) int {
	hashed_key := hashIP(key)
	return ring.binSearch(hashed_key)
}

// ghetto ass binary search
func (ring *CacheRing) binSearch(hashed_key uint32) int {
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

func hashIP(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s)) // write to buffer
	return h.Sum32()
}
