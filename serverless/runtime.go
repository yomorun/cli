package serverless

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/reactivex/rxgo/v2"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/yomo/pkg/client"
	"github.com/yomorun/yomo/pkg/quic"
	"github.com/yomorun/yomo/pkg/yomo"
)

const (
	StreamTypeSource       string = "source"
	StreamTypeFlow         string = "flow"
	StreamTypeSink         string = "sink"
	StreamTypeZipperSender string = "zipper-sender"
)

// QuicConn represents the QUIC connection.
type QuicConn struct {
	Session    quic.Session
	Signal     quic.Stream
	Stream     io.ReadWriter
	StreamType string
	Name       string
	Heartbeat  chan byte
	IsClosed   bool
	Ready      bool
}

// zipperServerConf represents the config of zipper servers
type zipperServerConf struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// SendSignal sends the signal to clients.
func (c *QuicConn) SendSignal(b []byte) error {
	_, err := c.Signal.Write(b)
	return err
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
				// get negotiation payload
				var payload client.NegotiationPayload
				err := json.Unmarshal(value, &payload)
				if err != nil {
					log.FailureStatusEvent(os.Stdout, "Zipper inits the connection failed: %s", err.Error())
					return
				}

				streamType, err := c.getStreamType(payload, conf)
				if err != nil {
					log.FailureStatusEvent(os.Stdout, "Zipper get the stream type from the connection failed: %s", err.Error())
					return
				}

				c.Name = payload.AppName
				c.StreamType = streamType
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

func (c *QuicConn) getStreamType(payload client.NegotiationPayload, conf *WorkflowConfig) (string, error) {
	switch payload.ClientType {
	case client.ClientTypeSource:
		return StreamTypeSource, nil
	case client.ClientTypeZipperSender:
		return StreamTypeZipperSender, nil
	case client.ClientTypeServerless:
		// check if the app name is in flows
		for _, app := range conf.Flows {
			if app.Name == payload.AppName {
				return StreamTypeFlow, nil
			}
		}
		// check if the app name is in sinks
		for _, app := range conf.Sinks {
			if app.Name == payload.AppName {
				return StreamTypeSink, nil
			}
		}
	}
	return "", fmt.Errorf("the client type %s isn't matched any stream type", payload.ClientType)
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
			err := c.SendSignal(client.SignalHeartbeat)
			if err != nil {
				break
			}
		}
	}()
}

// Close the QUIC connections.
func (c *QuicConn) Close() {
	c.Session.CloseWithError(0, "")
	c.IsClosed = true
	c.Ready = true
}

// Start QUIC service.
func Start(endpoint string, handler quic.ServerHandler) error {
	server := quic.NewServer(handler)

	return server.ListenAndServe(context.Background(), endpoint)
}

// Build the workflow by config (.yaml).
// It will create one stream for each flows/sinks.
func Build(wfConf *WorkflowConfig, connMap *map[int64]*QuicConn) ([]yomo.FlowFunc, []yomo.SinkFunc) {
	//init workflow
	flows := make([]yomo.FlowFunc, 0)
	sinks := make([]yomo.SinkFunc, 0)

	for _, app := range wfConf.Flows {
		flows = append(flows, createReadWriter(app, connMap, StreamTypeFlow))
	}

	for _, app := range wfConf.Sinks {
		sinks = append(sinks, createWriter(app, connMap, StreamTypeSink))
	}

	return flows, sinks
}

// GetSinks get sinks from config and connMap
func GetSinks(wfConf *WorkflowConfig, connMap *map[int64]*QuicConn) []yomo.SinkFunc {
	sinks := make([]yomo.SinkFunc, 0)

	for _, app := range wfConf.Sinks {
		sinks = append(sinks, createWriter(app, connMap, StreamTypeSink))
	}

	return sinks
}

