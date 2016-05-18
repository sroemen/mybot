/*

mybot - Illustrative Slack bot in Go

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// http vars
	client    *http.Client
	transport *http.Transport
)

// This does what the name implies, creates a Dialer that will timeout after
// it's configured time period.
// used by http.transport()
func timedDialer(tout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		c, err := net.DialTimeout(netw, addr, tout)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

func init() {
	// and adds a timeout, ignore self signed certs
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            timedDialer(6 * time.Second),
	}
	client = &http.Client{Transport: transport}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")

	for {
		// read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if we're mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)

			log.Printf("channel: %s message::  %v", m.Channel, parts)

			if len(parts) >= 2 && isValidCommand(parts[1]) {
				// looks good, continue
				args := make([]string, 0)
				v := 0

				go func(m Message) {
					if len(parts) > 1 {
						for x, _ := range parts {
							if x <= 1 {
								continue
							}
							args = append(args, parts[x])
							v++
						}
					}
					m.Text = runCommand(parts[1], args)
					postMessage(ws, m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else {
				// unknown command, send help command back
				m.Text = runCommand("help", []string{})
				postMessage(ws, m)
			}
		} else if m.Type == "message" && isValidHashCommand(strings.Split(m.Text, " ")) {
			// if so try to parse if
			parts := strings.Fields(m.Text)
			command := parts[0]
	
			log.Printf("channel: %s message::  %v", m.Channel, parts)

			if len(parts) >= 1 {
				// looks good, continue
				args := make([]string, 0)
				v := 0

				go func(m Message) {
					if len(parts) > 1 {
						for x, _ := range parts {
							if x < 1 {
								continue
							}
							args = append(args, parts[x])
							v++
						}
					}
					m.Text = runCommand(command[1:], args)
					postMessage(ws, m)
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else {
				// unknown command, send help command back
				m.Text = runCommand("help", []string{})
				postMessage(ws, m)
			}
		}
	}
}
