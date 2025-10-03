package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

const (
	natsURL      = "nats://nats:4222"
	chatSubject  = "chat.global"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	natsConn   *nats.Conn
}

func newHub(nc *nats.Conn) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		natsConn:   nc,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("Новый клиент зарегистрирован в хабе")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Клиент отключен от хаба")
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		err = c.hub.natsConn.Publish(chatSubject, message)
		if err != nil {
			log.Printf("Ошибка публикации в NATS: %v", err)
		} else {
			log.Printf("Сообщение опубликовано в NATS: %s", message)
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		if _, err := w.Write(message); err != nil {
			return
		}

		if err := w.Close(); err != nil {
			return
		}
	}

	_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}


func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	log.Println("Клиент успешно подключен по WebSocket")

	go client.writePump()
	go client.readPump()
}

func main() {
	nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	if err != nil {
		log.Fatalf("Не удалось подключиться к NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Успешно подключено к NATS на", natsURL)

	hub := newHub(nc)
	go hub.run()

	_, err = nc.Subscribe(chatSubject, func(msg *nats.Msg) {
		log.Printf("Получено сообщение из NATS: %s", msg.Data)
		hub.broadcast <- msg.Data
	})
	if err != nil {
		log.Fatalf("Не удалось подписаться на тему NATS: %v", err)
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("HTTP-сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}