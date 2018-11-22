package rind

import (
	"encoding/json"
	"io/ioutil"
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
	Host string          `json:"host"`
	TTL  int             `json:"ttl"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
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

// Create is HTTP handler of POST request.
// Use for adding new record to DNS server.
func (s *RestService) Create() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var req post
		if err = json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resources, err := toResources(req.Host, req.Type, req.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.dns.save("later", resources)
		w.WriteHeader(http.StatusCreated)
	})
}
