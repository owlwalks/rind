# Rind

[![GoDoc](https://godoc.org/github.com/owlwalks/rind?status.svg)](https://godoc.org/github.com/owlwalks/rind)
[![Build Status](https://travis-ci.com/owlwalks/rind.svg?branch=master)](https://travis-ci.com/owlwalks/rind)

Rind is a DNS server with REST interface for records management, best use is for your local service discovery, DNS forwarding and caching.

Example (complete example [here](https://github.com/owlwalks/rind/blob/master/rind/main.go)):

Start DNS server:
```golang
import github.com/owlwalks/rind

rind.Start("rw-dirpath", []net.UDPAddr{{IP: net.IP{1, 1, 1, 1}, Port: 53}})
```

Start Http Server
```golang
import github.com/owlwalks/rind

rest := rind.RestService{}

withAuth := func(h http.HandlerFunc) http.HandlerFunc {
  // authentication intercepting
  var _ = "intercept"
  return func(w http.ResponseWriter, r *http.Request) {
    h(w, r)
  }
}

dnsHandler := func() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
      rest.Create(w, r)
    case http.MethodGet:
      rest.Read(w, r)
    case http.MethodPut:
      rest.Update(w, r)
    case http.MethodDelete:
      rest.Delete(w, r)
    }
  }
}

http.Handle("/dns", withAuth(dnsHandler()))
```

Features:
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
