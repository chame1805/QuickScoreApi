package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message es el formato de todos los mensajes enviados por WebSocket
type Message struct {
	Event    string      `json:"event"` // "score_update", "participant_connected", etc.
	RoomCode string      `json:"room"`
	Payload  interface{} `json:"payload"`
}

// ClientInfo contiene los datos públicos de un cliente conectado
type ClientInfo struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

// Client representa una conexión WebSocket activa con identidad conocida
type Client struct {
	conn     *websocket.Conn
	roomCode string
	send     chan []byte
	Info     ClientInfo // quién es este cliente
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

// Register agrega un cliente identificado a una sala y notifica a todos
func (h *Hub) Register(conn *websocket.Conn, roomCode string, info ClientInfo) *Client {
	client := &Client{
		conn:     conn,
		roomCode: roomCode,
		send:     make(chan []byte, 64),
		Info:     info,
	}

	h.mu.Lock()
	if h.rooms[roomCode] == nil {
		h.rooms[roomCode] = make(map[*Client]bool)
	}
	h.rooms[roomCode][client] = true
	h.mu.Unlock()

	// Notificar a todos en la sala que este usuario se conectó
	h.Broadcast(roomCode, Message{
		Event:    "participant_connected",
		RoomCode: roomCode,
		Payload:  info,
	})

	// Enviarle al recién conectado la lista de quiénes ya están en la sala
	h.sendOnlineList(client)

	go client.writePump()
	go client.readPump(h)

	return client
}

// sendOnlineList envía al cliente recién conectado la lista de presentes
func (h *Hub) sendOnlineList(target *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	online := make([]ClientInfo, 0)
	for c := range h.rooms[target.roomCode] {
		if c != target { // excluirse a sí mismo, él ya sabe que está
			online = append(online, c.Info)
		}
	}

	msg := Message{
		Event:    "online_list",
		RoomCode: target.roomCode,
		Payload:  online,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case target.send <- data:
	default:
	}
}

// GetOnlineUsers devuelve los usuarios conectados actualmente en una sala
func (h *Hub) GetOnlineUsers(roomCode string) []ClientInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	online := make([]ClientInfo, 0)
	for c := range h.rooms[roomCode] {
		online = append(online, c.Info)
	}
	return online
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
			close(client.send)
			delete(h.rooms[roomCode], client)
		}
	}
}

// unregister elimina un cliente de su sala y notifica la desconexión
func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	if clients, ok := h.rooms[client.roomCode]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
		}
		if len(clients) == 0 {
			delete(h.rooms, client.roomCode)
		}
	}
	h.mu.Unlock()

	// Notificar a los demás que este usuario se desconectó
	h.Broadcast(client.roomCode, Message{
		Event:    "participant_disconnected",
		RoomCode: client.roomCode,
		Payload:  client.Info,
	})
}

// writePump envía mensajes pendientes al cliente
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump lee del cliente para detectar desconexiones
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
