package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
)

type server struct {
	hostname    string
	connections []*client
	transmit    chan *client
	register    chan *client
	unregister  chan *client
	msgLog      io.Writer
	errLog      *log.Logger
}

func (s *server) listen() {
	ln, err := net.Listen("tcp", ":5222")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go s.run()

	for {
		c, err := ln.Accept()
		if err != nil {
			s.errLog.Printf("%s", err.Error())
		}
		go s.serve(c)
	}

}

func (s *server) run() {
	for {
		select {
		case message := <-s.transmit:
			s.send(message)
		case conn := <-s.register:
			s.connections = append(s.connections, conn)
		case conn := <-s.unregister:
			s.removeConn(conn)
		}
	}
}

func (s *server) streamError(c *client) {
	fmt.Fprintf(c.conn, "</stream:stream>\n")
	c.conn.Close()
	s.errLog.Printf("stream error")
}

func (s *server) serve(conn net.Conn) {
	c := &client{
		conn: conn,
		p:    xml.NewDecoder(conn),
		e:    xml.NewEncoder(tee{conn, s.msgLog}),
	}

	defer func() {
		if c.jid != "" {
			s.unregister <- c
		}
	}()

	for {
		se, _ := nextStart(c.p)
		switch se.Name.Local {
		case "stream":
			s.sendFeatures(c)
			break
		case "auth":
			if !s.auth(c, se) {
				s.streamError(c)
				return
			}
			break
		case "iq":
			if !s.bind(c, se) {
				s.streamError(c)
				return
			}
			s.register <- c
			break
		case "presence":
			if c.jid != "" {
				c.msg = &presence{}
				s.transmit <- c
			}
			break
		case "message":
			if c.jid != "" {
				c.msg = &message{}
				if err := c.p.DecodeElement(c.msg, &se); err != nil {
					s.errLog.Printf("%s", err.Error())
					return
				}
				s.transmit <- c
			}
			break
			// TODO: add default
		}
	}
}

func (s *server) send(c *client) {
	switch t := c.msg.(type) {
	case *message:
		for i := range s.connections {
			if t.To == s.connections[i].jid {
				s.connections[i].e.Encode(c.msg)
			}
		}
		return
	case *presence:
		for i := range s.connections {
			t.From = c.jid
			t.To = s.connections[i].jid
			s.connections[i].e.Encode(c.msg)
		}
		for i := range s.connections {
			if s.connections[i].jid != c.jid {
				t.From = s.connections[i].jid
				t.To = c.jid
				c.e.Encode(c.msg)
			}
		}
		return
	}
}

func (s *server) removeConn(c *client) {
	var i int
	for i = range s.connections {
		if s.connections[i].conn == c.conn {
			break
		}
	}
	s.connections = append(s.connections[:i], s.connections[i+1:]...)
}

func nextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err != nil || t == nil {
			return xml.StartElement{}, err
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}
