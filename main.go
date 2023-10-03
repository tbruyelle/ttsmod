package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var flagPath = flag.String("path", ".", "path to tts mod code")

func main() {
	flag.Parse()
	if flag.Arg(0) == "sync" {
		err := loadFiles()
		if err != nil {
			panic(err)
		}
		return
	}
	// Listen for incoming connections.
	l, err := net.Listen("tcp", ":39998")
	if err != nil {
		panic(err)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Working dir:", *flagPath)
	fmt.Println("Listening on :39998")
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func loadFiles() error {
	// Prepare msg
	entries, err := os.ReadDir(*flagPath)
	if err != nil {
		return err
	}
	objs := make(map[string]ScriptState)
	for _, e := range entries {
		parts := strings.Split(e.Name(), ".")
		// fmt.Println(e.Name(), parts)
		if e.IsDir() || len(parts) != 3 {
			continue
		}
		if len(parts) != 3 {
			return fmt.Errorf("bad filename %s", e.Name())
		}
		bz, err := os.ReadFile(filepath.Join(*flagPath, e.Name()))
		if err != nil {
			return err
		}
		ss := objs[parts[1]]
		ss.Name = parts[0]
		ss.GUID = parts[1]
		switch parts[2] {
		case "xml":
			ss.UI = string(bz)
		case "lua":
			ss.Script = string(bz)
		default:
			return fmt.Errorf("bad extension %s for file %s", parts[2], e.Name())
		}
		objs[parts[1]] = ss
	}
	if len(objs) == 0 {
		fmt.Println("Nothing to load")
		return nil
	}
	msg := msg{
		MessageID: 1,
	}
	for _, v := range objs {
		msg.ScriptStates = append(msg.ScriptStates, v)
	}
	conn, err := net.Dial("tcp", ":39999")
	if err != nil {
		return err
	}
	err = json.NewEncoder(conn).Encode(msg)
	if err != nil {
		return err
	}
	fmt.Println("Files loaded")
	return nil
}

type msg struct {
	MessageID int

	// for MessageID=0,1
	ScriptStates []ScriptState
	// for MessageID=2
	Message string

	// for MessageID=3
	Error              string
	ErrorMessagePrefix string
	GUID               string

	// for MessageID=6
	SavePath string
}
type ScriptState struct {
	Name   string
	GUID   string
	Script string
	UI     string
}

func syncFiles(msg msg) error {
	// Save all files
	for _, s := range msg.ScriptStates {
		name := fmt.Sprintf("%s.%s", s.Name, s.GUID)
		name = filepath.Join(*flagPath, name)
		if s.Script != "" {
			err := os.WriteFile(name+".lua", []byte(s.Script), os.ModePerm)
			if err != nil {
				return err
			}
		}
		if s.UI != "" {
			err := os.WriteFile(name+".xml", []byte(s.UI), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	// TODO handle removed files
	return nil
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Read the incoming connection into the buffer.
	var (
		buf bytes.Buffer
		msg msg
	)
	if _, err := io.Copy(&buf, conn); err != nil {
		panic(err)
	}
	if err := os.WriteFile("/tmp/output", buf.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}
	err := json.Unmarshal(buf.Bytes(), &msg)
	if err != nil {
		panic(err)
	}
	// Send a response back to person contacting us.
	fmt.Println("received msg ID", msg.MessageID)
	// Close the connection when you're done with it.
	conn.Close()
	switch msg.MessageID {
	case 0: // Pushing new object
		fmt.Println("Push object", buf.String())
		err := syncFiles(msg)
		if err != nil {
			panic(err)
		}

	case 1: // Loading new game
		err := syncFiles(msg)
		if err != nil {
			panic(err)
		}

	case 2: // Print/Debug message
		fmt.Println("TTS:", msg.Message)

	case 3: // Error message
		fmt.Fprintf(os.Stderr, "TTS error: %s %s GUID=%s", msg.ErrorMessagePrefix, msg.Error, msg.GUID)

	case 6: // Game saved
		fmt.Println("Game saved")
		bz, err := os.ReadFile(msg.SavePath)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(*flagPath, "savegame.json"), bz, os.ModePerm); err != nil {
			panic(err)
		}

	case 7: // Object created
		fmt.Println("Object created", msg.GUID)

	default:
		fmt.Println("unhandled messageID", msg.MessageID)
	}
}
