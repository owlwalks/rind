package rind

import (
	"log"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer will do Listen, Query and Send
type DNSServer interface {
	Listen()
	Query(Packet)
	Send()
}

// DNSService is the implementation of DNSServer interface
type DNSService struct {
	packets chan Packet
}

// Packet carries DNS packet payload and sender address
type Packet struct {
	addr    *net.UDPAddr
	message []byte
}

// Listen starts a DNS server on port 53
func (s *DNSService) Listen() {
	c, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	for {
		b := make([]byte, 512)
		_, addr, err := c.ReadFromUDP(b)
		if err != nil {
			continue
		}
		go s.Query(Packet{addr, b})
	}
}

// Query lookup answers for DNS message
func (s *DNSService) Query(p Packet) {
	// unpack
	var m dnsmessage.Message
	err := m.Unpack(p.message)
	if err != nil {
		return
	}

	// parse raw message
	var parser dnsmessage.Parser
	if _, err := parser.Start(p.message); err != nil {
		log.Println(err)
		return
	}

	// locate question
	q, err := parser.Question()
	if err != nil && err != dnsmessage.ErrSectionDone {
		log.Println(err)
		return
	}

	// append answer based on q.Name.String(), fake for now
	name, err := dnsmessage.NewName(q.Name.String())
	if err != nil {
		log.Println(err)
		return
	}

	m.Answers = append(m.Answers, dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  name,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
		},
		Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
	})

	p.message, err = m.Pack()
	if err != nil {
		log.Println(err)
		return
	}

	s.packets <- p
}

// Send sends DNS message back with answer
func (s *DNSService) Send() {
	for p := range s.packets {
		go sendPacket(p)
	}
}

func sendPacket(p Packet) {
	c, err := net.DialUDP("udp", nil, p.addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()
	_, err = c.WriteToUDP(p.message, p.addr)
	if err != nil {
		log.Println(err)
		return
	}
}
