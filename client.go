package main

import (
	"encoding/xml"
	"fmt"
	"net"
)

type client struct {
	conn          net.Conn
	p             *xml.Decoder
	e             *xml.Encoder
	jid           string
	msg           interface{}
	authenticated bool
}

// TODO: move methods into server.go
// * change fmt.Fprintf to c.e.Encode() if possible.
// * proper error handling.
func (s *server) sendFeatures(c *client) {
	if !c.authenticated {
		s.restart(c)
		fmt.Fprintf(
			c.conn,
			"<stream:features><mechanisms xmlns='%s'><mechanism>ANONYMOUS</mechanism></mechanisms></stream:features>\n",
			nsSASL,
		)
        return
	}
	s.restart(c)
	fmt.Fprintf(
		c.conn,
		"<stream:features><bind xmlns='%s'/></stream:features>\n",
		nsBind,
	)
}

func (s *server) auth(c *client, se xml.StartElement) bool {
	for _, a := range se.Attr {
		switch a.Value {
		case "ANONYMOUS":
			fmt.Fprintf(c.conn, "<success xmlns='%s'/>\n", nsSASL)
			c.authenticated = true
			return true
		}
	}
	fmt.Fprintf(
		c.conn,
		"<failure xmlns='%s'><malformed-request/></failure>\n",
		nsSASL,
	)
	return false
}

func (s *server) restart(c *client) error {
	// TODO: try xml.Header
	_, err := fmt.Fprintf(
		c.conn,
		"<?xml version='1.0'?><stream:stream id='%x' version='1.0' xml:lang='en' xmlns:stream='%s' from='%s' xmlns='%s'>\n",
		rng(),
		nsStreams,
		s.hostname,
		nsClient,
	)
	return err
}

func (s *server) bind(c *client, se xml.StartElement) bool {
	var i iq
	if err := c.p.DecodeElement(&i, &se); err != nil {
		s.errLog.Printf("%s", err.Error())
		return false
	}
	if &i.Bind == nil {
		fmt.Fprintf(
			c.conn,
			"<stream:error><not-well-formed xmlns='%s'/></stream:error>\n",
			nsStreams,
		)
		s.errLog.Printf("<iq> result missing <bind>")
		return false
	}
	c.jid = fmt.Sprintf("%x@%s/%x", rng(), s.hostname, rng())
	fmt.Fprintf(
		c.conn,
		"<iq type='result' id='%x'><bind xmlns='%s'><jid>%s</jid></bind></iq>\n",
		&i.ID,
		nsBind,
		c.jid,
	)
	return true
}
