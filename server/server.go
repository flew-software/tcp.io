package server

import (
	"fmt"
	"net"
	"tcpio/event"
	"tcpio/events"
	"tcpio/utils"
)

// the server struct
type Server struct {
	Config      Config
	Events      map[string]ConnectionHandler
	connections []Socket
	doListen    bool
}

var emptySocket = Socket{}

func (s *Server) On(eventName string, cb ConnectionHandler) {

	s.Events[eventName] = cb
}

func (s *Server) Emit(eventName string, data []byte) {
	for _, val := range s.connections {
		val.connection.Write(append(append([]byte(eventName), []byte("\n")...), data...))
	}
}

func (s *Server) Listen() error {
	s.doListen = true
	l, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		return err
	}
	for s.doListen {
		c, err := l.Accept()
		if err == nil {
			go func() {

				_, socket := s.newConnection(c)

				defer s.UserDisconnect(socket)
				for {
					err, eventName, message := utils.ReadData(c)
					if err != nil {
						fmt.Println(err)
						break
					}
					socket.Events[eventName](message)
				}
			}()
		}
	}
	return nil
}

func (s *Server) newConnection(c net.Conn) (int, Socket) {
	sLocation := len(s.connections)
	socket := Socket{c, map[string]event.Handler{}, utils.RandomID()}
	s.connections = append(s.connections, socket)
	if val, ok := s.Events[events.Connection]; ok {
		val(socket)
	}
	return sLocation, socket
}

func (s *Server) StopListening() {
	s.doListen = false
}

func Create(config Config) Server {
	return Server{Config: config, Events: make(map[string]ConnectionHandler)}
}

type ConnectionHandler func(Socket)

func (s *Server) UserDisconnect(user Socket) {
	if val, ok := s.Events[events.Disconnect]; ok {
		val(user)
	}
	for i := 0; i < len(s.connections); i++ {
		if s.connections[i].Id == user.Id {
			s.connections[i] = emptySocket
		}

	}
}
