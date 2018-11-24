package rind

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func Test_store_save_load(t *testing.T) {
	dirPath, err := ioutil.TempDir("", "")
	if err != nil {
		t.Skip(err)
	}
	name, _ := dnsmessage.NewName("test")
	data := map[string]entry{
		"test": {
			Resources: []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}},
				},
			},
		},
	}
	book := store{data: data, rwDirPath: dirPath}
	book.save()
	main, _ := os.Stat(filepath.Join(dirPath, storeName))
	bk, _ := os.Stat(filepath.Join(dirPath, storeBkName))
	if main.Size() != 664 || bk.Size() != 0 {
		t.Fail()
	}
	bookNew := store{rwDirPath: dirPath}
	bookNew.load()
	r, _ := bookNew.get("test")
	body, _ := r[0].Body.(*dnsmessage.AResource)
	if body.A != [4]byte{127, 0, 0, 1} {
		t.Fail()
	}
}
