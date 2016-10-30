package main

import (
    "fmt"
    "net"
    "os"
    "flag"
    "strconv"
    "bufio"
    "time"
    "strings"
    "github.com/arcaneiceman/GoVector/capture"
    "github.com/arcaneiceman/GoVector/govec"
    "github.com/acarb95/DistributedAsserts/assert"
)

// ============================================ STRUCTS ============================================
type Neighbor struct {
    address string
}

type Message struct {
    MessageType int
    ObjectName string
    RequestingNode string
    SendingNode string
    Result interface{}
    RoundNumber int
}

type AssertInfo struct {
    RequestingNode string
    RoundNumber int
}

type AssertObject struct {
    Object *assert.AssertableObject
    ActiveAsserts []AssertInfo
}

// ============================================= CONST =============================================
// Message types
const (
    RUN_ASSERT = iota
    RETURN_RESULT
)

// =======================================  GLOBAL VARIABLES =======================================
var address string
var neighborsfile string = ""
var logfile string
var neighbors []Neighbor
var debug bool = false
var LOG *govec.GoLog
var listener *net.UDPConn
var testAssertable *assert.AssertableObject = assert.create_assertable_object("test", 0)
var process_name string = ""

// ========================================  HELPER METHODS ========================================
func usage() {
    fmt.Println("USAGE: program <process_name> <ip:port> <neighborsfile> <logfile>")
    fmt.Println("Where each argument is required")
    fmt.Println("\tprocess_name: the name of the process for the log file.")
    fmt.Println("\tip_port: should be in the form of \"ip:port\". This is the address the " + 
                "node will use to communicate with the neighbors.")
    fmt.Println("\tneighborsfile: a text file with the addresses for each neighbors per line. Should " +
                "be in the form of \"ip:port\".")
    fmt.Println("\tlogfile: filename for the GoVector log.")
}


func checkResult(err error) {
    if err != nil {
        fmt.Println("ERROR: ", err)
        os.Exit(-1)
    }
}

func getArgs(args []string) {
    var err error;
    if len(args) != 4 {
        fmt.Println("ERROR: must have four arguments")
        masterUsage()
        os.Exit(-1)
    } else {
        process_name = args[0]
        address = args[1]
        neighborsfile = args[2]
        logfile = args[3]
    }
}

// Function prints the master variables
func printArgs() {
    fmt.Println("Process Name: ", process_name)
    fmt.Println("Address: ", address)
    fmt.Println("Neighbors File: ", neighborsfile)
    fmt.Println("Log File: ", logfile)
}

func parseNeighborsFile() {
    // Get slave connections from file
    file, err := os.Open(neighborsfile)
    checkResult(err)
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        neighbor_address := scanner.Text()
        neighbors = append(neighbors, Neighbor{neighbor_address})
        if debug {
            fmt.Println("Added ", neighbor_address)
        }
    }
    checkResult(scanner.Err())
    if debug {
        fmt.Println("Final Struct Array: ", neighbors)
    }
}

// ========================================= TIMER METHODS =========================================
func timeSyncTimer() {
    // Set off alarm when T seconds have elapses and clocks should sync
    // Use a channel, just have another process that will send the DATEREQ messages
    for {
        timer := time.NewTimer(time.Minute * time.Duration(T))
        <- timer.C
        // Timer done, call send DATEREQ message
        sync_round++
        logMessage := fmt.Sprintf("Sending assertion message")
        broadcastMessage(Message{MessageType: RUN_ASSERT, Name: "test", RequestingNode: address, SendingNode: address}, logMessage)
        frozen_process_time := process_time
        time.Sleep(time.Second)
    }
}

func TickTock() {
    previous := time.Now().Unix()
    for {
        current := time.Now().Unix()
        // time.Sleep(time.Second)
        if (current - previous  != 0) {
            process_time += int(current - previous)
            if (process_time % 10 == 0) {
                logMessage := fmt.Sprintf("Current time: %d", process_time)
                fmt.Println(logMessage)
                // LOG.LogLocalEvent(logMessage)
            }
        }
        previous = current
    }
}

