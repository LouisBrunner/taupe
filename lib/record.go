package lib

import (
  "fmt"
  "strings"
)

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

const TAB string = "\t"

type Record struct {
  Type byte
  Display string
  Address string
  Label string
  String string
}

func (self *Record) Parse(source string) bool {
  fields := strings.Split(source, TAB)
  if len(fields[0]) < 1 {
    return false
  }
  self.Type = fields[0][0]
  self.Display = fields[0][1:]
  self.Label = self._GetLabel()
  if self.Label != "" {
    self.String = fmt.Sprintf("[%s] %s", self.Label, self.Display)
  } else {
    self.String = self.Display
  }
  if len(fields) >= 4 {
    self.Address = fmt.Sprintf("gopher://%s:%s/?q=%s&t=%c", fields[2], fields[3], fields[1], self.Type)
  }
  return true
}

func (self *Record) IsLink() bool {
  return self.Type == TypeSubMenu || self.Type == TypeHTML
}

func (self *Record) ToString() string {
  // return fmt.Sprintf("%c %s", self.Type, self.Display)
  // return fmt.Sprintf("%s %s", self.Display, self.Address)
  return self.String
}

func (self *Record) _GetLabel() string {
  switch self.Type {
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
