package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message es el formato de todos los mensajes que se envían por WebSocket
type Message struct {
	Event    string      `json:"event"` // ej: "score_update", "session_started"
	RoomCode string      `json:"room"`
	Payload  interface{} `json:"payload"`
}

// Client representa una conexión WebSocket activa
type Client struct {
	conn     *websocket.Conn
	roomCode string
	send     chan []byte
}

// Hub gestiona todas las conexiones activas agrupadas por sala
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]bool // roomCode → set de clientes
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
	}
}

// Register agrega un cliente a una sala
func (h *Hub) Register(conn *websocket.Conn, roomCode string) *Client {
	client := &Client{
		conn:     conn,
		roomCode: roomCode,
		send:     make(chan []byte, 64),
	}

	h.mu.Lock()
	if h.rooms[roomCode] == nil {
		h.rooms[roomCode] = make(map[*Client]bool)
	}
	h.rooms[roomCode][client] = true
	h.mu.Unlock()

	// Goroutine que escribe mensajes al cliente
	go client.writePump()
	// Goroutine que lee mensajes del cliente (mantiene la conexión viva)
	go client.readPump(h)

	return client
}

// Broadcast envía un mensaje a todos los clientes de una sala
func (h *Hub) Broadcast(roomCode string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("error al serializar mensaje ws:", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.rooms[roomCode] {
		select {
		case client.send <- data:
		default:
			// Si el canal está lleno el cliente está lento, se desconecta
			close(client.send)
			delete(h.rooms[roomCode], client)
		}
	}
}

// unregister elimina un cliente de su sala
func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[client.roomCode]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
		}
		if len(clients) == 0 {
			delete(h.rooms, client.roomCode)
		}
	}
}

// writePump envía mensajes pendientes al cliente por WebSocket
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump lee mensajes del cliente (necesario para detectar desconexiones)
func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister(c)
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
