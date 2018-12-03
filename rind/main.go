package main

import (
	"log"
	"net"
	"net/http"

	"github.com/owlwalks/rind"
	"flag"
	"os"
	"github.com/golang/glog"
)

var rwDirPath   = flag.String("rwdir","/var/dns","dns storage dir")
var listenIP    = flag.String("listenip", "8.8.8.8", "dns forward ip")
var listenPort  = flag.Int("listenport", 53, "dns forward port")

func main() {
	flag.Parse()
	if err := ensureDir(*rwDirPath); err != nil {
		glog.Errorf("create rwdirpath: %v error: %v", *rwDirPath, err)
		return
	}
	glog.Info("starting rind")
	dns := rind.Start(*rwDirPath, []net.UDPAddr{{IP: net.ParseIP(*listenIP), Port: *listenPort}})
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

func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.Mkdir(path, 0666)
	}
	return nil
}