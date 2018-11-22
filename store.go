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

func (b *kv) set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	changed := false
	b.Lock()
	if _, ok := b.data[key]; ok {
		if old != nil {
			for i, rec := range b.data[key] {
				if rString(rec) == rString(*old) {
					b.data[key][i] = resource
					changed = true
					break
				}
			}
		} else {
			b.data[key] = append(b.data[key], resource)
			changed = true
		}
	} else {
		b.data[key] = []dnsmessage.Resource{resource}
		changed = true
	}
	b.Unlock()

	return changed
}

func (b *kv) remove(key string, r *dnsmessage.Resource) bool {
	ok := false
	b.Lock()
	if r == nil {
		_, ok = b.data[key]
		delete(b.data, key)
	} else {
		if _, ok = b.data[key]; ok {
			for i, rec := range b.data[key] {
				if rString(rec) == rString(*r) {
					b.data[key] = append(b.data[key][:i], b.data[key][i+1:]...)
					ok = true
					break
				}
			}
		}
	}
	b.Unlock()
	return ok
}

func (b *kv) save() {
	bk, err := os.OpenFile(filepath.Join(b.filePath, storeBkName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer bk.Close()

	dst, err := os.OpenFile(filepath.Join(b.filePath, storeName), os.O_RDWR|os.O_CREATE, 0666)
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
	book := b.clone()
	if err = enc.Encode(book); err != nil {
		// main store file is corrupted
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
