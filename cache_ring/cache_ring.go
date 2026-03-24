package cache_ring

import (
	"hash/fnv"
	"slices"
	"strings"
	"sync"
)

type CacheRing struct {
	cache_hashes []uint32
	cache_ips    []string
	mtx          sync.Mutex
}

// create empty ring
func NewRing() *CacheRing {
	return &CacheRing{
		cache_hashes: []uint32{},
		cache_ips:    []string{},
	}
}

// TODO IF TIME: CREATE CONSTRUCTOR THAT TAKES A LIST

// reset ring
func (ring *CacheRing) ClearRing() {
	ring.mtx.Lock()
	ring.cache_hashes = nil
	ring.cache_ips = nil
	ring.mtx.Unlock()
}

// add cache to ring
func (ring *CacheRing) AddIP(ip string) {
	hash := hashIP(ip)

	ring.mtx.Lock()
	i := 0
	for ; i < len(ring.cache_hashes); i++ {
		if hash < ring.cache_hashes[i] {
			break
		}
	}
	ring.cache_hashes = slices.Insert(ring.cache_hashes, i, hash)
	ring.cache_ips = slices.Insert(ring.cache_ips, i, ip)
	ring.mtx.Unlock()
}

// remove specified cache from ring
func (ring *CacheRing) RemoveIP(ip string) {
	ring.mtx.Lock()
	i := 0
	for ; i < len(ring.cache_ips); i++ {
		if ip == ring.cache_ips[i] {
			break
		}
	}
	ring.cache_hashes = slices.Delete(ring.cache_hashes, i, i+1)
	ring.cache_ips = slices.Delete(ring.cache_ips, i, i+1)
	ring.mtx.Unlock()
}

// figure out which cache to put it in
func (ring *CacheRing) FindCache(key string) string {
	key = strings.Fields(key)[1]
	hashed_key := hashIP(key)

	ring.mtx.Lock()
	targ_idx := ring.binSearch(hashed_key)
	cache_ip := ring.cache_ips[targ_idx]
	ring.mtx.Unlock()

	return cache_ip
}

func (ring *CacheRing) binSearch(hashed_key uint32) int {
	// edge case if hashed_key is smallest
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
