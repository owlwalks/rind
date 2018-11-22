package rind

import (
	"net"
	"sync"
)

type addrBag struct {
	sync.RWMutex
	data map[string][]*net.UDPAddr
}

func (b *addrBag) get(key string) ([]*net.UDPAddr, bool) {
	b.RLock()
	val, ok := b.data[key]
	b.RUnlock()
	return val, ok
}

func (b *addrBag) set(key string, addr *net.UDPAddr) {
	b.Lock()
	if _, ok := b.data[key]; ok {
		b.data[key] = append(b.data[key], addr)
	} else {
		b.data[key] = []*net.UDPAddr{addr}
	}
	b.Unlock()
}

func (b *addrBag) remove(key string) bool {
	b.Lock()
	_, ok := b.data[key]
	delete(b.data, key)
	b.Unlock()
	return ok
}
