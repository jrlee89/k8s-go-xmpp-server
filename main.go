package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stderr, "XMPP Error: ", log.LstdFlags)
	s := &server{
		hostname:   "localhost",
		transmit:   make(chan *client),
		register:   make(chan *client),
		unregister: make(chan *client),
		msgLog:     os.Stdout,
		errLog:     logger,
	}
	s.listen()
}
