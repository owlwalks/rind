# Rind

[![GoDoc](https://godoc.org/github.com/owlwalks/rind?status.svg)](https://godoc.org/github.com/owlwalks/rind)
[![Build Status](https://travis-ci.com/owlwalks/rind.svg?branch=master)](https://travis-ci.com/owlwalks/rind)

Rind is a DNS server with REST interface for records management, best use is for your local service discovery, DNS forwarding and caching.

## Examples
See complete example [here](https://github.com/owlwalks/rind/blob/master/rind/main.go))

Start DNS server:
```golang
import github.com/owlwalks/rind

rind.Start("rw-dirpath", []net.UDPAddr{{IP: net.IP{1, 1, 1, 1}, Port: 53}})
```

## Manage records
```shell
// Add a SRV record
curl -X POST \
  http://localhost/dns \
  -H 'Content-Type: application/json' \
  -d '{
	"Host": "_sip._tcp.example.com.",
	"TTL": 300,
	"Type": "SRV",
	"SRV": {
		"Priority": 0,
		"Weight": 5,
		"Port": 5060,
		"Target": "sipserver.example.com."
	}
}'

// Update a A record from 124.108.115.87 to 127.0.0.1
curl -X PUT \
  http://localhost/dns \
  -H 'Content-Type: application/json' \
  -d '{
	"Host": "example.com.",
	"TTL": 600,
	"Type": "A",
	"OldData": "124.108.115.87",
	"Data": "127.0.0.1"
}'

// Delete a record
curl -X DELETE \
  http://localhost/dns \
  -H 'Content-Type: application/json' \
  -d '{
	"Host": "example.com.",
	"Type": "A"
}'
```

## Features:
- [x] DNS server
  - [x] DNS forwarding
  - [x] DNS caching
  - [x] A record
  - [x] NS record
  - [x] CNAME record
  - [x] SOA record
  - [x] PTR record
  - [x] MX record
  - [x] AAAA record
  - [x] SRV record
- [x] REST server
  - [x] Create records
  - [x] Read records
  - [x] Update records
  - [x] Delete records

Todo:
- [ ] Primary, secondary model
