package taupe

import (
	"fmt"
	"os"
	"time"

	"github.com/LouisBrunner/taupe/core"

	"github.com/gdamore/tcell"
)

type uiStatus struct {
	enabled bool
	message string
	created time.Time
}

type uiHistory struct {
	wasPrevious bool
	before      []string
	after       []string
}

type uiContent struct {
	line  int
	kind  NetworkEventType
	lines []*core.Record
	html  []string
}

// UI represents the ncurses user interface that someone use to interact with the Gophernet
type UI struct {
	screen  tcell.Screen
	address string
	loading bool
	network NetworkManager
	request <-chan *NetworkEvent
	content uiContent
	status  uiStatus
	history uiHistory
}

// NewUI construct a UI correctly initialized
func NewUI(network NetworkManager) *UI {
	return &UI{network: network}
}

// Run registers the UI with the Network (to get responses) and starts the internal loop
func (ui *UI) Run(address string) {
	ui.address = address
	ui.run()
}

func (ui *UI) fatalError(err error) {
	fmt.Fprintf(os.Stderr, "Fatal Error: %v\n", err)
	os.Exit(1)
}

func (ui *UI) run() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	screen, err := tcell.NewScreen()
	if err != nil {
		ui.fatalError(err)
	}
	if err = screen.Init(); err != nil {
		ui.fatalError(err)
	}
	ui.screen = screen
	defer screen.Fini()

	screen.HideCursor()
	ui.render()
	ui.refresh()

	uiEvents := make(chan tcell.Event)
	go func() {
		for {
			event := screen.PollEvent()
			uiEvents <- event
			if ui.isQuitKey(event) {
				break
			}
		}
	}()

out:
	for {
		select {
		case event := <-ui.request:
			ui.parseNetworkEvent(event)
		case event := <-uiEvents:
			if ui.isQuitKey(event) {
				break out
			}
			switch event := event.(type) {
			case *tcell.EventResize:
				ui.render()
			case *tcell.EventKey:
				ui.handleKey(event)
			}
		case <-time.After(100 * time.Millisecond):
		}
		if time.Since(ui.status.created) >= 5*time.Second {
			ui.status.enabled = false
			ui.render()
		}
	}
}

func (ui *UI) getContentLength() int {
	length := 0
	if ui.content.kind == NetworkEventOK {
		length = len(ui.content.lines)
	} else if ui.content.kind == NetworkEventHTML {
		length = len(ui.content.html)
	}
	return length
}

func (ui *UI) setStatus(message string) {
	ui.status = uiStatus{
		enabled: true,
		message: message,
		created: time.Now(),
	}
	ui.render()
}
