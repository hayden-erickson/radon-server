package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if len(r.Header["Origin"]) == 0 {
			return false
		}

		if r.Header["Origin"][0] == "http://localhost:3000" {
			return true
		}

		return false
	},
}

type UploadSinoRequest struct {
	Total int `json:"total"`
	Theta float32 `json:"theta"`
	SinoRow []uint8 `json:"sino_row"`
}

type UploadSinoResponse map[string]int

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	N := 0

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var req UploadSinoRequest
		if err := json.Unmarshal(message, &req); err != nil {
			break
		}

		N += 1
		resp := UploadSinoResponse{ "rowsProcessed": N }

		bytesResp, err := json.Marshal(resp)
		if err != nil {
			break
		}

		err = c.WriteMessage(mt, bytesResp)

		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

