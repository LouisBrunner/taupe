package lib

import (
  "fmt"
  "strings"
)

// Types of available Gopher entries
const (
  TypeFile byte = '0'
  TypeSubMenu byte = '1'
  TypeCCSO byte = '2'
  TypeError byte = '3'
  TypeBinHex byte = '4'
  TypeDOS byte = '5'
  TypeUUEncoded byte = '6'
  TypeSearch byte = '7'
  TypeTelnet byte = '8'
  TypeBinary byte = '9'
  TypeRedundant byte = '+'
  TypeTelnet3270 byte = 'T'
  TypeGIF byte = 'g'
  TypeImage byte = 'I'

  TypeHTML byte = 'h'
  TypeInformational byte = 'i'
  TypeSound byte = 's'
)

const tab string = "\t"

// Record represents one entry in a Gopher response
type Record struct {
  Type byte
  Display string
  Address string
  Label string
  String string
}

// Parse initialized the Record by parsing the provided `source`
func (record *Record) Parse(source string) bool {
  fields := strings.Split(source, tab)
  if len(fields[0]) < 1 {
    return false
  }
  record.Type = fields[0][0]
  record.Display = fields[0][1:]
  record.Label = record.getLabel(record.Type)
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

func (record *Record) getLabel(gtype byte) string {
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
