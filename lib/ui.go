package lib

import (
  "time"
  "os"
  "fmt"
  "net/url"
  "strings"
  "strconv"

  "github.com/gdamore/tcell"
)

const StatusTimeout time.Duration = 5 * time.Second

type EventHideStatus struct {
  tcell.EventTime
}

type UIStatus struct {
  Enabled bool
  Message string
  Cancel chan struct{}
}

type UI struct {
  Network *Network
  Address string

  Loading bool

  Line int
  Content int
  Lines []*Record
  HTML []string

  Status UIStatus

  Screen tcell.Screen

  WasPrevious bool
  HistoryBefore []string
  HistoryAfter []string
}

func MakeUI(network *Network) *UI {
  return &UI{Network: network}
}

func (self *UI) Start(address string) {
  self.Network.Register(self)
  self.Address = address
  self._Loop()
}

func (self *UI) _FatalError(err error) {
  fmt.Fprintf(os.Stderr, "Fatal Error: %v\n", err)
  os.Exit(1)
}

func imin(a, b int) int {
  if a < b {
    return a
  }
  return b
}

func imax(a, b int) int {
  if a > b {
    return a
  }
  return b
}

func (self *UI) _Loop() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
  if err != nil {
    self._FatalError(err)
	}
	if err = screen.Init(); err != nil {
    self._FatalError(err)
	}
  self.Screen = screen
  defer screen.Fini()

  screen.HideCursor()
  self._Render()
  self._Refresh()

  for {
    event := screen.PollEvent()
    switch event := event.(type) {
    case *tcell.EventResize:
      self._Render()

    case *EventHideStatus:
      self.Status.Enabled = false
      self._Render()

    case *tcell.EventInterrupt:
      switch result := event.Data().(type) {
      case *NetworkResult:
        self._HandleResult(result)
      }

		case *tcell.EventKey:
			switch event.Key() {
      case tcell.KeyRune:
        switch event.Rune() {
        case 'q', 'Q':
          return
        case 'r', 'R':
          if !self.Loading {
  		      self._Refresh()
          }
        case 'b', 'B':
          if !self.Loading {
            self._GoBack()
          }
        case 'f', 'F':
          if !self.Loading {
            self._GoForward()
          }
        case 'i', 'I':
          if !self.Loading {
            self._Input()
          }
        }
			case tcell.KeyEscape, tcell.KeyCtrlC:
			  return
			case tcell.KeyEnter:
        if !self.Loading {
		      self._RequestLine()
        }
			case tcell.KeyUp:
        if !self.Loading {
          self._SelectLink(-1)
          self._Render()
        }
			case tcell.KeyDown:
        if !self.Loading {
          self._SelectLink(1)
          self._Render()
        }
      case tcell.KeyBackspace:
        if !self.Loading {
          self._GoBack()
        }
			}

		}
  }
}

func ljust(s string, total int) string {
  return s + strings.Repeat(" ", total - len(s))
}

func (self *UI) _Render() {
  self.Screen.Clear()

  w, h := self.Screen.Size()

  st := tcell.StyleDefault

  header := fmt.Sprintf("Taupe: %s", self.Address)
  self._RenderLine(0, 0, ljust(header, w), st.Reverse(true))

  // TODO: scroll
  if self.Content == NetworkResultOK {
    for i, line := range self.Lines {
      style := st
      if line.IsLink() {
        style = style.Underline(true)
      }
      if i == self.Line {
        style = style.Bold(true)
      }
      self._RenderLine(0, i + 1, line.ToString(), style)
    }
  } else if self.Content == NetworkResultHTML {
    for i, line := range self.HTML {
      self._RenderLine(0, i + 1, line, st)
    }
  }

  var status string
  if self.Status.Enabled {
    status = self.Status.Message
  } else if self.Loading {
    status = "Loading..."
  }

  footer := "[Q]uit/Esc/Ctrl+C [R]efresh Up Down Enter [B]ack/Backspace [F]orward [I]nput"
  if len(status) > 0 {
    footer = footer + " | " + status
  }
  self._RenderLine(0, h - 1, ljust(footer, w), st.Reverse(true))

  self.Screen.Sync()
}

func (self *UI) _RenderLine(x, y int, line string, style tcell.Style) {
  w, h := self.Screen.Size()

  for i := x; i < len(line) && i < w && y < h; i++ {
    self.Screen.SetContent(i, y, rune(line[i]), nil, style)
  }
}

