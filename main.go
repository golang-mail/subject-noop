// subject-noop is the minimal go-mail socket-mode plugin example.
// It listens on the Unix socket path go-mail provides via GO_MAIL_PLUGIN_SOCKET,
// reads each email's subject (logging it via stderr) and returns no actions.
package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
)

type envelope struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Method  string          `json:"method"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type evalReq struct {
	Email struct {
		UID      uint32 `json:"uid"`
		Account  string `json:"account"`
		Mailbox  string `json:"mailbox"`
		Envelope struct {
			Subject string `json:"Subject"`
		} `json:"envelope"`
	} `json:"email"`
}

func main() {
	log.SetFlags(0)
	path := os.Getenv("GO_MAIL_PLUGIN_SOCKET")
	if path == "" {
		log.Fatal("GO_MAIL_PLUGIN_SOCKET not set")
	}
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		log.Fatalf("listen %s: %v", path, err)
	}
	_ = os.Chmod(path, 0o600)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			return
		}
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	defer conn.Close()
	in := bufio.NewScanner(conn)
	in.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	out := json.NewEncoder(conn)

	for in.Scan() {
		var msg envelope
		if err := json.Unmarshal(in.Bytes(), &msg); err != nil {
			log.Printf("bad envelope: %v", err)
			continue
		}

		switch msg.Method {
		case "init":
			_ = out.Encode(envelope{
				ID:      msg.ID,
				Type:    "response",
				Method:  "init",
				Payload: json.RawMessage(`{"name":"subject-noop","version":"0.2.0","actions":[],"wants":["envelope"]}`),
			})
		case "evaluate":
			var r evalReq
			_ = json.Unmarshal(msg.Payload, &r)
			log.Printf("subject: %q", r.Email.Envelope.Subject)
			_ = out.Encode(envelope{
				ID:      msg.ID,
				Type:    "response",
				Method:  "evaluate",
				Payload: json.RawMessage(`{}`),
			})
		case "shutdown":
			return
		}
	}
}
