package main

import (
	"fmt"
	"net"
	"time"
)

// основной TCP хандлер
func handleConnection(conn net.Conn, chmu chan<- int) {
	bufr := make([]byte, 2000)
	bufS := []byte("GET /request HTTP 1.1")
	_, _ = conn.Write(bufS[:])
	rlen, _ := conn.Read(bufr)
	fmt.Println(string(bufr[:rlen]))
}
func main() {
	var mytime time.Time
	chu := make(chan int)
	mytime = time.Now()
	for i := 0; i < 10000; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:80")
		if err != nil {
			fmt.Println("Connect error:", err.Error())
		} else {
			go handleConnection(conn, chu)
		}
	}
	ddur := time.Since(mytime)
	sec := int(ddur.Seconds())
	fmt.Println("Seconds:", sec)
	rps := 10000 / sec
	fmt.Println("rps:", rps)
}
