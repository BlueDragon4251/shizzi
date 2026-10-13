package datapath

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
)

// localDNSStore is immutable per snapshot, so a UI edit never races a query.
type localDNSStore struct { active atomic.Pointer[localDNSTable] }

type localDNSTable struct {
	enabled bool
	records map[string][4]byte
}

type localDNSConfig struct {
	Enabled bool `json:"enabled"`
	Records []struct {
		Name string `json:"name"`
		IP string `json:"ip"`
		Enabled bool `json:"enabled"`
	} `json:"records"`
}

func (s *localDNSStore) setConfig(raw string) error {
	var config localDNSConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil { return fmt.Errorf("invalid DNS configuration: %w", err) }
	if len(config.Records) > 256 { return fmt.Errorf("too many DNS entries (max 256)") }
	table := &localDNSTable{enabled: config.Enabled, records: make(map[string][4]byte)}
	for _, record := range config.Records {
		name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(record.Name), "."))
		if !validDNSName(name) { return fmt.Errorf("invalid local hostname: %q", record.Name) }
		ip := net.ParseIP(record.IP).To4()
		if ip == nil { return fmt.Errorf("invalid IPv4 for %s: %q", name, record.IP) }
		if _, exists := table.records[name]; exists { return fmt.Errorf("duplicate local hostname: %s", name) }
		if !record.Enabled { continue }
		var value [4]byte
		copy(value[:], ip)
		table.records[name] = value
	}
	s.active.Store(table)
	return nil
}

func validDNSName(name string) bool {
	if len(name) < 1 || len(name) > 253 { return false }
	for _, label := range strings.Split(name, ".") {
		if len(label) < 1 || len(label) > 63 || label[0]=='-' || label[len(label)-1]=='-' { return false }
		for _, char := range label {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') { return false }
		}
	}
	return true
}

// answer returns handled=false for unrelated or malformed queries (normal proxy).
// For a configured name, A returns an answer and AAAA/other types get NOERROR/NODATA.
func (s *localDNSStore) answer(query []byte) ([]byte, bool) {
	table := s.active.Load()
	if table == nil || !table.enabled || len(query)<17 { return nil, false }
	flags := binary.BigEndian.Uint16(query[2:4])
	if flags&0x8000 != 0 || flags&0x7800 != 0 { return nil, false }
	if binary.BigEndian.Uint16(query[4:6]) != 1 { return nil, false }
	if binary.BigEndian.Uint16(query[6:8]) != 0 { return nil, false }
	index := 12
	labels := make([]string, 0, 4)
	for {
		if index >= len(query) { return nil, false }
		length := int(query[index])
		index++
		if length == 0 { break }
		if length > 63 || index+length > len(query) { return nil, false }
		labels = append(labels, string(query[index:index+length]))
		index += length
	}
	if index+4 != len(query) { return nil, false } // no trailing EDNS/options
	name := strings.ToLower(strings.Join(labels, "."))
	addr, exists := table.records[name]
	if !exists { return nil, false }
	qtype := binary.BigEndian.Uint16(query[index:index+2])
	qclass := binary.BigEndian.Uint16(query[index+2:index+4])
	if qclass != 1 { return nil, false }
	answerA := qtype == 1
	// Response preserves the question exactly and never leaks local names upstream.
	response := make([]byte, len(query), len(query)+16)
	copy(response, query)
	binary.BigEndian.PutUint16(response[2:4], 0x8480|(flags&0x0100)) // QR AA RA + RD
	if answerA { binary.BigEndian.PutUint16(response[6:8], 1) }
	binary.BigEndian.PutUint16(response[8:10], 0)
	binary.BigEndian.PutUint16(response[10:12], 0)
	if answerA {
		response=append(response, 0xc0,0x0c, 0,1, 0,1, 0,0,0,60, 0,4)
		response=append(response, addr[:]...)
	}
	return response, true
}
