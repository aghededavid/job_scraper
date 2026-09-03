package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/remotescope/backend/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict to allowed origins in production
	},
}

// ──────────────────────────────────────────
// Message Types
// ──────────────────────────────────────────

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	MsgNewJob    = "new_job"
	MsgJobUpdate = "job_update"
	MsgJobExpire = "job_expire"
	MsgPing      = "ping"
	MsgPong      = "pong"
)

// ──────────────────────────────────────────
// Client
// ──────────────────────────────────────────

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			ping, _ := json.Marshal(Message{Type: MsgPing})
			if err := c.conn.WriteMessage(websocket.TextMessage, ping); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var m Message
		if json.Unmarshal(msg, &m) == nil && m.Type == MsgPong {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

// ──────────────────────────────────────────
// Hub
// ──────────────────────────────────────────

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast a new job event to all connected clients
func (h *Hub) BroadcastNewJob(job interface{}) {
	msg, err := json.Marshal(Message{Type: MsgNewJob, Payload: job})
	if err != nil {
		return
	}
	select {
	case h.broadcast <- msg:
	default:
	}
}

func (h *Hub) ConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ──────────────────────────────────────────
// HTTP Upgrade
// ──────────────────────────────────────────

func ServeWS(hub *Hub, jwtSvc *auth.JWTService, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	userID := ""
	// Optional JWT auth via query param
	if token := r.URL.Query().Get("token"); token != "" {
		if claims, err := jwtSvc.ValidateAccessToken(token); err == nil {
			userID = claims.UserID.String()
		}
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}
