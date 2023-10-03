package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

var flagPath = flag.String("path", "", "path to tts mod code")

func main() {
	flag.Parse()
	// Listen for incoming connections.
	l, err := net.Listen("tcp", ":39998")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Working dir:", *flagPath)
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
	switch msg.MessageID {
	case 1:
		// Save all files
		for _, s := range msg.ScriptStates {
			name := fmt.Sprintf("%s.%s", s.Name, s.GUID)
			name = filepath.Join(*flagPath, name)
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
		// TODO handle removed files
	default:
		fmt.Println("unhandled messageID", msg.MessageID)
	}
}
