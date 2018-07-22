package taupe

import (
	"fmt"

	"github.com/gdamore/tcell"
)

func (ui *UI) render() {
	ui.screen.Clear()

	w, h := ui.screen.Size()
	middle := h / 2

	st := tcell.StyleDefault

	header := fmt.Sprintf("Taupe: %s", ui.address)
	ui.renderLine(0, 0, ljust(header, w), st.Reverse(true))

	length := ui.getContentLength()
	offset := 0
	if ui.content.line > middle {
		offset = imin(ui.content.line-middle, length-h+2)
	}
	if ui.content.kind == NetworkEventOK {
		for i := offset; i-offset < h-2 && i < length; i++ {
			line := ui.content.lines[i]
			style := st
			if line.IsLink() {
				style = style.Underline(true)
			}
			if i == ui.content.line {
				style = style.Bold(true)
			}
			ui.renderLine(0, i-offset+1, line.ToString(), style)
		}
	} else if ui.content.kind == NetworkEventHTML {
		for i := offset; i-offset < h-2 && i < length; i++ {
			line := ui.content.html[i]
			style := st
			if i == ui.content.line {
				style = style.Bold(true)
			}
			ui.renderLine(0, i-offset+1, line, style)
		}
	}

	var status string
	if ui.status.enabled {
		status = ui.status.message
	} else if ui.loading {
		status = "Loading..."
	}

	footer := "[Q]uit/Esc/Ctrl+C [R]efresh Up Down Enter [B]ack/Backspace [F]orward [I]nput"
	if len(status) > 0 {
		footer = footer + " | " + status
	}
	ui.renderLine(0, h-1, ljust(footer, w), st.Reverse(true))

	ui.screen.Sync()
}

func (ui *UI) renderLine(x, y int, line string, style tcell.Style) {
	w, h := ui.screen.Size()

	for i := x; i < len(line) && i < w && y < h; i++ {
		ui.screen.SetContent(i, y, rune(line[i]), nil, style)
	}
}
