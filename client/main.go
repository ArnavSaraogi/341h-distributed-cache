package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
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
