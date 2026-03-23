package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"slices"
)

//given list of strings ips and a list of keys (index maps 1 : 1) -> binary search (left dominant)

func binSearch(cache_hashes []int, hashed_key int) int {
	//edge case if hashed_key is smallest
	if hashed_key < slices.Min(cache_hashes) {
		return len(cache_hashes) - 1
	}
	l := 0
	r := len(cache_hashes) - 1
	idx := l
	for l <= r {
		mid := l + (r-l)/2
		if cache_hashes[mid] == hashed_key {
			idx = mid
			return idx
		}
		if cache_hashes[mid] < hashed_key {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	idx = l - 1
	return max(idx, 0)
}

func hashKey(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func findCache(cache_hashes []int, key string, cache_ips []int) int {
	hashed_key := hashKey(key)

	cache_idx := binSearch(cache_hashes, int(hashed_key))

	targ_ip := cache_ips[cache_idx]
	return targ_ip
}

func main() {

	//biin search fucntionality test
	var numbers []int

	numbers = append(numbers, 1)
	numbers = append(numbers, 4)
	numbers = append(numbers, 5)
	numbers = append(numbers, 6)
	numbers = append(numbers, 9)
	fmt.Println("arr: ", numbers)

	hk := 9

	idx := binSearch(numbers, hk)
	fmt.Println("idx: ", idx)

	// create TCP socket
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	// send message
	fmt.Fprintf(conn, "PUT apples banana\n")

	// read message
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(buf[:n]))

	fmt.Fprintf(conn, "GET apples\n")

	// read message
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf[:n]))

	fmt.Fprintf(conn, "PUT Jayleen Sauce\n")

	// read message
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf[:n]))

	fmt.Fprintf(conn, "PUT Mikey Sanjiv\n")

	// read message
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf[:n]))

	fmt.Fprintf(conn, "PUT Yesjiv Cockjiv\n")

	// read message
	buf = make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf[:n]))
	conn.Close()
}
