package core

import (
	"fmt"
	"strings"
)

// GopherEntry represents the kind of link an entry is
type GopherEntry byte

// Types of available Gopher entries
const (
	TypeFile       GopherEntry = '0'
	TypeSubMenu    GopherEntry = '1'
	TypeCCSO       GopherEntry = '2'
	TypeError      GopherEntry = '3'
	TypeBinHex     GopherEntry = '4'
	TypeDOS        GopherEntry = '5'
	TypeUUEncoded  GopherEntry = '6'
	TypeSearch     GopherEntry = '7'
	TypeTelnet     GopherEntry = '8'
	TypeBinary     GopherEntry = '9'
	TypeRedundant  GopherEntry = '+'
	TypeTelnet3270 GopherEntry = 'T'
	TypeGIF        GopherEntry = 'g'
	TypeImage      GopherEntry = 'I'

	TypeHTML          GopherEntry = 'h'
	TypeInformational GopherEntry = 'i'
	TypeSound         GopherEntry = 's'
)

// Record represents one entry in a Gopher response
type Record struct {
	Type    GopherEntry
	Display string
	Address string
	Label   string
	String  string
}

// ParseEntry parses a byte into an entry type
func ParseEntry(entry byte) GopherEntry {
	return GopherEntry(entry)
}

// ParseRecord initializes a Record by parsing the provided `source`, or fail
func ParseRecord(source string) (*Record, error) {
	record := Record{}
	if !record.parse(source) {
		return nil, fmt.Errorf("failed to parse line '%s'", source)
	}
	return &record, nil
}

func (record *Record) parse(source string) bool {
	fields := strings.Split(source, "\t")
	if len(fields[0]) < 1 {
		return false
	}
	record.Type = ParseEntry(fields[0][0])
	record.Display = fields[0][1:]
	record.Label = record.Type.getLabel()
	if record.Label != "" {
		record.String = fmt.Sprintf("[%s] %s", record.Label, record.Display)
	} else {
		record.String = record.Display
	}
	if len(fields) >= 4 {
		record.Address = fmt.Sprintf("gopher://%s:%s/?q=%s&t=%c", fields[2], fields[3], fields[1], record.Type)
	}
	return true
}

// IsLink returns if the entry can be requested to a Gopher server
func (record *Record) IsLink() bool {
	return record.Type == TypeSubMenu || record.Type == TypeHTML
}

// ToString returns a displayable representation of the Record
func (record *Record) ToString() string {
	// return fmt.Sprintf("%c %s", record.Type, record.Display)
	// return fmt.Sprintf("%s %s", record.Display, record.Address)
	return record.String
}

func (gtype GopherEntry) getLabel() string {
	switch gtype {
	case TypeSubMenu:
		return "menu"
	case TypeHTML:
		return "html"
	case TypeImage, TypeGIF:
		return "image"
	case TypeBinHex, TypeDOS, TypeUUEncoded, TypeBinary:
		return "binary"
	case TypeTelnet, TypeTelnet3270:
		return "telnet"
	case TypeError:
		return "error"
	case TypeSearch:
		return "search"
	case TypeRedundant:
		return "mirror"
	case TypeCCSO:
		return "ccso"
	case TypeFile:
		return "file"
	case TypeSound:
		return "sound"
	}
	return ""
}
