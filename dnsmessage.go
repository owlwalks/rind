// Copyright 2018 The Rind Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rind

import (
	"fmt"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

// question to string
func qString(q dnsmessage.Question) string {
	b := make([]byte, q.Name.Length+2)
	for i := 0; i < int(q.Name.Length); i++ {
		b[i] = q.Name.Data[i]
	}
	b[q.Name.Length] = uint8(q.Type >> 8)
	b[q.Name.Length+1] = uint8(q.Type & 0xff)

	return string(b)
}

// resource name and type to string
func ntString(rName dnsmessage.Name, rType dnsmessage.Type) string {
	b := make([]byte, rName.Length+2)
	for i := 0; i < int(rName.Length); i++ {
		b[i] = rName.Data[i]
	}
	b[rName.Length] = uint8(rType >> 8)
	b[rName.Length+1] = uint8(rType & 0xff)

	return string(b)
}

// resource to string
func rString(r dnsmessage.Resource) string {
	var sb strings.Builder
	sb.Write(r.Header.Name.Data[:])
	sb.WriteString(r.Header.Type.String())
	sb.WriteString(r.Body.GoString())

	return sb.String()
}

// packet to string
func pString(p Packet) string {
	return fmt.Sprint(p.message.ID)
}
