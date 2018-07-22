package taupe

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/LouisBrunner/taupe/core"
)

func (ui *UI) doRequest(address string) {
	ui.loading = true
	ui.request = ui.network.Request(address)
	ui.render()
}

func (ui *UI) parseNetworkEvent(event *NetworkEvent) {
	ui.loading = false
	switch event.Event {
	case NetworkEventOK:
		result := event.Result
		ui.parseNetworkCommon(event.Event, result.Address)
		ui.content.lines = ui.parseLines(result.List)
		ui.selectLink(1)
	case NetworkEventHTML:
		result := event.ResultHTML
		ui.parseNetworkCommon(event.Event, result.Address)
		ui.content.html = ui.parseHTML(result.HTML)
		ui.render()
	case NetworkEventError:
		ui.setStatus(fmt.Sprintf("Network error: %v", event.ResultError))
	}
	ui.history.wasPrevious = false
}

func (ui *UI) parseNetworkCommon(event NetworkEventType, address string) {
	ui.content.kind = event
	history := ui.address
	oldURL, err := url.Parse(ui.address)
	if err == nil {
		q := oldURL.Query()
		q.Set("l", strconv.Itoa(ui.content.line))
		oldURL.RawQuery = q.Encode()
		history = oldURL.String()
	}
	if ui.history.wasPrevious {
		ui.history.after = append([]string{history}, ui.history.after...)
	} else {
		ui.history.before = append([]string{history}, ui.history.before...)
	}
	ui.address = address
	newURL, err := url.Parse(ui.address)
	ui.content.line = -1
	if err == nil {
		if val, ok := newURL.Query()["l"]; ok {
			line, err := strconv.Atoi(val[0])
			if err == nil {
				ui.content.line = line - 1
			}
		}
	}
}

func (ui *UI) parseHTML(html string) []string {
	lines := strings.Split(html, "\n")
	result := []string{}

	w, _ := ui.screen.Size()
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

func (ui *UI) parseLines(lines []string) []*core.Record {
	result := []*core.Record{}
	for _, line := range lines {
		record, err := core.ParseRecord(line)
		if err != nil {
			ui.setStatus(fmt.Sprintf("Error: while parsing `%s`", line))
			return []*core.Record{}
		}
		result = append(result, record)
	}
	return result
}
