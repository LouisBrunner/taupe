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

const statusTimeout time.Duration = 5 * time.Second

type eventHideStatus struct {
  tcell.EventTime
}

type uiStatus struct {
  Enabled bool
  Message string
  Cancel chan struct{}
}

// UI represents the ncurses user interface that someone use to interact with the Gophernet
type UI struct {
  Network *Network
  Address string

  Loading bool

  Line int
  Content int
  Lines []*Record
  HTML []string

  Status uiStatus

  Screen tcell.Screen

  WasPrevious bool
  HistoryBefore []string
  HistoryAfter []string
}

// MakeUI construct a UI correctly initialized
func MakeUI(network *Network) *UI {
  return &UI{Network: network}
}

// Start registers the UI with the Network (to get responses) and starts the internal loop
func (ui *UI) Start(address string) {
  ui.Network.Register(ui)
  ui.Address = address
  ui._Loop()
}

func (ui *UI) _FatalError(err error) {
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

func (ui *UI) _Loop() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
  if err != nil {
    ui._FatalError(err)
	}
	if err = screen.Init(); err != nil {
    ui._FatalError(err)
	}
  ui.Screen = screen
  defer screen.Fini()

  screen.HideCursor()
  ui._Render()
  ui._Refresh()

  for {
    event := screen.PollEvent()
    switch event := event.(type) {
    case *tcell.EventResize:
      ui._Render()

    case *eventHideStatus:
      ui.Status.Enabled = false
      ui._Render()

    case *tcell.EventInterrupt:
      switch result := event.Data().(type) {
      case *NetworkResult:
        ui._HandleResult(result)
      }

		case *tcell.EventKey:
			switch event.Key() {
      case tcell.KeyRune:
        switch event.Rune() {
        case 'q', 'Q':
          return
        case 'r', 'R':
          if !ui.Loading {
  		      ui._Refresh()
          }
        case 'b', 'B':
          if !ui.Loading {
            ui._GoBack()
          }
        case 'f', 'F':
          if !ui.Loading {
            ui._GoForward()
          }
        case 'i', 'I':
          if !ui.Loading {
            ui._Input()
          }
        }
			case tcell.KeyEscape, tcell.KeyCtrlC:
			  return
			case tcell.KeyEnter:
        if !ui.Loading {
		      ui._RequestLine()
        }
			case tcell.KeyUp:
        if !ui.Loading {
          ui._SelectLink(-1)
          ui._Render()
        }
			case tcell.KeyDown:
        if !ui.Loading {
          ui._SelectLink(1)
          ui._Render()
        }
      case tcell.KeyBackspace:
        if !ui.Loading {
          ui._GoBack()
        }
			}

		}
  }
}

func ljust(s string, total int) string {
  return s + strings.Repeat(" ", total - len(s))
}

func (ui *UI) _Render() {
  ui.Screen.Clear()

  w, h := ui.Screen.Size()
  middle := h / 2

  st := tcell.StyleDefault

  header := fmt.Sprintf("Taupe: %s", ui.Address)
  ui._RenderLine(0, 0, ljust(header, w), st.Reverse(true))

  length := ui._GetLength()
  offset := 0
  if ui.Line > middle {
    offset = imin(ui.Line - middle, length - h + 2)
  }
  if ui.Content == NetworkResultOK {
    for i := offset; i - offset < h - 2 && i < length; i++ {
      line := ui.Lines[i]
      style := st
      if line.IsLink() {
        style = style.Underline(true)
      }
      if i == ui.Line {
        style = style.Bold(true)
      }
      ui._RenderLine(0, i - offset + 1, line.ToString(), style)
    }
  } else if ui.Content == NetworkResultHTML {
    for i := offset; i - offset < h - 2 && i < length; i++ {
      line := ui.HTML[i]
      style := st
      if i == ui.Line {
        style = style.Bold(true)
      }
      ui._RenderLine(0, i - offset + 1, line, style)
    }
  }

  var status string
  if ui.Status.Enabled {
    status = ui.Status.Message
  } else if ui.Loading {
    status = "Loading..."
  }

  footer := "[Q]uit/Esc/Ctrl+C [R]efresh Up Down Enter [B]ack/Backspace [F]orward [I]nput"
  if len(status) > 0 {
    footer = footer + " | " + status
  }
  ui._RenderLine(0, h - 1, ljust(footer, w), st.Reverse(true))

  ui.Screen.Sync()
}

func (ui *UI) _RenderLine(x, y int, line string, style tcell.Style) {
  w, h := ui.Screen.Size()

  for i := x; i < len(line) && i < w && y < h; i++ {
    ui.Screen.SetContent(i, y, rune(line[i]), nil, style)
  }
}

