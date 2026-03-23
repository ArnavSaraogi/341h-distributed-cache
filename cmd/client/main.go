package main

import (
	cachering "distributedCache/cache_ring"
	"fmt"
	"log"
	"net"
)

/*
Note added a bunch of new lines since GPT was saying some bs about how testing is cleaner like this
since tcp requests can come in byte chunks (probably bullshit to be honest)
*/

//given list of strings ips and a list of keys (index maps 1 : 1) -> binary search (left dominant)

func main() {

	// TODO: GET LIST OF CACHES FROM CONFIG SERVICE
	ips := []string{}

	// add caches to the ring
	ring := cachering.NewRing()
	for i := 0; i < len(ips); i++ {
		ring.AddIP(ips[i])
	}

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
