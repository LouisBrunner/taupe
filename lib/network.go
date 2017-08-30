package lib

import (
  "sync"
  "fmt"

  "net"
  "net/url"

  "io"
  "bufio"
  "strings"
)

// Type of possible network results
const (
  NetworkResultOK = iota
  NetworkResultHTML
  NetworkResultError
)

const crlf, eom string = "\r\n", "."

const stopEvent, cancelEvent string = "STOP_EVENT", "CANCEL_EVENT"

// NetworkResult represents an answer from a request sent to the Network class
type NetworkResult struct {
  Result int
  Address string
  List []string
  HTML string
  Error string
}

type listener interface {
  OnNetworkResult(result *NetworkResult)
}

// Network starts its own thread to handle network requests asynchronously
type Network struct {
  Mutex sync.Mutex
  Listeners []listener

  Signals chan string
  Data chan string

  PendingRequest string
  StopFlag bool
}

// MakeNetwork builds a valid Network structure with channels, etc
func MakeNetwork() *Network {
  return &Network{
    Signals: make(chan string),
    Data: make(chan string),
    StopFlag: false,
  }
}

// Register adds the provided `listener` in its internal listeners list
func (network *Network) Register(listener listener) {
  network.Mutex.Lock()
  defer network.Mutex.Unlock()
  network.Listeners = append(network.Listeners, listener)
}

// Start spawns the Network internal loop in a thread
func (network *Network) Start() {
  go network._Loop()
}

// Stop sends the stop event to the internal network thread
func (network *Network) Stop() {
  network.Signals <- stopEvent
}

// Request sends a cancel event for the current request and starts a new request to the provided `address`
func (network *Network) Request(address string) {
  network.Signals <- cancelEvent
  network.Data <- address
}

func (network *Network) _Loop() {
  for {
    select {
    case signal := <-network.Signals:
      if signal == stopEvent {
        network.StopFlag = true
      } else if signal == cancelEvent {
        continue
      }
    case request := <-network.Data:
      network._MakeRequest(request)
    }
    if network.StopFlag {
      break
    }
  }
}

func (network *Network) _MakeRequest(request string) {
  url, err := url.Parse(request)
  if err != nil {
    network._SendError(fmt.Sprintf("invalid url `%s`: %s", request, err))
    return
  }
  if url.Scheme != "gopher" && url.Scheme != "" {
    network._SendError(fmt.Sprintf("invalid scheme `%s`", url.Scheme))
    return
  }

  if url.Host == "" {
    network._SendError(fmt.Sprintf("missing host for `%s`", request))
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
    network._SendError(fmt.Sprintf("cannot connect to `%s`: %s", host, err))
    return
  }
  if network._ShouldStop() {
    return
  }

  path := ""
  if val, ok := url.Query()["q"]; ok {
    path = val[0]
  }
  fmt.Fprintf(conn, fmt.Sprintf("%s%s", path, crlf))

  reader := bufio.NewReader(conn)

  linkType := TypeSubMenu
  if val, ok := url.Query()["t"]; ok {
    linkType = val[0][0]
  }

  // TODO: support images, binaries...
  if linkType == TypeHTML {
    network._ParseHTML(request, reader)
  } else {
    network._ParseGopher(request, reader)
  }
}

func (network *Network) _ParseHTML(request string, reader *bufio.Reader) {
  html := ""
  buffer := make([]byte, 1024)
  for {
    if network._ShouldStop() {
      return
    }
    n, err := reader.Read(buffer)

    if err != nil && err != io.EOF {
      network._SendError(fmt.Sprintf("while reading HTML: %s", err))
      return
    }

    if n == 0 && err == io.EOF {
      break
    }

    html += string(buffer[:n])
  }

  network._SendHTML(request, html)
}

func (network *Network) _ParseGopher(request string, reader *bufio.Reader) {
  lines := []string{}

  for {
    if network._ShouldStop() {
      return
    }

    char, err := reader.Peek(1)
    if err != nil && err != io.EOF {
      network._SendError(fmt.Sprintf("while reading EOM: %s", err))
      return
    }
    if string(char) == eom || err == io.EOF {
      break
    }

    line := ""
    for {
      if network._ShouldStop() {
        return
      }

      result, err := reader.ReadString(crlf[0])
      if err != nil {
        network._SendError(fmt.Sprintf("while reading line: %s", err))
        return
      }
      line += result

      char, err := reader.Peek(1)
      if err != nil {
        network._SendError(fmt.Sprintf("while peeking LF: %s", err))
        return
      }
      if char[0] == crlf[1] {
        char, err := reader.ReadByte()
        if err != nil {
          network._SendError(fmt.Sprintf("while reading LF: %s", err))
          return
        }
        line += string(char)
        break
      }
    }
    lines = append(lines, strings.Replace(line, crlf, "", -1))
  }

  network._SendResult(request, lines)
}

func (network *Network) _ShouldStop() (stop bool) {
  stop = false

  select {
  case signal := <-network.Signals:
    if signal == stopEvent {
      network.StopFlag = true
      stop = true
    } else if signal == cancelEvent {
      stop = true
    }
  default:
  }

  return
}

func (network *Network) _SendHTML(address string, html string) {
  if network._ShouldStop() {
    return
  }
  network._SendMessage(&NetworkResult{Result: NetworkResultHTML, Address: address, HTML: html})
}

func (network *Network) _SendResult(address string, list []string) {
  if network._ShouldStop() {
    return
  }
  network._SendMessage(&NetworkResult{Result: NetworkResultOK, Address: address, List: list})
}

func (network *Network) _SendError(message string) {
  if network._ShouldStop() {
    return
  }
  network._SendMessage(&NetworkResult{Result: NetworkResultError, Error: message})
}

func (network *Network) _SendMessage(result *NetworkResult) {
  network.Mutex.Lock()
  defer network.Mutex.Unlock()
  for _, listener := range network.Listeners {
    listener.OnNetworkResult(result)
  }
}
