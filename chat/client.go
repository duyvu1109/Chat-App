// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chat

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	pb "gitlab.com/vund5/chatapp/api"
	"gitlab.com/vund5/chatapp/ent"
	"google.golang.org/protobuf/proto"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// database
	db *ent.Client
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	username string
}

// readPump: pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var temp pb.Message
		err = proto.Unmarshal(message, &temp)
		if err != nil {
			log.Println("Get trouble when Unmarshal []byte message")
		}
		if temp.GetChatRequest() != nil {
			log.Println("chat")
			c.serveChat(&temp)
			c.serveChatBD(&temp)
		} else if temp.GetCreateRoomRequest() != nil {
			log.Println("Create")
			c.serveCreate(&temp)
		}
	}
}

func (c *Client) serveChat(temp *pb.Message) {
	// Check room is exist in Hub or not
	for i, room := range c.hub.room {
		// If Room is already exist in Hub
		if room.room_id == temp.GetChatRequest().Room {
			// check Member is exist in Room or not
			var isMember bool = false
			for _, x := range c.hub.room[i].members {
				if x == c {
					isMember = true
					break
				}
			}
			// If member is not in room -> add member into room
			if !isMember {
				c.hub.room[i].members = append(c.hub.room[i].members, c)
			}
			// Convert from protobuf to []byte
			var response = pb.Message{
				Payload: &pb.Message_ChatReply{
					ChatReply: &pb.ChatReply{
						Msg:  temp.GetChatRequest().Msg,
						User: temp.GetChatRequest().User,
						Room: temp.GetChatRequest().Room,
					},
				},
			}
			converted_response, err := proto.Marshal(&response)
			if err != nil {
				log.Println("Can't convert from protobuf to []byte")
			}
			c.hub.room[i].room_broadcast <- converted_response
			c.hub.room[i].doBroadcast(converted_response)
		}
	}
}

func (c *Client) serveChatBD(temp *pb.Message) {
	c.username = temp.GetChatRequest().GetUser()
	// put message's data into database

	u, err := c.db.User.
		Create().
		SetMsg(temp.GetChatRequest().Msg).
		SetRoom(temp.GetChatRequest().Room).
		SetName(temp.GetChatRequest().User).
		Save(context.Background())
	// u, err := c.db.User.
	// 	Create().
	// 	SetTime(time.Now().String()).
	// 	SetFrom(request.User.String()).
	// 	SetPassword(request.User.Password).
	// 	SetRoomID(request.RoomId).
	// 	SetContent(request.Msg).
	// 	Save(context.Background())
	if err != nil {
		fmt.Println("Failed creating use`: %w", err)
	} else {
		fmt.Println("Message was received: ", u)
	}
}

func (c *Client) serveCreate(temp *pb.Message) {
	var newRoomId = temp.GetCreateRoomRequest().GetRoom()
	c.hub.room[newRoomId] = &Room{room_id: newRoomId, room_broadcast: make(chan []byte, 256), members: []*Client{}}
	// Update Active_Rooms in Hub
	c.hub.room[newRoomId].members = append(c.hub.room[newRoomId].members, c)
	// After creating new room -> broadcast message into that room
	var response = pb.Message{
		Payload: &pb.Message_CreateRoomReply{
			CreateRoomReply: &pb.CreateRoomReply{
				Room:    newRoomId,
				Created: true,
			},
		},
	}
	converted_response, err := proto.Marshal(&response)
	if err != nil {
		log.Println("Can't convert from protobuf to []byte")
	}
	c.hub.room[newRoomId].room_broadcast <- converted_response
	c.hub.room[newRoomId].doBroadcast(converted_response)
}

// writePump: pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, db *ent.Client, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, db: db, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	log.Println("Registered client: ", client)
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
