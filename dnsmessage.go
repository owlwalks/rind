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
	var sb strings.Builder
	sb.Write(q.Name.Data[:])
	sb.WriteString(q.Type.String())
	return sb.String()
}

// resource name and type to string
func ntString(name string, rType string) string {
	return name + rType
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
