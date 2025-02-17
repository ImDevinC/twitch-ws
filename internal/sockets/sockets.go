package sockets

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	broadcast = make(chan []byte)
)

type Server struct {
	logger *logrus.Logger
	hub    *Hub
	mutex  *sync.Mutex
	close  bool
	port   int
}

func New(logger *logrus.Logger, port int) *Server {
	return &Server{
		logger: logger,
		hub:    newHub(),
		mutex:  &sync.Mutex{},
		close:  false,
		port:   port,
	}
}

func (s *Server) Start() {
	go s.hub.run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(s.hub, w, r)
	})
	s.logger.WithField("port", s.port).Info("server started")
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("failed to upgrade connection")
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), logger: s.logger}
	client.hub.register <- client
	go client.writePump()
	go client.readPump()
}

func (s *Server) BroadcastMessage(message []byte) {
	s.logger.WithField("message", string(message)).Info("broadcasting message")
	s.hub.broadcast <- message
}

func (s *Server) Close() {
	s.close = true
}
