package assert

import (
"fmt"
"net"
"os"
"time"
"reflect"
"github.com/arcaneiceman/GoVector/capture"
"github.com/arcaneiceman/GoVector/govec"
)

// ============================================= CONST =============================================
// Message types
type messageType int
const (
    RTT_REQUEST messageType = iota
    RTT_RETURN = iota
    ASSERT_REQUEST = iota
    ASSERT_RETURN = iota
)

// ============================================ STRUCTS ============================================

type assertionFunction func(map[string]nameToValueMap)bool
type processFunction func(interface{})interface{}
type nameToValueMap map[string]interface{}


type message struct {
    MessageType messageType
    RequestingNode string
    RoundNumber int
    MessageTime time.Time
    Result interface{}
}

// =======================================  GLOBAL VARIABLES =======================================

var address string
var neighbors []string
var listener *net.UDPConn
var assertableDictionary nameToValueMap
var assertableFunctions map[string]processFunction
var roundToResponseMap map[int]*map[string]nameToValueMap
var roundTripTime map[string]time.Duration
var LOG *govec.GoLog
var roundNumber = 0;
var debug = true

// ========================================  HELPER METHODS ========================================

func checkResult(err error) {
    if err != nil {
        fmt.Println("ERROR: ", err)
        os.Exit(-1)
    }
}

// ===================================== COMMUNICATION METHODS =====================================

func broadcastMessage(payload message, logMessage string) {
    for _, v := range neighbors {
        go sendToAddr(payload, v, logMessage)
    }
}

func sendToAddr(payload message, addr string, logMessage string) {
    address, err := net.ResolveUDPAddr("udp", addr)
    checkResult(err)

    if debug {
        fmt.Println(logMessage)
        fmt.Printf("Attempting to send [MessageType: %d] to %s\n", 
         payload.MessageType, address)
    }
    capture.WriteToUDP(listener.WriteToUDP, LOG.PrepareSend(logMessage, payload), address)
}

// ===================================== RTT METHODS =====================================

func getRTT(addr string) {
	RTTmessage := message{MessageType: RTT_REQUEST, RequestingNode:address};
	broadcastMessage(RTTmessage, "Round Trip Request");

}

// ===================================== TIMING METHODS =====================================

func getTime() time.Time {
	return time.Now();
}

func getAssertDelay() time.Duration {
    duration := 0 * time.Second;
    for _, v := range roundTripTime {
        if (v > duration) {
            duration = v;
        }
    }
    return duration + 50 * time.Millisecond
}


// =======================================  PUBLIC METHODS =======================================

func InitDistributedAssert(addr string, neighbours []string, processName string) {
	address = addr;
	neighbors = neighbours;
    listen_address, err := net.ResolveUDPAddr("udp", address)
    listener, err = net.ListenUDP("udp", listen_address)
    if &listener == nil {
        fmt.Println("Error could not listen on ", address)
        fmt.Println("Error: ", err)
        os.Exit(-1)
    }
    defer listener.Close()

    LOG = govec.Initialize(processName, "logfile")

    assertableDictionary = make(map[string]interface{});
    assertableFunctions = make(map[string]processFunction);
}

func AddAssertable(name string, pointer interface{}, f processFunction) {
	assertableDictionary[name] = pointer;
    assertableFunctions[name] = f;
}

func Assert(f assertionFunction, requestedValues map[string][]string) {
    localRoundNumber := roundNumber;
    roundNumber++;

    maxRTT := getAssertDelay();
    responseMap := make(map[string]nameToValueMap);
    roundToResponseMap[localRoundNumber] = &responseMap;

    for k, v := range requestedValues {
        assertTime := getTime();
        assertTime = assertTime.Add(maxRTT);
        AssertRequestMessage :=  message{MessageType: ASSERT_REQUEST, RequestingNode: address, RoundNumber: localRoundNumber, MessageTime: assertTime, Result: v};
        go sendToAddr(AssertRequestMessage, k, "Requesting Assertion")
    }

    time.Sleep(maxRTT);

    if (len(requestedValues[address]) > 0) {
        responseMap[address] = make(map[string]interface{});
    }

    for _, v := range requestedValues[address] {
        f, ok := assertableFunctions[v]
        localVal := assertableDictionary[v];
        if (ok) {
            localVal = f(reflect.ValueOf(localVal));
        } 
        responseMap[address][v] = reflect.ValueOf(localVal);
    }

    time.Sleep(maxRTT);
    delete(roundToResponseMap, localRoundNumber);

    if (!f(responseMap)) {
        fmt.Println("ASSERTION FAILED: ", responseMap)
        os.Exit(-1)
    }
}