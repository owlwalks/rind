package rind

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

// DNSServer will do Listen, Query and Send.
type DNSServer interface {
	Listen()
	Query(Packet)
	Send()
}

// DNSService is an implementation of DNSServer interface.
type DNSService struct {
	conn       *net.UDPConn
	book       store
	memo       addrBag
	forwarders []net.UDPAddr
}

// Packet carries DNS packet payload and sender address.
type Packet struct {
	addr    *net.UDPAddr
	message dnsmessage.Message
}

const (
	// DNS server default port
	udpPort int = 53
	// DNS packet max length
	packetLen int = 512
)

var (
	errTypeNotSupport = errors.New("type not support")
	errIPInvalid      = errors.New("invalid IP address")
)

// Listen starts a DNS server on port 53
func (s *DNSService) Listen() {
	var err error
	s.conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	if err != nil {
		log.Fatal(err)
	}
	defer s.conn.Close()

	for {
		buf := make([]byte, packetLen)
		_, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		var m dnsmessage.Message
		err = m.Unpack(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(m.Questions) == 0 {
			continue
		}
		go s.Query(Packet{addr, m})
	}
}

// Query lookup answers for DNS message.
func (s *DNSService) Query(p Packet) {
	// got response from forwarder, send it back to client
	if p.message.Header.Response {
		key := pString(p)
		if addrs, ok := s.memo.get(key); ok {
			go s.saveBulk(qString(p.message.Questions[0]), p.message.Answers)
			for _, addr := range addrs {
				go sendPacket(s.conn, p.message, &addr)
			}
			s.memo.remove(key)
		}
		return
	}

	// was checked before entering this routine
	q := p.message.Questions[0]

	// answer the question
	val, ok := s.book.get(qString(q))

	if ok {
		p.message.Answers = append(p.message.Answers, val...)
		go sendPacket(s.conn, p.message, p.addr)
	} else {
		// forwarding
		for i := 0; i < len(s.forwarders); i++ {
			s.memo.set(pString(p), *p.addr)
			go sendPacket(s.conn, p.message, &(s.forwarders[i]))
		}
	}
}

func sendPacket(conn *net.UDPConn, message dnsmessage.Message, addr *net.UDPAddr) {
	packed, err := message.Pack()
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conn.WriteToUDP(packed, addr)
	if err != nil {
		log.Println(err)
	}
}

// New setups a DNSService, rwDirPath is read-writable directory path for storing dns records.
func New(rwDirPath string, forwarders []net.UDPAddr) DNSService {
	return DNSService{
		book:       store{data: make(map[string][]dnsmessage.Resource), rwDirPath: rwDirPath},
		memo:       addrBag{data: make(map[string][]net.UDPAddr)},
		forwarders: forwarders,
	}
}

// Start conveniently init every parts of DNS service.
func Start(rwDirPath string, forwarders []net.UDPAddr) {
	s := New(rwDirPath, forwarders)
	go s.Listen()
}

func (s *DNSService) save(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	ok := s.book.set(key, resource, old)
	go s.book.save()

	return ok
}

func (s *DNSService) saveBulk(key string, resources []dnsmessage.Resource) {
	s.book.override(key, resources)
	go s.book.save()
}

func (s *DNSService) all() []get {
	book := s.book.clone()
	var recs []get
	for _, r := range book {
		for _, v := range r {
			body := v.Body.GoString()
			i := strings.Index(body, "{")
			recs = append(recs, get{
				Host: v.Header.Name.String(),
				TTL:  v.Header.TTL,
				Type: v.Header.Type.String()[4:],
				Data: body[i : len(body)-1], // get content within "{" and "}"
			})
		}
	}
	return recs
}

func (s *DNSService) remove(key string, r *dnsmessage.Resource) bool {
	ok := s.book.remove(key, r)
	if ok {
		go s.book.save()
	}
	return ok
}

func toResource(name string, ttl uint32, sType string, data []byte) (dnsmessage.Resource, error) {
	rName, err := dnsmessage.NewName(name)
	rNil := dnsmessage.Resource{}
	if err != nil {
		return rNil, err
	}

	var rType dnsmessage.Type
	var rBody dnsmessage.ResourceBody

	switch sType {
	case "A":
		rType = dnsmessage.TypeA
		ip := net.ParseIP(string(data))
		if ip == nil {
			return rNil, errIPInvalid
		}
		rBody = &dnsmessage.AResource{A: [4]byte{ip[12], ip[13], ip[14], ip[15]}}
	case "NS":
		rType = dnsmessage.TypeNS
		ns, err := dnsmessage.NewName(string(data))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.NSResource{NS: ns}
	case "CNAME":
		rType = dnsmessage.TypeCNAME
		cname, err := dnsmessage.NewName(string(data))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.CNAMEResource{CNAME: cname}
	case "SOA":
		rType = dnsmessage.TypeSOA
		var soa postSOA
		if err = json.Unmarshal(data, &soa); err != nil {
			return rNil, err
		}
		soaNS, err := dnsmessage.NewName(string(soa.NS))
		if err != nil {
			return rNil, err
		}
		soaMBox, err := dnsmessage.NewName(string(soa.MBox))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.SOAResource{NS: soaNS, MBox: soaMBox, Serial: soa.Serial, Refresh: soa.Refresh, Retry: soa.Retry, Expire: soa.Expire}
	case "PTR":
		rType = dnsmessage.TypePTR
		ptr, err := dnsmessage.NewName(string(data))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.PTRResource{PTR: ptr}
	case "MX":
		rType = dnsmessage.TypeMX
		var mx postMX
		if err = json.Unmarshal(data, &mx); err != nil {
			return rNil, err
		}
		mxName, err := dnsmessage.NewName(string(mx.MX))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.MXResource{Pref: mx.Pref, MX: mxName}
	case "AAAA":
		rType = dnsmessage.TypeAAAA
		ip := net.ParseIP(string(data))
		if ip == nil {
			return rNil, errIPInvalid
		}
		var ipV6 [16]byte
		copy(ipV6[:], ip)
		rBody = &dnsmessage.AAAAResource{AAAA: ipV6}
	case "SRV":
		rType = dnsmessage.TypeSRV
		var srv postSRV
		if err = json.Unmarshal(data, &srv); err != nil {
			return rNil, err
		}
		srvTarget, err := dnsmessage.NewName(string(srv.Target))
		if err != nil {
			return rNil, err
		}
		rBody = &dnsmessage.SRVResource{Priority: srv.Priority, Weight: srv.Weight, Port: srv.Port, Target: srvTarget}
	case "TXT":
		fallthrough
	case "OPT":
		fallthrough
	default:
		return rNil, errTypeNotSupport
	}

	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  rName,
			Type:  rType,
			Class: dnsmessage.ClassINET,
			TTL:   ttl,
		},
		Body: rBody,
	}, nil
}
