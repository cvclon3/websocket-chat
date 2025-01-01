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
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"sync"
	"wschat.cvclon3.net/internal/messages"
	"wschat.cvclon3.net/internal/room"
	"wschat.cvclon3.net/pkg/web_errors"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var clients_ sync.Map

func WSechoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/echo" {
		web_errors.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	err = conn.WriteMessage(1, []byte("CONNECTION SUCCESFUL"))
	if err != nil {
		log.Print("49 - ON CONN ERROR", err)
		return
	}

	fmt.Println(clients_)
	fmt.Println("new conn: ", conn.NetConn())

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Print("60 - close error:", err)
			return
		}
	}(conn)

	var initMsg messages.InitMessage

	err = conn.ReadJSON(&initMsg)
	if err != nil {
		fmt.Println(err)
		return
	} else {

		if initMsg.Username == "" || len(initMsg.Username) > 10 {
			err = conn.WriteMessage(1, []byte("NULL OR TOO LONG USERNAME (MAX LEN 10)"))
			if err != nil {
				fmt.Println("75 - error")
			}
			return
		}

		if _, ok := clients_.Load(initMsg.Username); ok {
			err = conn.WriteMessage(1, []byte("USERNAME IS ALREADY EXIST"))
			if err != nil {
				fmt.Println("83 - error")
			}
			return
		}

		clients_.Store(initMsg.Username, conn)

		defer fmt.Println("END CHAT ", initMsg.Username, clients_)
		defer clients_.Delete(initMsg.Username)

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
				err = conn.WriteMessage(1, []byte("Message_ doenst send"))
				if err != nil {
					fmt.Println("114 - error", err)
					break
				}
				break
			}

			fmt.Println("SENDER: ", msg.Sender, initMsg.Username)
			if msg.Sender != initMsg.Username {
				err = conn.WriteMessage(1, []byte("SENDER USERNAMES DOESN'T MATCH"))
				if err != nil {
					fmt.Println("123 - error", err)
					break
				}
				continue
			}

			_, ok := clients_.Load(msg.Receiver)
			if msg.Receiver == "" || len(msg.Receiver) > 10 || !ok {
				err = conn.WriteMessage(1, []byte("WRONG RECEIVER USERNAME"))
				if err != nil {
					fmt.Println("133 - error", err)
					break
				}
				continue
			}

			//fmt.Println(msg.Receiver)

			log.Printf("recv: %s", msg.Message, "R", msg.Receiver)

			client_, ok := clients_.Load(msg.Receiver)
			if client__, ok := client_.(*websocket.Conn); ok {
				err = client__.WriteMessage(1, []byte("From: "+msg.Sender+`: `+msg.Message))
				if err != nil {
					fmt.Println("144 - error", err)
					break
				}
			} else {
				log.Printf("ERROR\nITS NOT A WS.CONN")
				return
			}
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		web_errors.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	homeTemplate.Execute(w, "ws://"+r.Host+"/WSechoHandler")
	//fmt.Println("r.Host: ", r.Host)
	//http.FileServer(http.Dir("./web/main.js"))
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/room" {
		web_errors.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	roomTemplate.Execute(w, "ws://"+r.Host+"/room/WSechoHandler")
	//fmt.Println("ROOM", r.Host)
}

var rooms = make(map[string]*room.Room)
var broadcast = make(chan messages.RoomMessage)

func roomEcho(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/room/WSechoHandler" {
		web_errors.ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("204 - error", err)
		}
	}(conn)

	err = conn.WriteMessage(1, []byte("CONNECTION SUCCESFULL"))
	if err != nil {
		fmt.Println("210 - error", err)
	}

	var initMsg messages.InitRoomMessage

	err = conn.ReadJSON(&initMsg)

	fmt.Println(initMsg.Username, initMsg.RoomName)

	if err != nil {
		fmt.Println(err)
		return
	} else {

		if initMsg.RoomName == "" {
			err = conn.WriteMessage(1, []byte("NULL ROOMNAME"))
			if err != nil {
				fmt.Println("227 - error", err)
			}
			return
		}

		if initMsg.Username == "" || len(initMsg.Username) > 10 {
			err = conn.WriteMessage(1, []byte("NULL OR TOO LONG USERNAME (MAX LEN 10)"))
			if err != nil {
				fmt.Println("235 - error", err)
			}
			return
		}

		_, ok := rooms[initMsg.RoomName]

		if ok {

			if rooms[initMsg.RoomName].AddConn(initMsg.Username, conn) {
				err = conn.WriteMessage(1, []byte("USERNAME IN ROOM IS ALREADY EXIST"))
				if err != nil {
					fmt.Println("246 - error", err)
				}
				return
			}
		} else {

			var r room.Room
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
				err = conn.WriteMessage(1, []byte("SENDER USERNAMES DOESN'T MATCH"))
				if err != nil {
					fmt.Println("283 - error", err)
				}
				continue
			}

			_, ok := rooms[msg.RoomName]
			if msg.RoomName == "" || !ok {
				err = conn.WriteMessage(1, []byte("WRONG ROOMNAME"))
				if err != nil {
					fmt.Println("292 - error", err)
				}
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

//func RoomIsNull() {
//	for {
//		for roomName, room := range rooms {
//			if room.GetLen() == 0 {
//				delete(rooms, roomName)
//				fmt.Println("DELETE ROOM")
//			}
//		}
//	}
//}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", WSechoHandler)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/room/echo", roomEcho)
	http.HandleFunc("/room", roomHandler)

	go handleRoomMessage()
	//go RoomIsNull()

	log.Fatal(http.ListenAndServe(*addr, nil))
	fmt.Println("RUNNING")
}

var homeTemplate = template.Must(template.ParseFiles("web/home.html"))
var roomTemplate = template.Must(template.ParseFiles("web/room.html"))
