package main

import (
	"encoding/binary"
	"fmt"
	"github.com/arcaneiceman/GoVector/capture"
	"math/rand"
	"net"
	"os"
    "github.com/acarb95/DistributedAsserts/assert"
)

const (
	LARGEST_TERM = 100
	RUNS         = 500
)

var n int
var m int
var id          int                  //Id of this host
var hosts       int                  //number of hosts in the group
var neighbors = []string{}
const BASEPORT = 10000

func assertValue(values map[string]map[string]interface{}) bool {
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

		// Add requested value for current program, then for every other neighbor
		requestedValues := make(map[string][]string)
		requestedValues[":"+fmt.Sprintf("%d", BASEPORT+id+2*hosts)] = append(requestedValues[":"+fmt.Sprintf("%d", BASEPORT+id+2*hosts)], "inCritical")
		for _, v := range neighbors {
			requestedValues[v] = append(requestedValues[v], "inCritical")
		}

		// Assert on those requested things. 
		assert.Assert(assertValue, requestedValues)
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
