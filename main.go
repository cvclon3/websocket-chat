// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
//go:build ignore
//+build ignore
*/

// started on November 12, 2023

// USED:
// https://github.com/gorilla/websocket/blob/main/examples/echo/server.go - main source
// https://medium.com/@parvjn616/building-a-websocket-chat-application-in-go-388fff758575 - as ROOM prototype

package main

import (
	game_main "awesomeProject/game-main"
	game_room "awesomeProject/game-room"
	"awesomeProject/messages"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	//"awesomeProject/game-main"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var clients = make(map[string]*websocket.Conn)

func echo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/echo" {
		game_main.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	conn.WriteMessage(1, []byte("CONNECTION SUCCESFUL"))

	fmt.Println(clients)
	fmt.Println("new conn: ", conn.NetConn())

	defer conn.Close()

	var initMsg messages.InitMessage

	err = conn.ReadJSON(&initMsg)
	if err != nil {
		fmt.Println(err)
		return
	} else {

		if initMsg.Username == "" || len(initMsg.Username) > 10 {
			conn.WriteMessage(1, []byte("NULL OR TOO LONG USERNAME (MAX LEN 10)"))
			return
		}

		if _, ok := clients[initMsg.Username]; ok {
			conn.WriteMessage(1, []byte("USERNAME IS ALREADY EXIST"))
			return
		}

		clients[initMsg.Username] = conn
		defer fmt.Println("END CHAT ", initMsg.Username, clients)
		defer delete(clients, initMsg.Username)

		for {
			var msg messages.Message_
			//mt, message, err := conn.ReadMessage()
			////fmt.Println("mt:", mt)
			//if err != nil {
			//	log.Println("read:", err)
			//	break
			//}

			err = conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println(err)
				conn.WriteMessage(1, []byte("Message_ doenst send"))
				break
			}

			fmt.Println("SENDER: ", msg.Sender, initMsg.Username)
			if msg.Sender != initMsg.Username {
				conn.WriteMessage(1, []byte("SENDER USERNAMES DOESN'T MATCH"))
				continue
			}

			_, ok := clients[msg.Receiver]
			if msg.Receiver == "" || len(msg.Receiver) > 10 || !ok {
				conn.WriteMessage(1, []byte("WRONG RECEIVER USERNAME"))
				continue
			}

			//fmt.Println(msg.Receiver)

			log.Printf("recv: %s", msg.Message, "R", msg.Receiver)
			err = clients[msg.Receiver].WriteMessage(1, []byte("From: "+msg.Sender+`: `+msg.Message))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		game_main.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
	//fmt.Println("r.Host: ", r.Host)
	//http.FileServer(http.Dir("./static/main.js"))
}

func room(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/room" {
		game_main.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	roomTemplate.Execute(w, "ws://"+r.Host+"/room/echo")
	//fmt.Println("ROOM", r.Host)
}

var rooms = make(map[string]*game_room.Room)
var broadcast = make(chan messages.RoomMessage)

func roomEcho(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/room/echo" {
		game_main.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer conn.Close()

	conn.WriteMessage(1, []byte("CONNECTION SUCCESFULL"))

	var initMsg messages.InitRoomMessage

	err = conn.ReadJSON(&initMsg)

	fmt.Println(initMsg.Username, initMsg.RoomName)

	if err != nil {
		fmt.Println(err)
		return
	} else {

		if initMsg.RoomName == "" {
			conn.WriteMessage(1, []byte("NULL ROOMNAME"))
			return
		}

		if initMsg.Username == "" || len(initMsg.Username) > 10 {
			conn.WriteMessage(1, []byte("NULL OR TOO LONG USERNAME (MAX LEN 10)"))
			return
		}

		_, ok := rooms[initMsg.RoomName]

		if ok {

			if rooms[initMsg.RoomName].AddConn(initMsg.Username, conn) {
				conn.WriteMessage(1, []byte("USERNAME IN ROOM IS ALREADY EXIST"))
				return
			}
		} else {

			var r game_room.Room
			r.Players = make(map[string]*websocket.Conn)
			r.AddConn(initMsg.Username, conn)
			rooms[initMsg.RoomName] = &r
		}

		fmt.Println("ROOMS", rooms)

		defer func() {
			delete(rooms[initMsg.RoomName].Players, initMsg.Username)
			if rooms[initMsg.RoomName].GetLen() == 0 {
				delete(rooms, initMsg.RoomName)
				fmt.Println("DELETE ROOM")
			}
		}()

		for {
			var msg messages.RoomMessage

			fmt.Println("ROOM_MESS", msg.Message)

			err := conn.ReadJSON(&msg)
			if err != nil {
				fmt.Println(err)
				return
			}

			if msg.Username != initMsg.Username {
				conn.WriteMessage(1, []byte("SENDER USERNAMES DOESN'T MATCH"))
				continue
			}

			_, ok := rooms[msg.RoomName]
			if msg.RoomName == "" || !ok {
				conn.WriteMessage(1, []byte("WRONG ROOMNAME"))
				continue
			}

			broadcast <- msg
		}
	}
}

func handleRoomMessage() {
	for {
		msg := <-broadcast

		for username_, userconn_ := range rooms[msg.RoomName].Players {
			if msg.Username == username_ {
				continue
			}

			log.Printf("recv: %s", msg.Message)
			err := userconn_.WriteMessage(1, []byte("From: "+msg.Username+`: `+msg.Message))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func RoomIsNull() {
	for {
		for roomName, room := range rooms {
			if room.GetLen() == 0 {
				delete(rooms, roomName)
				fmt.Println("DELETE ROOM")
			}
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	http.HandleFunc("/room/echo", roomEcho)
	http.HandleFunc("/room", room)

	go handleRoomMessage()
	//go RoomIsNull()

	log.Fatal(http.ListenAndServe(*addr, nil))

}

var homeTemplate = template.Must(template.ParseFiles("static/home.html"))
var roomTemplate = template.Must(template.ParseFiles("static/room.html"))