func (self *UI) _Input() {
  // TODO: manual input
}

func (self *UI) _GoBack() {
  if len(self.HistoryBefore) < 1 {
    self._SetStatus("Error: no previous page")
    return
  }
  self.Loading = true
  previous := self.HistoryBefore[0]
  self.HistoryBefore = self.HistoryBefore[1:]
  self.WasPrevious = true
  self.Network.Request(previous)
}

func (self *UI) _GoForward() {
  if len(self.HistoryAfter) < 1 {
    self._SetStatus("Error: no next page")
    return
  }
  self.Loading = true
  next := self.HistoryAfter[0]
  self.HistoryAfter = self.HistoryAfter[1:]
  self.Network.Request(next)
}

func (self *UI) _RequestLine() {
  if self.Line < 0 {
    self._SetStatus("Error: nothing selectable")
    return
  }
  line := self.Lines[self.Line]
  if line.IsLink() {
    self.Loading = true
    self.Network.Request(line.Address)
  } else {
    self._SetStatus("Error: cannot follow a non-gopher items")
  }
}

func (self *UI) _Refresh() {
  self.Loading = true
  self._Render()
  self.Network.Request(self.Address)
}

func (self *UI) OnNetworkResult(result *NetworkResult) {
  if self.Screen == nil {
    return
  }
  self.Screen.PostEvent(tcell.NewEventInterrupt(result))
}

func (self *UI) _HandleCommon(result *NetworkResult) {
  self.Content = result.Result
  history := self.Address
  oldUrl, err := url.Parse(self.Address)
  if err == nil {
    q := oldUrl.Query()
    q.Set("l", strconv.Itoa(self.Line))
    oldUrl.RawQuery = q.Encode()
    history = oldUrl.String()
  }
  if self.WasPrevious {
    self.HistoryAfter = append([]string{history}, self.HistoryAfter...)
  } else {
    self.HistoryBefore = append([]string{history}, self.HistoryBefore...)
  }
  self.Address = result.Address
  newUrl, err := url.Parse(self.Address)
  self.Line = -1
  if err == nil {
    if val, ok := newUrl.Query()["l"]; ok {
      line, err := strconv.Atoi(val[0])
      if err == nil {
        self.Line = line - 1
      }
    }
  }
}

func (self *UI) _HandleResult(result *NetworkResult) {
  self.Loading = false
  switch result.Result {
  case NetworkResultOK:
    self._HandleCommon(result)
    self.Lines = self._ParseLines(result.List)
    self._SelectLink(1)
    self._Render()
  case NetworkResultHTML:
    self._HandleCommon(result)
    self.HTML = self._ParseHTML(result.HTML)
    self._Render()
  case NetworkResultError:
    self._SetStatus(fmt.Sprintf("Network error: %s", result.Error))
  }
  self.WasPrevious = false
}

func (self *UI) _ParseHTML(html string) []string {
  lines := strings.Split(html, "\n")
  result := []string{}

  w, _ := self.Screen.Size()
  pivot := w - 2

  for _, line := range lines {
    if len(line) > w {
      result = append(result, line[:w])
      rest := line[w:]
      for {
        irest := rest
        if len(rest) > w {
          irest = rest[:pivot]
        }
        result = append(result, fmt.Sprintf("| %s", irest))
        if len(rest) > w {
          rest = rest[pivot:]
        } else {
          break
        }
      }
    } else {
      result = append(result, line)
    }
  }
  return result
}

func (self *UI) _ParseLines(lines []string) []*Record {
  result := []*Record{}
  for _, line := range lines {
    record := Record{}
    if !record.Parse(line) {
      self._SetStatus(fmt.Sprintf("Error: while parsing `%s`", line))
      return []*Record{}
    }
    result = append(result, &record)
  }
  return result
}

func (self *UI) _SelectLink(diff int) {
  for i := self.Line + diff; 0 <= i && i < len(self.Lines); i += diff {
    if self.Lines[i].IsLink() {
      self.Line = i
      break
    }
  }
 }

func (self *UI) _SetStatus(message string) {
  if self.Screen == nil {
    return
  }

  if self.Status.Enabled {
    close(self.Status.Cancel)
  }

  quit := make(chan struct{})

  self.Status.Enabled = true
  self.Status.Message = message
  self.Status.Cancel = quit
  self._Render()

  go func() {
    select {
    case <-quit:
    case <-time.After(StatusTimeout):
      self.Screen.PostEvent(&EventHideStatus{})
    }
  }()
}
