package rind

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func Test_kv_save_load(t *testing.T) {
	dirPath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Skip(err)
	}
	name, _ := dnsmessage.NewName("test")
	data := map[string][]dnsmessage.Resource{
		"test": []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{
					Name:  name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
				Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
			},
		},
	}
	book := kv{data: data, rwDirPath: dirPath}
	book.save()
	store, _ := os.Stat(filepath.Join(dirPath, storeName))
	bk, _ := os.Stat(filepath.Join(dirPath, storeBkName))
	if store.Size() != 591 || bk.Size() != 0 {
		t.Fail()
	}

	bookNil := kv{rwDirPath: dirPath}
	bookNil.load()
	body, _ := bookNil.data["test"][0].Body.(*dnsmessage.AResource)
	if body.A != [4]byte{127, 0, 0, 1} {
		t.Fail()
	}
}
