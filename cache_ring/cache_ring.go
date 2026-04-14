package cache_ring

import (
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/cespare/xxhash/v2"
)

type CacheRing struct {
	cache_hashes []uint32
	cache_ips    []string
	mtx          sync.Mutex
}

/* create empty ring */
func NewRing() *CacheRing {
	return &CacheRing{
		cache_hashes: []uint32{},
		cache_ips:    []string{},
	}
}

/* reset ring */
func (ring *CacheRing) ClearRing() {
	ring.mtx.Lock()
	ring.cache_hashes = nil
	ring.cache_ips = nil
	ring.mtx.Unlock()
}

/* add cache to ring */
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

/* remove specified cache from ring */
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

/* figure out which cache the key goes in */
func (ring *CacheRing) FindCache(key string) string {

	// handle case where no caches
	if len(ring.cache_hashes) == 0 {
		log.Fatal("No caches in cache ring")
	}

	key = strings.Fields(key)[1] // gets the key from the GET / PUT request (ie GET Jayleen -> Jayleen)
	hashed_key := hashIP(key)

	ring.mtx.Lock()
	targ_idx := ring.binSearch(hashed_key)
	cache_ip := ring.cache_ips[targ_idx]
	ring.mtx.Unlock()

	log.Printf("Key %s has hash %d, going in cache with ip %s and hash %d", key, hashed_key, ring.cache_ips[targ_idx], ring.cache_hashes[targ_idx])

	return cache_ip
}

/* does binary search to find the nearest cache with hash <= the key hash*/
func (ring *CacheRing) binSearch(hashed_key uint32) int {
	l := 0
	r := len(ring.cache_hashes)

	for l < r {
		mid := (l + r) / 2
		if ring.cache_hashes[mid] < hashed_key {
			l = mid + 1
		} else {
			r = mid
		}
	}

	// wrap around
	if l == len(ring.cache_hashes) {
		return 0
	}

	return l
}

/* hashes IPs and request keys */
func hashIP(s string) uint32 {
	return uint32(xxhash.Sum64String(s))
}

/* debug print to see caches and cache ips */
func (ring *CacheRing) PrintCachesIPsAndHashes() {
	for i := 0; i < len(ring.cache_hashes); i++ {
		log.Printf("Cache IP: %s\nCache Hash: %d", ring.cache_ips[i], ring.cache_hashes[i])
	}
}
