package sockets

import (
	"fmt"
	"net/http"
	"sync"
	"time"

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
	logger  *logrus.Logger
	clients map[*websocket.Conn]bool
	mutex   *sync.Mutex
	close   bool
	port    int
}

func New(logger *logrus.Logger, port int) *Server {
	return &Server{
		logger:  logger,
		clients: make(map[*websocket.Conn]bool),
		mutex:   &sync.Mutex{},
		close:   false,
		port:    port,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/ws", s.wsHandler)
	go s.handleMessages()
	s.logger.WithField("port", s.port).Info("server started")
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("failed to upgrade connection")
		return
	}
	defer conn.Close()
	s.mutex.Lock()
	s.clients[conn] = true
	s.mutex.Unlock()
	s.logger.WithField("connectionCount", len(s.clients)).Info("new client connected")

	go func() {
		for {
			logrus.Info("sending ping")
			err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second))
			if err != nil {
				s.logger.WithError(err).Error("could not send ping to client")
				conn.Close()
				break
			}
			s.clients[conn] = false
			time.Sleep(10 * time.Second)
			if !s.clients[conn] {
				s.logger.Error("client did not response to ping")
				conn.Close()
				break
			}
		}
	}()

	for !s.close {
		mt, _, err := conn.ReadMessage()
		if err != nil {
			s.mutex.Lock()
			delete(s.clients, conn)
			s.mutex.Unlock()
			break
		}
		if mt == websocket.PongMessage {
			s.clients[conn] = true
		}
	}
}

func (s *Server) handleMessages() {
	for !s.close {
		message := <-broadcast
		s.mutex.Lock()
		for client := range s.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
		s.mutex.Unlock()
	}
}

func (s *Server) BroadcastMessage(message []byte) {
	s.logger.WithField("message", string(message)).Info("broadcasting message")
	s.mutex.Lock()
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
	s.mutex.Unlock()
}

func (s *Server) Close() {
	s.close = true
	s.mutex.Lock()
	for client := range s.clients {
		client.Close()
		delete(s.clients, client)
	}
	s.mutex.Unlock()
}
