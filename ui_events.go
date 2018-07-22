package taupe

import (
	"github.com/gdamore/tcell"
)

func (ui *UI) handleKey(event *tcell.EventKey) {
	if !ui.loading {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r', 'R':
				ui.refresh()
			case 'b', 'B':
				ui.goBack()
			case 'f', 'F':
				ui.goForward()
			case 'i', 'I':
				ui.input()
			}
		case tcell.KeyEnter:
			ui.requestLine()
		case tcell.KeyUp:
			ui.selectLink(-1)
		case tcell.KeyDown:
			ui.selectLink(1)
		case tcell.KeyBackspace:
			ui.goBack()
		}
	}
}

func (ui *UI) isQuitKey(event tcell.Event) bool {
	if event, casted := event.(*tcell.EventKey); casted {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				return true
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		}
	}
	return false
}

func (ui *UI) input() {
	// TODO: manual input
}

func (ui *UI) goBack() {
	if len(ui.history.before) < 1 {
		ui.setStatus("Error: no previous page")
		return
	}
	previous := ui.history.before[0]
	ui.history.before = ui.history.before[1:]
	ui.history.wasPrevious = true
	ui.doRequest(previous)
}

func (ui *UI) goForward() {
	if len(ui.history.after) < 1 {
		ui.setStatus("Error: no next page")
		return
	}
	next := ui.history.after[0]
	ui.history.after = ui.history.after[1:]
	ui.doRequest(next)
}

func (ui *UI) requestLine() {
	if ui.content.line < 0 {
		ui.setStatus("Error: nothing selectable")
		return
	}
	line := ui.content.lines[ui.content.line]
	if line.IsLink() {
		ui.doRequest(line.Address)
	} else {
		ui.setStatus("Error: cannot follow a non-gopher items")
	}
}

func (ui *UI) refresh() {
	ui.doRequest(ui.address)
}

func (ui *UI) selectLink(diff int) {
	for i := ui.content.line + diff; 0 <= i && i < ui.getContentLength(); i += diff {
		if ui.content.kind == NetworkEventHTML || ui.content.lines[i].IsLink() {
			ui.content.line = i
			break
		}
	}
	ui.render()
}
