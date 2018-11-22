package rind

import (
	"encoding/gob"
	"log"
	"os"
	"sync"

	"golang.org/x/net/dns/dnsmessage"
)

func init() {
	gob.Register(&dnsmessage.AResource{})
}

type kv struct {
	sync.RWMutex
	data     map[string][]dnsmessage.Resource
	filePath string
}

func (b *kv) get(key string) ([]dnsmessage.Resource, bool) {
	b.RLock()
	val, ok := b.data[key]
	b.RUnlock()
	return val, ok
}

func (b *kv) set(key string, resource dnsmessage.Resource) {
	b.Lock()
	if _, ok := b.data[key]; ok {
		b.data[key] = append(b.data[key], resource)
	} else {
		b.data[key] = []dnsmessage.Resource{resource}
	}
	b.Unlock()
}

func (b *kv) remove(key string, r *dnsmessage.Resource) bool {
	ok := false
	b.Lock()
	if r == nil {
		_, ok = b.data[key]
		delete(b.data, key)
	} else {
		for i, rec := range b.data[key] {
			if rString(rec) == rString(*r) {
				b.data[key] = append(b.data[key][:i], b.data[key][i+1:]...)
				ok = true
				break
			}
		}
	}
	b.Unlock()
	return ok
}

func (b *kv) save() {
	fWriter, err := os.Create(b.filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fWriter.Close()

	enc := gob.NewEncoder(fWriter)

	b.RLock()
	defer b.RUnlock()

	if err = enc.Encode(b.data); err != nil {
		log.Fatal(err)
	}
}

func (b *kv) load() {
	fReader, err := os.Open(b.filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fReader.Close()

	dec := gob.NewDecoder(fReader)

	b.Lock()
	defer b.Unlock()

	if err = dec.Decode(&b.data); err != nil {
		log.Fatal(err)
	}
}

func (b *kv) clone() map[string][]dnsmessage.Resource {
	cp := make(map[string][]dnsmessage.Resource)
	b.RLock()
	for k, v := range b.data {
		cp[k] = v
	}
	b.RUnlock()
	return cp
}
