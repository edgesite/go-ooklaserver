package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var payload []byte
var payloadLen = 1000
var download = []byte("DOWNLOAD ")
var downloadLen = len(download)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()
	for {
		mType, m, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var resp []byte
		var length int
		if mType == websocket.BinaryMessage {
			mType = websocket.TextMessage
			length = len(m)
			resp = []byte(fmt.Sprintf("OK %d %d\n", length, time.Now().UnixNano()/1000000))
		} else if mType == websocket.TextMessage {
			msg := strings.Split(strings.Trim(string(m), " "), " ")
			switch {
			case msg[0] == "PING":
				mType = websocket.TextMessage
				resp = []byte(fmt.Sprintf("PONG %d\n", time.Now().UnixNano()/1000000))
			case msg[0] == "DOWNLOAD":
				mType = websocket.BinaryMessage
				if length, err = strconv.Atoi(msg[1]); err != nil {
					log.Println(err)
					break
				}
				if length > 1024*1024 {
					break
				}
				if payload == nil {
					payload = make([]byte, payloadLen)
					if _, err := rand.Read(payload); err != nil {
						panic(err)
					}
				}
				resp = bytes.Repeat(payload, length/payloadLen)
				if length%payloadLen != 0 {
					resp = append(resp, payload[:length%payloadLen]...)
				}
				copy(resp[:downloadLen], download)
			}
		} else {
			break
		}
		if err := c.WriteMessage(mType, resp); err != nil {
			log.Println(err)
			break
		}
	}
}

func main() {
	log.SetFlags(0)
	http.HandleFunc("/ws", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
