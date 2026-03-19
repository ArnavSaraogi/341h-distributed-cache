package main

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type SafeCache struct {
	mu sync.Mutex

	cache map[string]*list.Element

	//capacity of cache
	capacity int

	//doubly LL for LRU Policy -> holds keys so we know which kvp to remove
	lru_list *list.List
}

// this what the doubly ll stores in each node asw as the val in the map
type entry struct {
	key   string
	value string
}

func handleConnection(conn net.Conn, Cache *SafeCache) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}
		if n == 0 {
			return
		}
		req := string(buf[:n])
		req = strings.TrimSpace(req)
		parts := strings.Fields(req)
		cmd := parts[0]
		if cmd == "GET" {
			response := handleGet(parts, Cache)
			conn.Write([]byte(response))
		}
		if cmd == "PUT" {
			response := handlePut(parts, Cache)
			conn.Write([]byte(response))
		}
	}
}

func handlePut(parts []string, Cache *SafeCache) string {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()

	//TODO -> support better parsing (comment below) and invalid requests
	//must parse for all strings >=2 not just == 2 (im lazy rn)
	newkey := strings.TrimSpace(parts[1])
	newval := strings.TrimSpace(parts[2])

	if len(Cache.cache) == Cache.capacity {
		//prints are for debugging
		fmt.Println(Cache.cache)

		//evict last elemeent by grabbing it->deleting from cache->deleting from list
		lastElement := Cache.lru_list.Back()
		e := lastElement.Value.(entry)
		key := e.key
		delete(Cache.cache, key)
		Cache.lru_list.Remove(lastElement)

		//prints are for debugging
		fmt.Println(Cache.cache)
	}
	elem := Cache.lru_list.PushFront(entry{newkey, newval})
	Cache.cache[newkey] = elem
	fmt.Println(Cache.cache)
	return "ok\n"
}

func handleGet(parts []string, Cache *SafeCache) string {
	Cache.mu.Lock()
	defer Cache.mu.Unlock()
	key := strings.TrimSpace(parts[1])
	if e, ok := Cache.cache[key]; ok {

		//getting element->casting into entry struct -> getting val -> moving to front -> return val
		elem := e.Value.(entry)
		val := elem.value
		Cache.lru_list.MoveToFront(e)
		return val
	} else {
		return "no such key"
	}
}

func main() {
	cache := SafeCache{
		cache:    make(map[string]*list.Element),
		capacity: 2,
		lru_list: list.New(),
	}
	ln, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		log.Fatalln(err)
	}

	defer ln.Close()

	for {
		// wait for incoming client connections
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		// handle the connection
		go handleConnection(conn, &cache)
	}
}
