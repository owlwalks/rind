package rind

import (
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
