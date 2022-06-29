// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"gitlab.com/vund5/chatapp/chat"
	"gitlab.com/vund5/chatapp/ent"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	// db_main_host := os.Getenv("DB.host")
	database, err := ent.Open("mysql", ("root:duyvu1109@tcp(mysqldb:3306)/chatapp?parseTime=True"))
	if err != nil {
		log.Fatalf("failed opening connection to mySQL: %v", err)
	}
	fmt.Print("Connecting to MySQL in port 3306 !\n")
	defer database.Close()
	// Run the auto migration tool.
	if err := database.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	//////////////////////////////

	hub := chat.NewHub()
	go hub.Run()

	fmt.Print("Created Hub successfully !\n")
	// http.HandleFunc("/", serveFrontend)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(hub, database, w, r)
	})

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
