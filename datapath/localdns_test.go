package datapath

import (
	"encoding/binary"
	"net"
	"testing"
)

func queryDNS(name string, kind uint16) []byte {
	q := []byte{0xab, 0xcd, 0x01, 0, 0, 1, 0, 0, 0, 0, 0, 0}
	for _, part := range splitName(name) {
		q = append(q, byte(len(part)))
		q = append(q, []byte(part)...)
	}
	q = append(q, 0, 0, 0, 0, 1)
	binary.BigEndian.PutUint16(q[len(q)-4:len(q)-2], kind)
	return q
}
func splitName(s string) []string {
	var parts []string
	for _, p := range []byte(s) {
		_ = p
	}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '.' {
			parts = append(parts, s[start:i])
			start=i+1
		}
	}
	return parts
}
func TestLocalDNSAnswersAndFallback(t *testing.T) {
	var dns localDNSStore
	if err := dns.setConfig(`{"enabled":true,"records":[{"name":"PhotoCraft","ip":"10.201.244.112","enabled":true},{"name":"disabled","ip":"10.0.0.1","enabled":false}]}`);err!=nil { t.Fatal(err) }
	packet, handled := dns.answer(queryDNS("photocraft", 1))
	if !handled || binary.BigEndian.Uint16(packet[6:8])!=1 || !net.IP(packet[len(packet)-4:]).Equal(net.ParseIP("10.201.244.112")) { t.Fatalf("bad A response: %v",packet) }
	packet, handled = dns.answer(queryDNS("photocraft",28))
	if !handled || binary.BigEndian.Uint16(packet[6:8]) != 0 { t.Fatal("expected AAAA NODATA") }
	for _, name := range []string{"other", "disabled"} {
		if _, handled := dns.answer(queryDNS(name,1)); handled { t.Fatalf("unexpected local answer for %s",name) }
	}
}
func TestLocalDNSConfigValidationAndDisable(t *testing.T) {
	var dns localDNSStore
	if err := dns.setConfig(`{"enabled":true,"records":[{"name":"pdf","ip":"10.201.244.112","enabled":true}]}`);err!=nil { t.Fatal(err) }
	if err := dns.setConfig(`{"enabled":true,"records":[{"name":"bad..host","ip":"10.0.0.1","enabled":true}]}`);err==nil { t.Fatal("accepted invalid hostname") }
	if _, handled := dns.answer(queryDNS("pdf",1)); !handled { t.Fatal("invalid update destroyed existing entries") }
	if err := dns.setConfig(`{"enabled":false,"records":[{"name":"pdf","ip":"10.201.244.112","enabled":true}]}`);err!=nil { t.Fatal(err) }
	if _, handled := dns.answer(queryDNS("pdf",1)); handled { t.Fatal("disabled DNS answered") }
}