func createReadWriter(app App, connMap *map[int64]*QuicConn, streamType string) yomo.FlowFunc {
	f := func() (io.ReadWriter, yomo.CancelFunc) {
		var conn *QuicConn = nil
		var id int64 = 0

		for i, c := range *connMap {
			if c.StreamType == streamType && c.Name == app.Name {
				conn = c
				id = i
			}
		}
		if conn == nil {
			return nil, func() {}
		} else if conn.Stream != nil {
			conn.Ready = true
			return conn.Stream, cancelStream(app, conn, connMap, id)
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

func createWriter(app App, connMap *map[int64]*QuicConn, streamType string) yomo.SinkFunc {
	f := func() (io.Writer, yomo.CancelFunc) {
		var conn *QuicConn = nil
		var id int64 = 0

		for i, c := range *connMap {
			if c.StreamType == streamType && c.Name == app.Name {
				conn = c
				id = i
			}
		}

		if conn == nil {
			return nil, func() {}
		} else if conn.Stream != nil {
			conn.Ready = true
			return conn.Stream, cancelStream(app, conn, connMap, id)
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
	meshConfigURL    string
	connMap          map[int64]*QuicConn
	source           chan io.Reader
	zipperSenders    []io.Writer
	zipperReceiver   chan io.Reader
	mutex            sync.RWMutex
}

func NewQuicHandler(conf *WorkflowConfig, meshConfURL string) *QuicHandler {
	quicHandler := QuicHandler{
		serverlessConfig: conf,
		meshConfigURL:    meshConfURL,
		connMap:          map[int64]*QuicConn{},
		source:           make(chan io.Reader),
		zipperSenders:    make([]io.Writer, 0),
		zipperReceiver:   make(chan io.Reader),
	}
	return &quicHandler
}

func (s *QuicHandler) Listen() error {
	go func() {
		s.receiveDataFromSources()
	}()

	go func() {
		s.receiveDataFromZipperSenders()
	}()

	if s.meshConfigURL != "" {
		go func() {
			s.GetZipperSenders()
		}()
	}

	return nil
}

func (s *QuicHandler) receiveDataFromSources() {
	for {
		select {
		case item, ok := <-s.source:
			if !ok {
				return
			}

			// one stream for each flows/sinks.
			flows, sinks := Build(s.serverlessConfig, &s.connMap)
			stream := DispatcherWithFunc(flows, item)

			go func() {
				for customer := range stream.Observe(rxgo.WithErrorStrategy(rxgo.ContinueOnError)) {
					if customer.Error() {
						fmt.Println(customer.E.Error())
						continue
					}

					value := customer.V.([]byte)

					// sinks
					for _, sink := range sinks {
						go func(sf yomo.SinkFunc, buf []byte) {
							writer, cancel := sf()

							if writer != nil {
								_, err := writer.Write(buf)
								if err != nil {
									cancel()
								}
							}
						}(sink, value)
					}

					// Zipper-Senders
					for i, sender := range s.zipperSenders {
						if sender == nil {
							continue
						}

						go func(w io.Writer, buf []byte, index int) {
							// send data to donwstream zippers
							_, err := w.Write(value)
							if err != nil {
								log.FailureStatusEvent(os.Stdout, err.Error())
								// remove writer
								s.zipperSenders = append(s.zipperSenders[:index], s.zipperSenders[index+1:]...)
							}
						}(sender, value, i)
					}
				}
			}()
		}
	}
}

func (s *QuicHandler) receiveDataFromZipperSenders() {
	for {
		select {
		case receiver, ok := <-s.zipperReceiver:
			if !ok {
				return
			}

			sinks := GetSinks(s.serverlessConfig, &s.connMap)
			if len(sinks) == 0 {
				continue
			}

			go func() {
				for {
					buf := make([]byte, 3*1024)
					n, err := receiver.Read(buf)
					if err != nil {
						break
					} else {
						value := buf[:n]
						// send data to sinks
						for _, sink := range sinks {
							go func(sf yomo.SinkFunc, buf []byte) {
								writer, cancel := sf()

								if writer != nil {
									_, err := writer.Write(buf)
									if err != nil {
										cancel()
									}
								}
							}(sink, value)
						}
					}
				}
			}()
		}
	}
}

func (s *QuicHandler) Read(id int64, sess quic.Session, st quic.Stream) error {
	s.mutex.Lock()

	if conn, ok := s.connMap[id]; ok {
		if conn.StreamType == StreamTypeSource {
			s.source <- st
		} else if conn.StreamType == StreamTypeZipperSender {
			s.zipperReceiver <- st
		} else {
			conn.Stream = st
		}
	} else {
		conn := &QuicConn{
			Session:    sess,
			Signal:     st,
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

// GetZipperSenders connects to downstream zippers and get Zipper-Senders.
func (s *QuicHandler) GetZipperSenders() error {
	log.InfoStatusEvent(os.Stdout, "Connecting to downstream zippers...")

	// download mesh conf
	res, err := http.Get(s.meshConfigURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var configs []zipperServerConf
	err = decoder.Decode(&configs)
	if err != nil {
		return err
	}

	if len(configs) == 0 {
		return nil
	}

	for _, conf := range configs {
		if conf.Host == s.serverlessConfig.Host && conf.Port == s.serverlessConfig.Port {
			// skip current zipper, only need to connect other zippers in edge-mesh.
			continue
		}

		go func(conf zipperServerConf) {
			cli, err := client.NewZipperSender(s.serverlessConfig.Name).
				Connect(conf.Host, conf.Port)
			if err != nil {
				cli.Retry()
			}

			s.mutex.Lock()
			s.zipperSenders = append(s.zipperSenders, cli)
			s.mutex.Unlock()
		}(conf)
	}

	return nil
}
