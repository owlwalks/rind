package rind

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
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

type row struct {
	k string
	v []dnsmessage.Resource
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
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for k, v := range b.data {
		err := enc.Encode(row{k, v})
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Fprintln(fWriter, buf)
	}
}

func (b *kv) load() {
	fReader, err := os.Open(b.filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fReader.Close()
	scanner := bufio.NewScanner(fReader)
	var buf bytes.Buffer
	dec := gob.NewDecoder(&buf)
	for scanner.Scan() {
		_, err = buf.Write(scanner.Bytes())
		if err != nil {
			log.Println(err)
			continue
		}
		var r row
		err = dec.Decode(&r)
		if err != nil {
			log.Println(err)
			continue
		}
		b.set(r.k, r.v)
	}
}
