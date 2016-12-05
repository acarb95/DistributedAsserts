package main

import (
	"encoding/binary"
	"fmt"
	"github.com/arcaneiceman/GoVector/capture"
	"math/rand"
	"net"
	"os"
    "DistributedAsserts/assert"
)

const (
	LARGEST_TERM = 100
	RUNS         = 500
)

var n int
var m int

func assertValue(values map[string][string]int) bool {
	return true;
}

func main() {
	assert.InitDistributedAssert(":18589", []string{":9099"}, "client");
	assert.AddAssertable("n", &n, nil);
	assert.AddAssertable("m", &m, nil);
	localAddr, err := net.ResolveUDPAddr("udp4", ":18585")
	printErrAndExit(err)
	remoteAddr, err := net.ResolveUDPAddr("udp4", ":9090")
	printErrAndExit(err)
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	printErrAndExit(err)

	for t := 1; t <= RUNS; t++ {
		n, m := rand.Int()%LARGEST_TERM, rand.Int()%LARGEST_TERM
		sum, err := reqSum(conn, n, m)
		if err != nil {
			fmt.Printf("[CLIENT] %s", err.Error())
			continue
		}
		requestedValues := map[string][]string;
		assert.Assert(assertValue, )
		fmt.Printf("[CLIENT] %d/%d: %d + %d = %d\n", t, RUNS, n, m, sum)
	}
	fmt.Println()
	os.Exit(0)
}

func reqSum(conn *net.UDPConn, n, m int) (sum int64, err error) {
	msg := make([]byte, 256)
	binary.PutVarint(msg[:8], int64(n))
	binary.PutVarint(msg[8:], int64(m))

	fmt.Println(msg)

	// after instrumentation
	_, err = capture.Write(conn.Write, msg[:])
	// _, err = conn.Write(msg)
	if err != nil {
		return
	}

	//@dump

	buf := make([]byte, 256)
	// after instrumentation
	_, err = capture.Read(conn.Read, buf[:])
	// _, err = conn.Read(buf)
	if err != nil {
		return
	}

	sum, _ = binary.Varint(buf[0:])

	fmt.Println(sum, buf)

	//@dump

	return
}

func printErrAndExit(err error) {
	if err != nil {
		fmt.Printf("[CLIENT] %s\n" + err.Error())
		os.Exit(1)
	}
}
