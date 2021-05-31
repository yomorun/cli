package serverless

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/reactivex/rxgo/v2"
	"github.com/yomorun/yomo/pkg/client"
	"github.com/yomorun/yomo/pkg/quic"
)

const (
	StreamTypeSource string = "source"
	StreamTypeFlow   string = "flow"
	StreamTypeSink   string = "sink"
)

var GlobalApp = ""

// QuicConn represents the QUIC connection.
type QuicConn struct {
	Session    quic.Session
	Signal     quic.Stream
	Stream     []io.ReadWriter
	StreamType string
	Name       string
	Heartbeat  chan byte
	IsClosed   bool
	Ready      bool
}

// SendSignal sends the signal to clients.
func (c *QuicConn) SendSignal(b []byte) {
	c.Signal.Write(b)
}

// Init the QUIC connection.
func (c *QuicConn) Init(conf *WorkflowConfig) {
	isInit := true
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := c.Signal.Read(buf)

			if err != nil {
				break
			}
			value := buf[:n]

			if isInit {
				c.Name = string(value)
				c.StreamType = StreamTypeSource
				for _, app := range conf.Flows {
					if app.Name == c.Name {
						c.StreamType = StreamTypeFlow
						break
					}
				}
				for _, app := range conf.Sinks {
					if app.Name == c.Name {
						c.StreamType = StreamTypeSink
						break
					}
				}
				fmt.Println("Receive App:", c.Name, c.StreamType)
				isInit = false
				c.SendSignal(client.SignalAccepted)
				c.Beat()
				continue
			}

			if bytes.Equal(value, client.SignalHeartbeat) {
				c.Heartbeat <- value[0]
			}
		}
	}()
}

// Beat sends the heartbeat to clients and checks if receiving the heartbeat back.
func (c *QuicConn) Beat() {
	go func() {
		defer c.Close()
		for {
			select {
			case _, ok := <-c.Heartbeat:
				if !ok {
					return
				}

			case <-time.After(time.Second):
				// close the connection if didn't receive the heartbeat after 1s.
				c.Close()
			}
		}
	}()

	go func() {
		for {
			// send heartbeat in every 200ms.
			time.Sleep(200 * time.Millisecond)
			c.SendSignal(client.SignalHeartbeat)
		}
	}()
}

func (c *QuicConn) Close() {
	c.Session.CloseWithError(0, "")
	c.IsClosed = true
	c.Ready = true
}

// Start QUIC service.
func Start(endpoint string, handle quic.ServerHandler) error {
	server := quic.NewServer(handle)

	return server.ListenAndServe(context.Background(), endpoint)
}

// Build the workflow by config (.yaml).
func Build(wfConf *WorkflowConfig, connMap *map[int64]*QuicConn, index int) ([]func() (io.ReadWriter, func()), []func() (io.Writer, func())) {
	//init workflow
	if GlobalApp == "" {
		for i, v := range wfConf.Sinks {
			if i == 0 {
				GlobalApp = v.Name
			}
		}

		for i, v := range wfConf.Flows {
			if i == 0 {
				GlobalApp = v.Name
			}
		}
	}

	flows := make([]func() (io.ReadWriter, func()), 0)
	sinks := make([]func() (io.Writer, func()), 0)

	for _, app := range wfConf.Flows {
		flows = append(flows, createReadWriter(app, connMap, index))
	}

	for _, app := range wfConf.Sinks {
		sinks = append(sinks, createWriter(app, connMap, index))
	}

	return flows, sinks

}

func createReadWriter(app App, connMap *map[int64]*QuicConn, index int) func() (io.ReadWriter, func()) {
	fmt.Println("flow s.index.:", index)
	f := func() (io.ReadWriter, func()) {
		if app.Name != GlobalApp {
			index = 0
		}

		var conn *QuicConn = nil
		var id int64 = 0

		for i, c := range *connMap {
			if c.Name == app.Name {
				conn = c
				id = i
			}
		}
		if conn == nil {
			return nil, func() {}
		} else if len(conn.Stream) > index && conn.Stream[index] != nil {
			conn.Ready = true
			return conn.Stream[index], cancelStream(app, conn, connMap, id)
		} else {
			if conn.Ready {
				conn.Ready = false
				conn.SendSignal(client.SignalFlowSink)
			}
			return nil, func() {}
		}

	}

	return f
}

func createWriter(app App, connMap *map[int64]*QuicConn, index int) func() (io.Writer, func()) {
	fmt.Println("sink s.index.:", index)
	f := func() (io.Writer, func()) {
		// if app.Name != GlobalApp {
		// 	index = 0
		// }

		var conn *QuicConn = nil
		var id int64 = 0

		for i, c := range *connMap {
			if c.Name == app.Name {
				conn = c
				id = i
			}
		}

		if conn == nil {
			return nil, func() {}
		} else if len(conn.Stream) > index && conn.Stream[index] != nil {
			conn.Ready = true
			return conn.Stream[index], cancelStream(app, conn, connMap, id)
		} else {
			if conn.Ready {
				conn.Ready = false
				conn.SendSignal(client.SignalFlowSink)
			}
			return nil, func() {}
		}

	}
	return f
}

func cancelStream(app App, conn *QuicConn, connMap *map[int64]*QuicConn, id int64) func() {
	f := func() {
		conn.Close()
		delete(*connMap, id)
	}
	return f
}

type QuicHandler struct {
	serverlessConfig *WorkflowConfig
	connMap          map[int64]*QuicConn
	build            chan quic.Stream
	index            int
	mutex            sync.RWMutex
}

func NewQuicHandler(conf *WorkflowConfig) *QuicHandler {
	quicHandler := QuicHandler{
		serverlessConfig: conf,
		connMap:          map[int64]*QuicConn{},
		build:            make(chan quic.Stream),
		index:            0,
	}
	return &quicHandler
}

func (s *QuicHandler) Listen() error {
	go func() {
		for {
			select {
			case item, ok := <-s.build:
				if !ok {
					return
				}

				flows, sinks := Build(s.serverlessConfig, &s.connMap, s.index)
				stream := DispatcherWithFunc(flows, item)

				go func() {
					for customer := range stream.Observe(rxgo.WithErrorStrategy(rxgo.ContinueOnError)) {
						if customer.Error() {
							fmt.Println(customer.E.Error())
							continue
						}

						value := customer.V.([]byte)

						for _, sink := range sinks {
							go func(_sink func() (io.Writer, func()), buf []byte) {
								writer, cancel := _sink()

								if writer != nil {
									_, err := writer.Write(buf)
									if err != nil {
										cancel()
									}
								}
							}(sink, value)
						}
					}
				}()
				s.index++

			}
		}
	}()
	return nil
}

func (s *QuicHandler) Read(id int64, sess quic.Session, st quic.Stream) error {
	s.mutex.Lock()

	if conn, ok := s.connMap[id]; ok {
		if conn.StreamType == StreamTypeSource {
			conn.Stream = append(conn.Stream, st)
			s.build <- st
		} else {
			conn.Stream = append(conn.Stream, st)
		}
	} else {
		conn := &QuicConn{
			Session:    sess,
			Signal:     st,
			Stream:     make([]io.ReadWriter, 0),
			StreamType: "",
			Name:       "",
			Heartbeat:  make(chan byte),
			IsClosed:   false,
			Ready:      true,
		}
		conn.Init(s.serverlessConfig)
		s.connMap[id] = conn
	}
	s.mutex.Unlock()
	return nil
}
