package taupe

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/LouisBrunner/taupe/core"
)

const crlf, eom string = "\r\n", "."

func (network *Network) doRequest(request string) *NetworkEvent {
	url, err := url.Parse(request)
	if err != nil {
		return createErrorEvent(fmt.Errorf("invalid url `%s`: %s", request, err))
	}
	if url.Scheme != "gopher" && url.Scheme != "" {
		return createErrorEvent(fmt.Errorf("invalid scheme `%s`", url.Scheme))
	}
	if url.Host == "" {
		return createErrorEvent(fmt.Errorf("missing host for `%s`", request))
	}

	port := "70"
	if url.Port() != "" {
		port = url.Port()
	}

	host := fmt.Sprintf("%s:%s", url.Hostname(), port)
	conn, err := net.Dial("tcp", host)
	defer conn.Close()
	if err != nil {
		return createErrorEvent(fmt.Errorf("cannot connect to `%s`: %s", host, err))
	}

	path := ""
	if val, ok := url.Query()["q"]; ok {
		path = val[0]
	}
	fmt.Fprintf(conn, fmt.Sprintf("%s%s", path, crlf))

	reader := bufio.NewReader(conn)

	linkType := core.TypeSubMenu
	if val, ok := url.Query()["t"]; ok {
		linkType = core.ParseEntry(val[0][0])
	}

	// TODO: support images, binaries...
	var event *NetworkEvent
	if linkType == core.TypeHTML {
		event, err = network.parseHTML(request, reader)
	} else {
		event, err = network.parseGopher(request, reader)
	}
	if err != nil {
		return createErrorEvent(err)
	}
	return event
}

func createErrorEvent(err error) *NetworkEvent {
	return &NetworkEvent{Event: NetworkEventError, ResultError: err}
}

func (network *Network) parseHTML(request string, reader io.Reader) (*NetworkEvent, error) {
	html, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return &NetworkEvent{
		Event:      NetworkEventHTML,
		ResultHTML: &NetworkResultHTML{Address: request, HTML: string(html)},
	}, nil
}

func (network *Network) parseGopher(request string, reader *bufio.Reader) (*NetworkEvent, error) {
	lines := []string{}

	for {
		char, err := reader.Peek(1)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("while reading EOM: %s", err)
		}
		if string(char) == eom || err == io.EOF {
			break
		}

		line := ""
		for {
			result, err := reader.ReadString(crlf[0])
			if err != nil {
				return nil, fmt.Errorf("while reading line: %s", err)
			}
			line += result

			char, err := reader.Peek(1)
			if err != nil {
				return nil, fmt.Errorf("while peeking LF: %s", err)
			}
			if char[0] == crlf[1] {
				char, err := reader.ReadByte()
				if err != nil {
					return nil, fmt.Errorf("while reading LF: %s", err)
				}
				line += string(char)
				break
			}
		}
		lines = append(lines, strings.Replace(line, crlf, "", -1))
	}

	return &NetworkEvent{
		Event:  NetworkEventOK,
		Result: &NetworkResult{Address: request, List: lines},
	}, nil
}
