package main

import (
	"net"
	"net/url"
	"fmt"
	"os"
	"bufio"
	"net/http"
	"flag"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	isHost = flag.Bool("host", false, "if this is the host")
	wsUrl = flag.String("url", "localhost:8000", "the websocket url")
)

func read(conn *websocket.Conn, channel chan string) {
	for {
		_, msg, err := conn.ReadMessage()
	
		if err == nil {
			channel <- string(msg)
		} else {
			fmt.Println(err)
		}
	}
}

func write(conn *websocket.Conn, channel chan string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		
		if err == nil {
			channel <- text
		} else {
			fmt.Println(err)
		}
	}
}

func loop(conn *websocket.Conn) {
	readChannel := make(chan string)
	writeChannel := make(chan string)
	go read(conn, readChannel)
	go write(conn, writeChannel)

	for {
		select {
			case msg := <-readChannel:
				fmt.Println(msg)
			case msg := <-writeChannel:
				conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

func upgrade(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err == nil {
		loop(conn)
	} else {
		fmt.Println(err)
	}

}

func main() {
	flag.Parse()

	if *isHost {
		http.HandleFunc("/", upgrade)
		http.ListenAndServe(*wsUrl, nil)
	} else {
		u, err := url.Parse(*wsUrl)
		if err != nil {
			return
		}

		tcp, err := net.Dial("tcp", u.Host)
		if err != nil {
			return
		}

		wsHeaders := http.Header{
			"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame"},
		}

		conn, _, err := websocket.NewClient(tcp, u, wsHeaders, 1024, 1024)
		if err == nil {
			loop(conn)
		} else {
			fmt.Println(err)
		}

	}
}
