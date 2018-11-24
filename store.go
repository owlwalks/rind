package rind

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	storeName   string = "store"
	storeBkName string = "store_bk"
)

func init() {
	gob.Register(&dnsmessage.AResource{})
}

type store struct {
	sync.RWMutex
	data      map[string]entry
	rwDirPath string
}

type entry struct {
	resources []dnsmessage.Resource
	ttl       uint32
	t         int64
}

func (s *store) get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	e, ok := s.data[key]
	s.RUnlock()
	now := time.Now().Unix()
	if e.ttl > 1 && (e.t+int64(e.ttl)) >= now {
		s.remove(key, nil)
		return nil, false
	}
	return e.resources, ok
}

func (s *store) set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	changed := false
	s.Lock()
	if _, ok := s.data[key]; ok {
		if old != nil {
			for i, rec := range s.data[key].resources {
				if rString(rec) == rString(*old) {
					s.data[key].resources[i] = resource
					changed = true
					break
				}
			}
		} else {
			e := s.data[key]
			e.resources = append(e.resources, resource)
			s.data[key] = e
			changed = true
		}
	} else {
		e := entry{
			resources: []dnsmessage.Resource{resource},
			ttl:       resource.Header.TTL,
			t:         time.Now().Unix(),
		}
		s.data[key] = e
		changed = true
	}
	s.Unlock()

	return changed
}

func (s *store) override(key string, resources []dnsmessage.Resource) {
	s.Lock()
	e := entry{
		resources: resources,
		ttl:       resources[0].Header.TTL,
		t:         time.Now().Unix(),
	}
	s.data[key] = e
	s.Unlock()
}

func (s *store) remove(key string, r *dnsmessage.Resource) bool {
	ok := false
	s.Lock()
	if r == nil {
		_, ok = s.data[key]
		delete(s.data, key)
	} else {
		if _, ok = s.data[key]; ok {
			for i, rec := range s.data[key].resources {
				if rString(rec) == rString(*r) {
					e := s.data[key]
					copy(e.resources[i:], e.resources[i+1:])
					var blank dnsmessage.Resource
					e.resources[len(e.resources)-1] = blank
					e.resources = e.resources[:len(e.resources)-1]
					s.data[key] = e
					ok = true
					break
				}
			}
		}
	}
	s.Unlock()
	return ok
}

func (s *store) save() {
	bk, err := os.OpenFile(filepath.Join(s.rwDirPath, storeBkName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer bk.Close()

	dst, err := os.OpenFile(filepath.Join(s.rwDirPath, storeName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer dst.Close()

	// backing up current store
	_, err = io.Copy(bk, dst)
	if err != nil {
		log.Println(err)
		return
	}

	enc := gob.NewEncoder(dst)
	book := s.clone()
	if err = enc.Encode(book); err != nil {
		// main store file is corrupted
		log.Fatal(err)
	}
}

func (s *store) load() {
	fReader, err := os.Open(filepath.Join(s.rwDirPath, storeName))
	if err != nil {
		log.Fatal(err)
	}
	defer fReader.Close()

	dec := gob.NewDecoder(fReader)

	s.Lock()
	defer s.Unlock()

	if err = dec.Decode(&s.data); err != nil {
		log.Fatal(err)
	}
}

func (s *store) clone() map[string]entry {
	cp := make(map[string]entry)
	s.RLock()
	for k, v := range s.data {
		cp[k] = v
	}
	s.RUnlock()
	return cp
}
