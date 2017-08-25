package lib
// import socket
// from urllib.parse import urlparse
// from threading import Thread, Event, Condition

import (
  "sync"
  "fmt"
  "net/url"
)

const (
  NetworkResultOK = iota
  NetworkResultError
)

type NetworkResult struct {
  Result int
  Address string
  List []string
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
    self._SendError(fmt.Sprintf("invalid scheme `%s` %d", url.Scheme, len(url.Scheme)))
    return
  }

  // TODO: do network request
  if self._ShouldStop() {
    return
  }

  self._SendResult(request, []string{"abc", "cde", "fde"})
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

func (self *Network) _SendResult(address string, list []string) {
  self._SendMessage(&NetworkResult{Result: NetworkResultOK, Address: address, List: list})
}

func (self *Network) _SendError(message string) {
  self._SendMessage(&NetworkResult{Result: NetworkResultError, Error: message})
}

func (self *Network) _SendMessage(result *NetworkResult) {
  self.Mutex.Lock()
  defer self.Mutex.Unlock()
  for _, listener := range self.Listeners {
    listener.OnNetworkResult(result)
  }
}
