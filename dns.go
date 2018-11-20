package rind

import (
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer will do Listen, Query and Send.
type DNSServer interface {
	Listen()
	Query(Packet)
	Send()
}

type kv struct {
	sync.RWMutex
	data map[string][]dnsmessage.Resource
}

// DNSService is the implementation of DNSServer interface.
type DNSService struct {
	packets chan Packet
	conn    *net.UDPConn
	book    kv
}

// Packet carries DNS packet payload and sender address.
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
	s.conn = c
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

// Query lookup answers for DNS message.
func (s *DNSService) Query(p Packet) {
	// unpack
	var m dnsmessage.Message
	err := m.Unpack(p.message)
	if err != nil {
		log.Println(err)
		return
	}

	// parse raw message
	var parser dnsmessage.Parser
	if _, err := parser.Start(p.message); err != nil {
		log.Println(err)
		return
	}

	// pick question
	q, err := parser.Question()
	if err != nil && err != dnsmessage.ErrSectionDone {
		log.Println(err)
		return
	}

	var sb strings.Builder
	sb.WriteString(q.Name.String())
	sb.WriteString(q.Class.String())
	sb.WriteString(q.Type.String())

	// answer the question
	s.book.RLock()
	if val, ok := s.book.data[sb.String()]; ok {
		m.Answers = append(m.Answers, val...)
	}
	s.book.RUnlock()

	p.message, err = m.Pack()
	if err != nil {
		log.Println(err)
		return
	}

	s.packets <- p
}

// Send sends DNS message back with answer.
func (s *DNSService) Send() {
	for p := range s.packets {
		go sendPacket(s.conn, p)
	}
}

func sendPacket(conn *net.UDPConn, p Packet) {
	_, err := conn.WriteToUDP(p.message, p.addr)
	if err != nil {
		log.Println(err)
	}
}

// New setups a DNSService with custom number of buffered packets.
// Set nPackets high enough for better throughput.
func New(nPackets int) DNSService {
	packets := make(chan Packet, nPackets)
	return DNSService{packets, nil, kv{data: make(map[string][]dnsmessage.Resource)}}
}

// Start convenient init every parts of DNS service.
// See New for nPackets.
func Start(nPackets int) {
	s := New(nPackets)
	go s.Listen()
	go s.Send()
}
