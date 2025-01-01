package room

import (
	"github.com/gorilla/websocket"
)

type Room struct {
	id         string
	playersNum int
	Players    map[string]*websocket.Conn
}

func (r *Room) NewRoom(id string, playersNum int) {

}

func (r *Room) AddConn(username string, conn *websocket.Conn) bool {
	if r.Players[username] == nil {
		r.Players[username] = conn
		return false
	} else {
		return true
	}
}

func (r *Room) GetLen() int {
	return len(r.Players)
}
