package main

import (
	"log"
	"net"
	"net/http"
	"rind"
)

func main() {
	dns := rind.Start("", []net.UDPAddr{{IP: net.IP{1, 1, 1, 1}, Port: 53}})
	rest := rind.RestService{Dn: dns}

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

	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		// authentication intercepting
		var _ = "intercept"
		return func(w http.ResponseWriter, r *http.Request) {
			h(w, r)
		}
	}

	http.Handle("/dns", withAuth(dnsHandler()))
	log.Fatal(http.ListenAndServe(":80", nil))
}
