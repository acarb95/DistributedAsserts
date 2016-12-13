package main

import (
	"encoding/binary"
	"fmt"
	"github.com/arcaneiceman/GoVector/capture"
	"net"
	"os"
    "github.com/acarb95/DistributedAsserts/assert"
)

const addr = ":9090"

var a int64
var b int64
var sum int64

func main() {
	client_assert_addr := ":18589"
	server_assert_addr := ":9099"
	assert.InitDistributedAssert(server_assert_addr, []string{client_assert_addr}, "server");
	conn, err := net.ListenPacket("udp4", addr)
	if err != nil {
		fmt.Printf("[SERVER] %s\n", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	assert.AddAssertable("a", &a, nil);
	assert.AddAssertable("b", &b, nil);
	assert.AddAssertable("sum", &sum, nil)

	fmt.Printf("[SERVER] listening on %s\n", addr)

	//main loop
	for {
		err = listenAndRespond(conn)
		if err != nil {
			fmt.Printf("[SERVER] %s\n", err.Error())
		}
	}
}

// expects messages with two 64 bit integers (2 * 8 bytes)
func listenAndRespond(conn net.PacketConn) (err error) {
	buf := make([]byte, 256)

	// after instrumentation:
	_, addr, err := capture.ReadFrom(conn.ReadFrom, buf[0:])
	// fmt.Printf("Received Message [%d] from %s\n", n, addr.String())
	// _, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return
	}

	// var readA, readB int
	//@dump

	a, _ = binary.Varint(buf[:8])
	b, _ = binary.Varint(buf[8:])

	// a = a + 5 // Uncomment to force error on server side reading of variables

	sum = a + b - 3

	// fmt.Println(buf)
	// fmt.Println(buf[:8], a, readA)
	// fmt.Println(buf[8:], b, readB)

	fmt.Printf("[SERVER] %d + %d = %d\n", a, b, sum)

	msg := make([]byte, 32)
	binary.PutVarint(msg, sum) //putN := 

	// fmt.Println(putN, msg)

	// after instrumentation:
	capture.WriteTo(conn.WriteTo, msg, addr)
	// conn.WriteTo(msg, addr)

	//@dump

	return nil
}
