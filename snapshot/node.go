package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
    "time"
    "strings"
    "github.com/arcaneiceman/GoVector/capture"
    "github.com/arcaneiceman/GoVector/govec"
    "github.com/acarb95/DistributedAsserts/asserts"
)

// ============================================ STRUCTS ============================================
type Neighbor struct {
    address string
}

type Message struct {
    MessageType int
    Name string
    RequestingNode string
    SendingNode string
    RoundNumber int
    Result interface{}
}

type AssertInfo struct {
    RequestingNode string
    RoundNumber int
}

type AssertObject struct {
    Object asserts.AssertableObject
    ActiveAsserts []AssertInfo
}

// ============================================= CONST =============================================
// Message types
const (
    RUN_ASSERT = iota
    RETURN_RESULT
    COMPLETE
)

// =======================================  GLOBAL VARIABLES =======================================
var address string
var neighborsfile string = ""
var logfile string
var neighbors []Neighbor
var debug bool = false
var LOG *govec.GoLog
var listener *net.UDPConn
var process_name string = ""
var assertableDictionary map[string]AssertObject

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
    if len(args) != 4 {
        fmt.Println("ERROR: must have four arguments")
        usage()
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

func populateGlobalAssertObjects() {
    testAssertable := asserts.CreateAssertableObject("test", 0)
    infoArray := []AssertInfo{}
    assertableDictionary["test"] = AssertObject{ Object: testAssertable, ActiveAsserts: infoArray }
}

// ========================================= TIMER METHODS =========================================
func timeSyncTimer() {
    // Set off alarm when T seconds have elapses and clocks should sync
    // Use a channel, just have another process that will send the DATEREQ messages
    for {
        timer := time.NewTimer(time.Minute * time.Duration(1))
        <- timer.C
        // Timer done, call send DATEREQ message
        logMessage := fmt.Sprintf("Sending assertion message")
        broadcastMessage(Message{MessageType: RUN_ASSERT, Name: "test", RequestingNode: address, SendingNode: address}, logMessage)
        time.Sleep(time.Second)
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
        fmt.Printf("Attempting to send [MessageType: %d] to %s\n", 
                   payload.MessageType, address)
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
                // assertable, assertInfo := findAssertObj(message.Name)
                // Check active asserts to see if you should run the assert
                // if running the assert, add to active asserts
                // LOCK OBJECT
                // assertable.Lock.Lock()
                // result = assertable.assert()
                // Send result back
                // sendToAddr(Message{MessageType: RETURN_RESULT, SendingNode: address, Result: result}, message.RequestingNode)
                // broadcastMessage(Message{MessageType: RUN_ASSERT, RequestingNode: message.RequestingNode, SendingNode: address})
                break
            case RETURN_RESULT:
                // Process returned result
                // result := message.Result
                // Once have all results (need other method to handle) --> send complete method
                break
            case COMPLETE:
                // Remove [message.RequestingNode, message.RoundNumber] from active asserts list
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
        fmt.Printf("Starting timer for syncs every %d minutes\n", 1)
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