// ===================================== COMMUNICATION METHODS =====================================
func broadcastMessage(payload Message, logMessage string) {
    for i := range(neighbors) {
        go sendToAddr(payload, neighbors[i].address, logMessage)
    }
}

func sendToAddr(payload Message, addr string, logMessage string) {
    address, err := net.ResolveUDPAddr("udp", addr)
    checkResult(err)

    if debug {
        fmt.Println(logMessage)
        fmt.Printf("Attempting to send [MessageType: %d, Time: %d, Address: %s] to %s\n", 
                   payload.MessageType, payload.Time, payload.Address, address)
    }

    capture.WriteToUDP(listener.WriteToUDP, LOG.PrepareSend(logMessage, payload), address)
}

func receiveConnections() chan Message {
    msg := make(chan Message)

    buf := make([]byte, 1024)

    go func() {
        for {
            n, addr, err := capture.ReadFromUDP(listener.ReadFromUDP, buf[0:])
            var incomingMessage Message
            LOG.UnpackReceive("Received Message From Node", buf[0:n], &incomingMessage)
            logMessage := fmt.Sprintf("Received message [MessageType: %d] from [%s]",
                                      incomingMessage.MessageType,
                                      addr)
            LOG.LogLocalEvent(logMessage)
            if err != nil {
                fmt.Println("READ ERROR: ", err)
                break;
            }
            if debug {
                fmt.Printf("Received message [MessageType: %d] from [%s]\n",
                           incomingMessage.MessageType,
                           addr)
            }
            msg <- incomingMessage
        }
    }()
    return msg
}

func processData(message_chan chan Message){
    go func() {
        for {
            message := <- message_chan
            msg_type := message.MessageType

            // Switch on the message type byte, each case should do it's own parsing with the buffer
            switch msg_type {
            case RUN_ASSERT:
                // TODO how to tell it is the first assert message for an object --> store requesting node, round number
                // Call run assert code for specific assert
                assertable = findAssertObj(message.Name)
                // LOCK OBJECT
                assertable.Lock.Lock()
                result = assertable.assert()
                // Send result back
                sendToAddr(Message{MessageType: RETURN_RESULT, SendingNode: address, Result: result}, message.RequestingNode)
                broadcastMessage(Message{MessageType: RUN_ASSERT, RequestingNode: message.RequestingNode, SendingNode: address})
                break
            case RETURN_RESULT:
                // Process returned result
                result = message.Result
                break
            default:
                fmt.Printf("Error: unknown message type received [%d]\n", msg_type)
            }
        }
    } ()
}


// ============================================= MAIN ==============================================
func main() {
    getArgs(os.Args[1:])
    if debug {
        printArgs()
    }

    if debug {
        fmt.Println("Starting timer tick. Should tick every second")
    }

    go TickTock()

    if debug {
     fmt.Printf("Initializing GoVector Log with name [%s] and log file [%s]\n", process_name, logfile)
    }

    LOG = govec.Initialize(process_name, logfile)

    if debug {
        fmt.Println("Opening listener with address ", address)
    }
    // Open port to listen on
    listen_address, err := net.ResolveUDPAddr("udp", address)
    listener, err = net.ListenUDP("udp", listen_address)
    if &listener == nil {
        fmt.Println("Error could not listen on ", address)
        fmt.Println("Error: ", err)
        os.Exit(-1)
    }
    defer listener.Close()

    parseNeighborsFile()

    if debug {
        fmt.Printf("Starting timer for syncs every %d minutes\n", T)
    }
    go timeSyncTimer()

    if debug {
        fmt.Println("Calling receive connections go routine")
    }

    // Separate goroutine to handle client connections
    message := receiveConnections()

    if debug {
        fmt.Println("Calling process data")
    }

    processData(message)

    fmt.Println("Type \"exit\" to quit.")

    for {
        reader := bufio.NewReader(os.Stdin)
        text, _ := reader.ReadString('\n')
        if strings.EqualFold(text, "exit\n") {
            break
        }
    }
    fmt.Println("Finished executing")
}
