package rind

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	data      map[string][]dnsmessage.Resource
	rwDirPath string
}

func (s *store) get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	val, ok := s.data[key]
	s.RUnlock()
	return val, ok
}

func (s *store) set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	changed := false
	s.Lock()
	if _, ok := s.data[key]; ok {
		if old != nil {
			for i, rec := range s.data[key] {
				if rString(rec) == rString(*old) {
					s.data[key][i] = resource
					changed = true
					break
				}
			}
		} else {
			s.data[key] = append(s.data[key], resource)
			changed = true
		}
	} else {
		s.data[key] = []dnsmessage.Resource{resource}
		changed = true
	}
	s.Unlock()

	return changed
}

func (s *store) remove(key string, r *dnsmessage.Resource) bool {
	ok := false
	s.Lock()
	if r == nil {
		_, ok = s.data[key]
		delete(s.data, key)
	} else {
		if _, ok = s.data[key]; ok {
			for i, rec := range s.data[key] {
				if rString(rec) == rString(*r) {
					s.data[key] = append(s.data[key][:i], s.data[key][i+1:]...)
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

func (s *store) clone() map[string][]dnsmessage.Resource {
	cp := make(map[string][]dnsmessage.Resource)
	s.RLock()
	for k, v := range s.data {
		cp[k] = v
	}
	s.RUnlock()
	return cp
}