func (ui *UI) _Input() {
  // TODO: manual input
}

func (ui *UI) _GoBack() {
  if len(ui.HistoryBefore) < 1 {
    ui._SetStatus("Error: no previous page")
    return
  }
  ui.Loading = true
  previous := ui.HistoryBefore[0]
  ui.HistoryBefore = ui.HistoryBefore[1:]
  ui.WasPrevious = true
  ui.Network.Request(previous)
}

func (ui *UI) _GoForward() {
  if len(ui.HistoryAfter) < 1 {
    ui._SetStatus("Error: no next page")
    return
  }
  ui.Loading = true
  next := ui.HistoryAfter[0]
  ui.HistoryAfter = ui.HistoryAfter[1:]
  ui.Network.Request(next)
}

func (ui *UI) _RequestLine() {
  if ui.Line < 0 {
    ui._SetStatus("Error: nothing selectable")
    return
  }
  line := ui.Lines[ui.Line]
  if line.IsLink() {
    ui.Loading = true
    ui.Network.Request(line.Address)
  } else {
    ui._SetStatus("Error: cannot follow a non-gopher items")
  }
}

func (ui *UI) _Refresh() {
  ui.Loading = true
  ui._Render()
  ui.Network.Request(ui.Address)
}

// OnNetworkResult is the function called with the network response after Network.Request was called (async)
func (ui *UI) OnNetworkResult(result *NetworkResult) {
  if ui.Screen == nil {
    return
  }
  ui.Screen.PostEvent(tcell.NewEventInterrupt(result))
}

func (ui *UI) _HandleCommon(result *NetworkResult) {
  ui.Content = result.Result
  history := ui.Address
  oldURL, err := url.Parse(ui.Address)
  if err == nil {
    q := oldURL.Query()
    q.Set("l", strconv.Itoa(ui.Line))
    oldURL.RawQuery = q.Encode()
    history = oldURL.String()
  }
  if ui.WasPrevious {
    ui.HistoryAfter = append([]string{history}, ui.HistoryAfter...)
  } else {
    ui.HistoryBefore = append([]string{history}, ui.HistoryBefore...)
  }
  ui.Address = result.Address
  newURL, err := url.Parse(ui.Address)
  ui.Line = -1
  if err == nil {
    if val, ok := newURL.Query()["l"]; ok {
      line, err := strconv.Atoi(val[0])
      if err == nil {
        ui.Line = line - 1
      }
    }
  }
}

func (ui *UI) _HandleResult(result *NetworkResult) {
  ui.Loading = false
  switch result.Result {
  case NetworkResultOK:
    ui._HandleCommon(result)
    ui.Lines = ui._ParseLines(result.List)
    ui._SelectLink(1)
    ui._Render()
  case NetworkResultHTML:
    ui._HandleCommon(result)
    ui.HTML = ui._ParseHTML(result.HTML)
    ui._Render()
  case NetworkResultError:
    ui._SetStatus(fmt.Sprintf("Network error: %s", result.Error))
  }
  ui.WasPrevious = false
}

func (ui *UI) _ParseHTML(html string) []string {
  lines := strings.Split(html, "\n")
  result := []string{}

  w, _ := ui.Screen.Size()
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

func (ui *UI) _ParseLines(lines []string) []*Record {
  result := []*Record{}
  for _, line := range lines {
    record := Record{}
    if !record.Parse(line) {
      ui._SetStatus(fmt.Sprintf("Error: while parsing `%s`", line))
      return []*Record{}
    }
    result = append(result, &record)
  }
  return result
}

func (ui *UI) _GetLength() int {
  length := 0
  if ui.Content == NetworkResultOK {
    length = len(ui.Lines)
  } else if ui.Content == NetworkResultHTML {
    length = len(ui.HTML)
  }
  return length
}

func (ui *UI) _SelectLink(diff int) {
  for i := ui.Line + diff; 0 <= i && i < ui._GetLength(); i += diff {
    if ui.Content == NetworkResultHTML || ui.Lines[i].IsLink() {
      ui.Line = i
      break
    }
  }
 }

func (ui *UI) _SetStatus(message string) {
  if ui.Screen == nil {
    return
  }

  if ui.Status.Enabled {
    close(ui.Status.Cancel)
  }

  quit := make(chan struct{})

  ui.Status.Enabled = true
  ui.Status.Message = message
  ui.Status.Cancel = quit
  ui._Render()

  go func() {
    select {
    case <-quit:
    case <-time.After(statusTimeout):
      ui.Screen.PostEvent(&eventHideStatus{})
    }
  }()
}
