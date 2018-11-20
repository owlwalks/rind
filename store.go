package rind

import (
	"encoding/gob"
	"log"
	"os"
	"sync"

	"golang.org/x/net/dns/dnsmessage"
)

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

func (b *kv) set(key string, resources []dnsmessage.Resource) {
	b.Lock()
	if _, ok := b.data[key]; ok {
		b.data[key] = append(b.data[key], resources...)
	} else {
		b.data[key] = resources
	}
	b.Unlock()
}

func (b *kv) save() {
	fWriter, err := os.Create(b.filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fWriter.Close()

	enc := gob.NewEncoder(fWriter)
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
	if err = dec.Decode(&b.data); err != nil {
		log.Fatal(err)
	}
}
