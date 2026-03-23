package main

import "distributedCache/cache_ring"

/*
Note added a bunch of new lines since GPT was saying some bs about how testing is cleaner like this
since tcp requests can come in byte chunks (probably bullshit to be honest)
*/

//given list of strings ips and a list of keys (index maps 1 : 1) -> binary search (left dominant)

func main() {

	// TODO: GET LIST OF CACHES FROM CONFIG SERVICE
	ips := []string{}

	// add caches to the ring
	ring := cache_ring.NewRing()
	for i := 0; i < len(ips); i++ {
		ring.AddIP(ips[i])
	}

}
