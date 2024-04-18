package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/moisespsena-go/httpdx/internal"
)

type Listener struct {
	id string
	l  net.Listener
}

func Run(cfg *Config) (err error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	serverURL := cfg.ServerURL + internal.ProxyPath

	if _, err = url.Parse(serverURL); err != nil {
		return
	}

	log.Println("URL: " + serverURL)

	c, _, err := websocket.DefaultDialer.Dial(serverURL+"?name="+internal.TestService, nil)
	if err != nil {
		return fmt.Errorf("test_service dial: %s", err)
	}

	var (
		pingPayload = "!!test!!"
	)

	c.SetPongHandler(func(appData string) error {
		if appData != pingPayload {
			return fmt.Errorf("test_service: expected ping message payload")
		}
		return nil
	})
	if err = c.WriteMessage(websocket.PingMessage, []byte(pingPayload)); err != nil {
		return fmt.Errorf("test_service: read message: %s", err)
	}
	c.Close()

	var (
		listeners []*Listener
		done      = make(chan int)
		doneCount int
	)

	for i, service := range cfg.Services {
		if l := runService(func() {
			done <- i + 1
		}, i, serverURL, service.Name, service.LocalAddr); l != nil {
			listeners = append(listeners, l)
		}
	}

	for doneCount < len(listeners) {
		select {
		case <-interrupt:
			for _, l := range listeners {
				if l != nil {
					l.l.Close()
				}
			}
		case i := <-done:
			listeners[i-1] = nil
			doneCount++
		}
	}

	return
}

func runService(done func(), i int, serverURL, name, localAddr string) (_ *Listener) {
	id := "[" + strconv.Itoa(i) + "]" + name + "@" + localAddr
	log.Println(id, "started")

	l, err := net.Listen("tcp4", localAddr)
	if err != nil {
		log.Printf("%s: listen: %v", id, err)
		return
	}

	go func() {
		defer func() {
			l.Close()
			log.Println(id, "done")
			done()
		}()
		for {
			c, err := l.Accept()
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					log.Printf("%s: accept: %v", id, err)
				}
				return
			}
			go handleConnection(serverURL+"?name="+name, id, c)
		}
	}()

	return &Listener{id, l}
}

func handleConnection(serverUrl, id string, con net.Conn) {
	log.Printf("%s: serving %s", id, con.RemoteAddr().String())

	c, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		log.Printf(id+": dial: %v", err)
		return
	}
	defer c.Close()

	go func() {
		var read = func() (msg []byte, err error) {
			defer func() {
				if r := recover(); r != nil {
					switch t := r.(type) {
					case string:
						if t == "repeated read on failed websocket connection" {
							err = nil
						} else {
							err = errors.New(t)
						}
					case error:
						err = t
					default:
						err = fmt.Errorf("%v", t)
					}
				}
			}()
			_, msg, err = c.ReadMessage()
			return
		}
		for {
			message, err := read()
			if err != nil {
				if !strings.HasSuffix(err.Error(), "use of closed network connection") {
					log.Printf(id+": read_message: %v", err)
				}
				con.Close()
				return
			} else {
				if _, err := con.Write(message); err != nil {
					log.Printf(id+": write_message: %v", err)
					return
				}
			}
		}
	}()

	defer func() {
		log.Printf("%s: %s: done", id, con.RemoteAddr().String())
	}()
	io.Copy(&wsw{c: c}, con)
}

type wsw struct {
	c *websocket.Conn
}

func (w *wsw) Write(p []byte) (_ int, err error) {
	err = w.c.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}