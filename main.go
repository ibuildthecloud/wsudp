package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	addr = flag.Int("port", 27960, "http service address")
	lo   = net.ParseIP("127.0.0.1")
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serve(w http.ResponseWriter, r *http.Request) {
	protocols := websocket.Subprotocols(r)
	header := http.Header{}
	if len(protocols) > 0 {
		log.Print("protocols", protocols)
		header.Set("Sec-Websocket-Protocol", protocols[0])
	}

	c, err := upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	pc, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		log.Println("listen:", err)
		return
	}
	defer pc.Close()

	go func() {
		<-r.Context().Done()
		log.Println("done close")
		pc.Close()
	}()

	go func() {
		buf := make([]byte, 70000, 70000)
		for {
			n, _, err := pc.ReadFrom(buf)
			if err != nil {
				log.Println("read from", err)
				c.Close()
				break
			}

			ms, err := c.NextWriter(websocket.BinaryMessage)
			if err != nil {
				log.Println("ws writer", err)
				c.Close()
				break
			}

			//log.Println("udp->ws write:", string(buf[:n]))
			_, err = ms.Write(buf[:n])
			if err != nil {
				log.Println("write", err)
				c.Close()
				break
			}
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		//log.Println("ws->udp write:", string(message))
		_, err = pc.WriteTo(message, &net.UDPAddr{
			IP:   lo,
			Port: *addr,
		})
		if err != nil {
			log.Println("udp write:", err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", serve)
	listen := fmt.Sprintf("0.0.0.0:%d", *addr)
	log.Print("listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
