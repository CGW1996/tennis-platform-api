package services

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketMessage WebSocket 訊息結構
type WebSocketMessage struct {
	Type       string      `json:"type"`
	ChatRoomID string      `json:"chatRoomId,omitempty"`
	Data       interface{} `json:"data"`
	Timestamp  time.Time   `json:"timestamp"`
}

// ChatMessageData 聊天訊息數據
type ChatMessageData struct {
	ID          string    `json:"id"`
	ChatRoomID  string    `json:"chatRoomId"`
	SenderID    string    `json:"senderId"`
	Content     string    `json:"content"`
	MessageType string    `json:"messageType"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Client WebSocket 客戶端
type Client struct {
	ID      string
	UserID  string
	Conn    *websocket.Conn
	Send    chan WebSocketMessage
	Hub     *Hub
	RoomIDs map[string]bool // 用戶加入的聊天室
	mu      sync.RWMutex
}

// Hub WebSocket 連接管理中心
type Hub struct {
	clients     map[string]*Client            // 客戶端連接 (key: clientID)
	userClients map[string]*Client            // 用戶連接映射 (key: userID)
	rooms       map[string]map[string]*Client // 聊天室連接 (key: roomID, value: clients)
	register    chan *Client
	unregister  chan *Client
	broadcast   chan WebSocketMessage
	mu          sync.RWMutex
}

// WebSocketService WebSocket 服務
type WebSocketService struct {
	hub      *Hub
	upgrader websocket.Upgrader
}

// NewWebSocketService 創建新的 WebSocket 服務
func NewWebSocketService() *WebSocketService {
	hub := &Hub{
		clients:     make(map[string]*Client),
		userClients: make(map[string]*Client),
		rooms:       make(map[string]map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan WebSocketMessage),
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 在生產環境中應該檢查來源
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	service := &WebSocketService{
		hub:      hub,
		upgrader: upgrader,
	}

	// 啟動 Hub
	go hub.run()

	return service
}

// HandleWebSocket 處理 WebSocket 連接
func (ws *WebSocketService) HandleWebSocket(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:      generateClientID(),
		UserID:  userID,
		Conn:    conn,
		Send:    make(chan WebSocketMessage, 256),
		Hub:     ws.hub,
		RoomIDs: make(map[string]bool),
	}

	ws.hub.register <- client

	// 啟動客戶端的讀寫協程
	go client.writePump()
	go client.readPump()
}

// BroadcastToRoom 向聊天室廣播訊息
func (ws *WebSocketService) BroadcastToRoom(roomID string, message WebSocketMessage) {
	message.ChatRoomID = roomID
	message.Timestamp = time.Now()

	ws.hub.mu.RLock()
	clients, exists := ws.hub.rooms[roomID]
	ws.hub.mu.RUnlock()

	if !exists {
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(clients, client.ID)
		}
	}
}

// SendToUser 向特定用戶發送訊息
func (ws *WebSocketService) SendToUser(userID string, message WebSocketMessage) {
	ws.hub.mu.RLock()
	client, exists := ws.hub.userClients[userID]
	ws.hub.mu.RUnlock()

	if !exists {
		return
	}

	select {
	case client.Send <- message:
	default:
		close(client.Send)
		delete(ws.hub.userClients, userID)
	}
}

// JoinRoom 加入聊天室
func (ws *WebSocketService) JoinRoom(userID, roomID string) {
	ws.hub.mu.Lock()
	defer ws.hub.mu.Unlock()

	client, exists := ws.hub.userClients[userID]
	if !exists {
		return
	}

	// 初始化聊天室
	if ws.hub.rooms[roomID] == nil {
		ws.hub.rooms[roomID] = make(map[string]*Client)
	}

	// 加入聊天室
	ws.hub.rooms[roomID][client.ID] = client
	client.mu.Lock()
	client.RoomIDs[roomID] = true
	client.mu.Unlock()
}

// LeaveRoom 離開聊天室
func (ws *WebSocketService) LeaveRoom(userID, roomID string) {
	ws.hub.mu.Lock()
	defer ws.hub.mu.Unlock()

	client, exists := ws.hub.userClients[userID]
	if !exists {
		return
	}

	// 從聊天室移除
	if clients, roomExists := ws.hub.rooms[roomID]; roomExists {
		delete(clients, client.ID)
		if len(clients) == 0 {
			delete(ws.hub.rooms, roomID)
		}
	}

	client.mu.Lock()
	delete(client.RoomIDs, roomID)
	client.mu.Unlock()
}

// GetOnlineUsers 獲取在線用戶列表
func (ws *WebSocketService) GetOnlineUsers() []string {
	ws.hub.mu.RLock()
	defer ws.hub.mu.RUnlock()

	users := make([]string, 0, len(ws.hub.userClients))
	for userID := range ws.hub.userClients {
		users = append(users, userID)
	}
	return users
}

// GetRoomUsers 獲取聊天室用戶列表
func (ws *WebSocketService) GetRoomUsers(roomID string) []string {
	ws.hub.mu.RLock()
	defer ws.hub.mu.RUnlock()

	clients, exists := ws.hub.rooms[roomID]
	if !exists {
		return []string{}
	}

	users := make([]string, 0, len(clients))
	for _, client := range clients {
		users = append(users, client.UserID)
	}
	return users
}

// Hub 運行方法
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.userClients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("Client connected: %s (User: %s)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				delete(h.userClients, client.UserID)
				close(client.Send)

				// 從所有聊天室移除
				client.mu.RLock()
				for roomID := range client.RoomIDs {
					if clients, exists := h.rooms[roomID]; exists {
						delete(clients, client.ID)
						if len(clients) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
				client.mu.RUnlock()
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s (User: %s)", client.ID, client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
					delete(h.userClients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Client 讀取協程
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// 處理客戶端訊息
		c.handleMessage(message)
	}
}

// Client 寫入協程
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 處理客戶端訊息
func (c *Client) handleMessage(message WebSocketMessage) {
	switch message.Type {
	case "join_room":
		if roomID, ok := message.Data.(string); ok {
			c.Hub.mu.Lock()
			if c.Hub.rooms[roomID] == nil {
				c.Hub.rooms[roomID] = make(map[string]*Client)
			}
			c.Hub.rooms[roomID][c.ID] = c
			c.mu.Lock()
			c.RoomIDs[roomID] = true
			c.mu.Unlock()
			c.Hub.mu.Unlock()
		}

	case "leave_room":
		if roomID, ok := message.Data.(string); ok {
			c.Hub.mu.Lock()
			if clients, exists := c.Hub.rooms[roomID]; exists {
				delete(clients, c.ID)
				if len(clients) == 0 {
					delete(c.Hub.rooms, roomID)
				}
			}
			c.mu.Lock()
			delete(c.RoomIDs, roomID)
			c.mu.Unlock()
			c.Hub.mu.Unlock()
		}

	case "ping":
		response := WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		select {
		case c.Send <- response:
		default:
			close(c.Send)
		}
	}
}

// generateClientID 生成客戶端 ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成隨機字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
