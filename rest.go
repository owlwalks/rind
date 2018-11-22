package rind

import (
	"encoding/json"
	"net/http"
)

// RestServer will do CRUD on DNS records
type RestServer interface {
	Create() http.HandlerFunc
	Read() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
}

// RestService is an implementation of RestServer interface.
type RestService struct {
	dns DNSService
}

type post struct {
	Host string
	TTL  uint32
	Type string
	Data json.RawMessage
}

type postSOA struct {
	NS      json.RawMessage
	MBox    json.RawMessage
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	MinTTL  uint32
}

type postMX struct {
	Pref uint16
	MX   json.RawMessage
}

type postSRV struct {
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   json.RawMessage
}

type get struct {
	Host string
	TTL  uint32
	Type string
	Data string
}

type update struct {
	Host    string
	TTL     uint32
	Type    string
	OldData json.RawMessage
	Data    json.RawMessage
}

type del struct {
	Host string
	Type string
	Data json.RawMessage
}

// Create is HTTP handler of POST request.
// Use for adding new record to DNS server.
func (s *RestService) Create() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req post
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resource, err := toResource(req.Host, req.TTL, req.Type, req.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.dns.save(ntString(req.Host, req.Type), resource)
		w.WriteHeader(http.StatusCreated)
	})
}

// Read is HTTP handler of GET request.
// Use for reading existed records from DNS server.
func (s *RestService) Read() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(s.dns.all())
	})
}

// Delete is HTTP handler of DELETE request.
// Use for removing records from DNS server.
func (s *RestService) Delete() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req del
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := ntString(req.Host, req.Type)
		ok := false
		if req.Data != nil {
			resource, err := toResource(req.Host, 0, req.Type, req.Data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ok = s.dns.remove(id, &resource)
		} else {
			ok = s.dns.remove(id, nil)
		}

		if ok {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "", http.StatusNotFound)
		return
	})
}
