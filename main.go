package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hayden-erickson/radon-server/transform"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if len(r.Header["Origin"]) == 0 {
			return false
		}

		origin := r.Header["Origin"][0]

		if origin == "http://localhost:3000" ||
			strings.Contains(origin, "hayden-erickson.github.io") {
			return true
		}

		return false
	},
}

type UploadSinoRequest struct {
	Total   int     `json:"total"`
	Theta   float64 `json:"theta"`
	SinoRow []uint8 `json:"sino_row"`
}

type UploadSinoResponse map[string]interface{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	sp := transform.SimpleProjection{}

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

		if sp.GetRowsProcessed() == req.Total {
			sp.Reset()
			continue
		}

		sp.BackProject(req.Total, req.Theta, req.SinoRow)

		// wait for all subroutines to finish before sending back the final image
		resp, err := json.Marshal(UploadSinoResponse{"image": sp.GetImg()})
		if err != nil {
			break
		}

		if err = c.WriteMessage(mt, resp); err != nil {
			log.Println("write:", err)
			break
		}

		// the processing is finished so we can clear the state and we do not
		// need to send the progress response
	}
}

func main() {
	addr := fmt.Sprintf("localhost:%s", os.Getenv("PORT"))

	http.HandleFunc("/echo", echo)
	fmt.Printf("Listening on %s\n", addr)

	log.Fatal(http.ListenAndServe(addr, nil))

}
