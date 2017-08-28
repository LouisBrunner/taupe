package lib
// import socket
// from urllib.parse import urlparse
// from threading import Thread, Event, Condition

import (
  "sync"
  "fmt"

  "net"
  "net/url"

  "io"
  "bufio"
  "strings"
)

const (
  NetworkResultOK = iota
  NetworkResultHTML
  NetworkResultError
)

const CR_LF, EOM string = "\r\n", "."

type NetworkResult struct {
  Result int
  Address string
  List []string
  HTML string
  Error string
}

type Listener interface {
  OnNetworkResult(result *NetworkResult)
}

type Network struct {
  Mutex sync.Mutex
  Listeners []Listener

  Signals chan string
  Data chan string

  PendingRequest string
  StopFlag bool
}

func MakeNetwork() *Network {
  return &Network{
    Signals: make(chan string),
    Data: make(chan string),
    StopFlag: false,
  }
}

const StopEvent, CancelEvent string = "STOP_EVENT", "CANCEL_EVENT"

func (self *Network) Register(listener Listener) {
  self.Mutex.Lock()
  defer self.Mutex.Unlock()
  self.Listeners = append(self.Listeners, listener)
}

func (self *Network) Start() {
  go self._Loop()
}

func (self *Network) Stop() {
  self.Signals <- StopEvent
}

func (self *Network) Request(address string) {
  self.Signals <- CancelEvent
  self.Data <- address
}

func (self *Network) _Loop() {
  for {
    select {
    case signal := <-self.Signals:
      if signal == StopEvent {
        self.StopFlag = true
      } else if signal == CancelEvent {
        continue
      }
    case request := <-self.Data:
      self._MakeRequest(request)
    }
    if self.StopFlag {
      break
    }
  }
}

func (self *Network) _MakeRequest(request string) {
  url, err := url.Parse(request)
  if err != nil {
    self._SendError(fmt.Sprintf("invalid url `%s`: %s", request, err))
    return
  }
  if url.Scheme != "gopher" && url.Scheme != "" {
    self._SendError(fmt.Sprintf("invalid scheme `%s`", url.Scheme))
    return
  }

  if url.Host == "" {
    self._SendError(fmt.Sprintf("missing host for `%s`", request))
    return
  }

  port := "70"
  if url.Port() != "" {
    port = url.Port()
  }

  host := fmt.Sprintf("%s:%s", url.Hostname(), port)
  conn, err := net.Dial("tcp", host)
  defer conn.Close()
  if err != nil {
    self._SendError(fmt.Sprintf("cannot connect to `%s`: %s", host, err))
    return
  }
  if self._ShouldStop() {
    return
  }

  path := ""
  if val, ok := url.Query()["q"]; ok {
    path = val[0]
  }
  fmt.Fprintf(conn, fmt.Sprintf("%s%s", path, CR_LF))

  reader := bufio.NewReader(conn)

  linkType := TypeSubMenu
  if val, ok := url.Query()["t"]; ok {
    linkType = val[0][0]
  }

  // TODO: support images, binaries...
  if linkType == TypeHTML {
    self._ParseHTML(request, reader)
  } else {
    self._ParseGopher(request, reader)
  }
}

func (self *Network) _ParseHTML(request string, reader *bufio.Reader) {
  html := ""
  buffer := make([]byte, 1024)
  for {
    if self._ShouldStop() {
      return
    }
    n, err := reader.Read(buffer)

    if err != nil && err != io.EOF {
      self._SendError(fmt.Sprintf("while reading HTML: %s", err))
      return
    }

    if n == 0 && err == io.EOF {
      break
    }

    html += string(buffer[:n])
  }

  self._SendHTML(request, html)
}

func (self *Network) _ParseGopher(request string, reader *bufio.Reader) {
  lines := []string{}

  for {
    if self._ShouldStop() {
      return
    }

    char, err := reader.Peek(1)
    if err != nil && err != io.EOF {
      self._SendError(fmt.Sprintf("while reading EOM: %s", err))
      return
    }
    if string(char) == EOM || err == io.EOF {
      break
    }

    line := ""
    for {
      if self._ShouldStop() {
        return
      }

      result, err := reader.ReadString(CR_LF[0])
      if err != nil {
        self._SendError(fmt.Sprintf("while reading line: %s", err))
        return
      }
      line += result

      char, err := reader.Peek(1)
      if err != nil {
        self._SendError(fmt.Sprintf("while peeking LF: %s", err))
        return
      }
      if char[0] == CR_LF[1] {
        char, err := reader.ReadByte()
        if err != nil {
          self._SendError(fmt.Sprintf("while reading LF: %s", err))
          return
        }
        line += string(char)
        break
      }
    }
    lines = append(lines, strings.Replace(line, CR_LF, "", -1))
  }

  self._SendResult(request, lines)
}

func (self *Network) _ShouldStop() (stop bool) {
  stop = false

  select {
  case signal := <-self.Signals:
    if signal == StopEvent {
      self.StopFlag = true
      stop = true
    } else if signal == CancelEvent {
      stop = true
    }
  default:
  }

  return
}

func (self *Network) _SendHTML(address string, html string) {
  if self._ShouldStop() {
    return
  }
  self._SendMessage(&NetworkResult{Result: NetworkResultHTML, Address: address, HTML: html})
}

func (self *Network) _SendResult(address string, list []string) {
  if self._ShouldStop() {
    return
  }
  self._SendMessage(&NetworkResult{Result: NetworkResultOK, Address: address, List: list})
}

func (self *Network) _SendError(message string) {
  if self._ShouldStop() {
    return
  }
  self._SendMessage(&NetworkResult{Result: NetworkResultError, Error: message})
}

func (self *Network) _SendMessage(result *NetworkResult) {
  self.Mutex.Lock()
  defer self.Mutex.Unlock()
  for _, listener := range self.Listeners {
    listener.OnNetworkResult(result)
  }
}
