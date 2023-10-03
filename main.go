package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", ":39998")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on :39998")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

type msg struct {
	MessageID    int
	SavePath     string
	ScriptStates []struct {
		Name   string
		GUID   string
		Script string
		UI     string
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Read the incoming connection into the buffer.
	var msg msg
	err := json.NewDecoder(conn).Decode(&msg)
	if err != nil {
		log.Fatal("Error reading:", err)
	}
	// Send a response back to person contacting us.
	fmt.Println("received msg ID", msg.MessageID)
	// Close the connection when you're done with it.
	conn.Close()
	// err = os.WriteFile("output", buf.Bytes(), os.ModePerm)
	// if err != nil {
	// panic(err)
	// }
	switch msg.MessageID {
	case 1:
		// Save all files
		for _, s := range msg.ScriptStates {
			name := fmt.Sprintf("/home/tom/src/tts/shatterpoint-tts/%s.%s", s.Name, s.GUID)
			if s.Script != "" {
				err := os.WriteFile(name+".lua", []byte(s.Script), os.ModePerm)
				if err != nil {
					panic(err)
				}
			}
			if s.UI != "" {
				err := os.WriteFile(name+".xml", []byte(s.UI), os.ModePerm)
				if err != nil {
					panic(err)
				}
			}
		}
	default:
		fmt.Println("unhandled messageID", msg.MessageID)
	}

	// fmt.Println(n, buf.Len())
}
