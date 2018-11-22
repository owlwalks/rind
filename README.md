# Rind (REST interfacing DNS server)

[![GoDoc](https://godoc.org/github.com/owlwalks/rind?status.svg)](https://godoc.org/github.com/owlwalks/rind)
[![Build Status](https://travis-ci.com/owlwalks/rind.svg?branch=master)](https://travis-ci.com/owlwalks/rind)

Rind is a DNS server with REST interface for records management, best use is for your local service discovery, DNS forwarding and caching (WIP)

Features:
- [x] DNS server
  - [x] DNS forwarding
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
- [ ] DNS caching
- [ ] TXT record
- [ ] OPT record
- [ ] Primary, secondary model
