package lib

import (
  "time"
  "os"
  "fmt"
  "strings"

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

  Line uint
  Lines []string

  Status UIStatus

  Screen tcell.Screen
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
        }
			case tcell.KeyEscape, tcell.KeyCtrlC:
			  return
			case tcell.KeyEnter:
        if !self.Loading {
		      self._RequestLine()
        }
			case tcell.KeyUp:
        if !self.Loading {
		      self.Line = uint(imax(int(self.Line) - 1, 0))
          self._Render()
        }
			case tcell.KeyDown:
        if !self.Loading {
		      self.Line = uint(imin(int(self.Line) + 1, len(self.Lines) - 1))
          self._Render()
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
  for i, line := range self.Lines {
    self._RenderLine(0, i + 1, line, st.Underline(uint(i) == self.Line))
  }

  var status string
  if self.Status.Enabled {
    status = self.Status.Message
  } else if self.Loading {
    status = "Loading..."
  }

  footer := "[Q]uit/Esc/Ctrl+C [R]efresh Up Down Enter"
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

func (self *UI) _RequestLine() {
  // TODO: check if link, then follow, else error
  // line := self.Lines[self.Lines]
  // if line.IsLink() {
  //   self.Loading = true
  //   self._Network.Request(self.Address + line.Path)
  // } else {
  self._SetStatus("Error: cannot follow a non-link item")
  // }
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

func (self *UI) _HandleResult(result *NetworkResult) {
  self.Loading = false
  switch result.Result {
  case NetworkResultOK:
    self.Address = result.Address
    self.Line = 0
    self.Lines = self._ParseLines(result.List)
    self._Render()
  case NetworkResultError:
    self._SetStatus(fmt.Sprintf("Network error: %s", result.Error))
  }
}

func (self *UI) _ParseLines(lines []string) []string {
  // TODO: better parsing
  return lines
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
