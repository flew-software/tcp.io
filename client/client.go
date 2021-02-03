package client

import (
	"fmt"
	"net"
	"tcpio/event"
	"tcpio/events"
	"tcpio/utils"
)

type Client struct {
	Config     Config
	Events     map[string]ConnectionHandler
	connection net.Conn
}

type Socket struct {
	connection net.Conn
	Events     map[string]event.Handler
}

type Config struct {
	Addr string
}

func (c *Client) Connect() net.Conn {
	conn, err := net.Dial("tcp", c.Config.Addr)
	if err != nil {
		return conn
	}
	c.connection = conn
	socket := Socket{
		connection: conn,
		Events:     map[string]event.Handler{},
	}
	if val, ok := c.Events[events.Connection]; ok {
		val(socket)
	}

	go func() {
		for {
			err, eventName, message := utils.ReadData(c.connection)
			if err != nil {
				fmt.Println(err)
				break
			}
			socket.Events[eventName](message)
		}
	}()
	return nil
}

type ConnectionHandler func(Socket)

func (c *Client) On(eventName string, cb ConnectionHandler) {
	c.Events[eventName] = cb
}
func (c *Client) Emit(eventName string, data []byte) {
	c.connection.Write(append(append([]byte(eventName), []byte("\n")...), data...))
}

func Create(config Config) Client {
	return Client{Config: config, Events: make(map[string]ConnectionHandler)}
}

func (s *Socket) Emit(eventName string, data []byte) {
	eventNameB := append([]byte(eventName), byte('\n'))
	data = append(data, byte('\n'))
	s.connection.Write(append(eventNameB, data...))
}

func (s *Socket) On(eventName string, cb event.Handler) {
	s.Events[eventName] = cb
}
