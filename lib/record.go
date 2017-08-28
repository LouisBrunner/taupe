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
}

func (self *Record) Parse(source string) bool {
  fields := strings.Split(source, TAB)
  if len(fields) < 4 {
    return false
  }
  if len(fields[0]) < 1 {
    return false
  }
  self.Type = fields[0][0]
  self.Display = fields[0][1:]
  self.Address = fmt.Sprintf("gopher://%s:%s/?q=%s&t=%c", fields[2], fields[3], fields[1], self.Type)
  return true
}

func (self *Record) IsLink() bool {
  return self.Type == TypeSubMenu || self.Type == TypeHTML
}

func (self *Record) ToString() string {
  // return fmt.Sprintf("%c %s", self.Type, self.Display)
  // return fmt.Sprintf("%s %s", self.Display, self.Address)
  return self.Display
}
