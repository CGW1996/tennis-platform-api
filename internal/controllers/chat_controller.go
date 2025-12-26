package controllers

import (
	"net/http"
	"strconv"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"
	"time"

	"github.com/gin-gonic/gin"
)

// ChatController 聊天控制器
type ChatController struct {
	chatUsecase      *usecases.ChatUsecase
	websocketService *services.WebSocketService
}

// NewChatController 創建新的聊天控制器
func NewChatController(chatUsecase *usecases.ChatUsecase, websocketService *services.WebSocketService) *ChatController {
	return &ChatController{
		chatUsecase:      chatUsecase,
		websocketService: websocketService,
	}
}

// HandleWebSocket 處理 WebSocket 連接
// @Summary WebSocket 連接
// @Description 建立 WebSocket 連接用於即時聊天
// @Tags chat
// @Security BearerAuth
// @Router /api/v1/chat/ws [get]
func (cc *ChatController) HandleWebSocket(c *gin.Context) {
	cc.websocketService.HandleWebSocket(c)
}

// CreateChatRoom 創建聊天室
// @Summary 創建聊天室
// @Description 創建新的聊天室
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatRoomRequest true "創建聊天室請求"
// @Success 201 {object} dto.ChatRoomResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms [post]
func (cc *ChatController) CreateChatRoom(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	var req dto.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請求格式錯誤", "details": err.Error()})
		return
	}

	chatRoom, err := cc.chatUsecase.CreateChatRoom(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建聊天室失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chatRoom)
}

// GetChatRooms 獲取聊天室列表
// @Summary 獲取聊天室列表
// @Description 獲取用戶參與的所有聊天室
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.ChatRoomResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms [get]
func (cc *ChatController) GetChatRooms(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	chatRooms, err := cc.chatUsecase.GetChatRooms(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取聊天室列表失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chatRooms)
}

// GetChatRoom 獲取聊天室詳情
// @Summary 獲取聊天室詳情
// @Description 獲取指定聊天室的詳細信息
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Success 200 {object} dto.ChatRoomResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId} [get]
func (cc *ChatController) GetChatRoom(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	chatRoom, err := cc.chatUsecase.GetChatRoom(roomID, userID)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" || err.Error() == "聊天室不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取聊天室失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chatRoom)
}

// SendMessage 發送訊息
// @Summary 發送訊息
// @Description 向聊天室發送訊息
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SendMessageRequest true "發送訊息請求"
// @Success 201 {object} dto.ChatMessageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/messages [post]
func (cc *ChatController) SendMessage(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請求格式錯誤", "details": err.Error()})
		return
	}

	message, err := cc.chatUsecase.SendMessage(userID, &req)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "發送訊息失敗", "details": err.Error()})
		return
	}

	// 通過 WebSocket 廣播訊息
	wsMessage := services.WebSocketMessage{
		Type: "new_message",
		Data: services.ChatMessageData{
			ID:          message.ID,
			ChatRoomID:  message.ChatRoomID,
			SenderID:    message.SenderID,
			Content:     message.Content,
			MessageType: message.MessageType,
			CreatedAt:   message.CreatedAt,
		},
		Timestamp: time.Now(),
	}
	cc.websocketService.BroadcastToRoom(req.ChatRoomID, wsMessage)

	c.JSON(http.StatusCreated, message)
}

// GetMessages 獲取聊天室訊息
// @Summary 獲取聊天室訊息
// @Description 獲取指定聊天室的訊息列表
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(50)
// @Param before query string false "獲取此時間之前的訊息 (RFC3339格式)"
// @Success 200 {array} dto.ChatMessageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId}/messages [get]
func (cc *ChatController) GetMessages(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	// 解析查詢參數
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	var beforeTime *time.Time
	if beforeStr := c.Query("before"); beforeStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, beforeStr); err == nil {
			beforeTime = &parsedTime
		}
	}

	req := dto.GetMessagesRequest{
		ChatRoomID: roomID,
		Page:       page,
		Limit:      limit,
		Before:     beforeTime,
	}

	messages, err := cc.chatUsecase.GetMessages(userID, &req)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取訊息失敗", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// MarkMessagesAsRead 標記訊息為已讀
// @Summary 標記訊息為已讀
// @Description 標記聊天室中的訊息為已讀
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId}/read [post]
func (cc *ChatController) MarkMessagesAsRead(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	err := cc.chatUsecase.MarkMessagesAsRead(userID, roomID)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "標記已讀失敗", "details": err.Error()})
		return
	}

	// 通過 WebSocket 通知其他用戶
	wsMessage := services.WebSocketMessage{
		Type: "messages_read",
		Data: map[string]interface{}{
			"userId":     userID,
			"chatRoomId": roomID,
			"readAt":     time.Now(),
		},
		Timestamp: time.Now(),
	}
	cc.websocketService.BroadcastToRoom(roomID, wsMessage)

	c.JSON(http.StatusOK, gin.H{"message": "標記已讀成功"})
}

// LeaveChatRoom 離開聊天室
// @Summary 離開聊天室
// @Description 離開指定的聊天室
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId}/leave [post]
func (cc *ChatController) LeaveChatRoom(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	err := cc.chatUsecase.LeaveChatRoom(userID, roomID)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "離開聊天室失敗", "details": err.Error()})
		return
	}

	// 從 WebSocket 聊天室移除用戶
	cc.websocketService.LeaveRoom(userID, roomID)

	// 通知其他用戶
	wsMessage := services.WebSocketMessage{
		Type: "user_left",
		Data: map[string]interface{}{
			"userId":     userID,
			"chatRoomId": roomID,
			"leftAt":     time.Now(),
		},
		Timestamp: time.Now(),
	}
	cc.websocketService.BroadcastToRoom(roomID, wsMessage)

	c.JSON(http.StatusOK, gin.H{"message": "離開聊天室成功"})
}

// JoinChatRoom 加入聊天室
// @Summary 加入聊天室
// @Description 加入指定的聊天室 (WebSocket)
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId}/join [post]
func (cc *ChatController) JoinChatRoom(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	// 檢查用戶是否有權限加入聊天室
	_, err := cc.chatUsecase.GetChatRoom(roomID, userID)
	if err != nil {
		if err.Error() == "您不是此聊天室的參與者" || err.Error() == "聊天室不存在" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "檢查聊天室權限失敗", "details": err.Error()})
		return
	}

	// 加入 WebSocket 聊天室
	cc.websocketService.JoinRoom(userID, roomID)

	// 通知其他用戶
	wsMessage := services.WebSocketMessage{
		Type: "user_joined",
		Data: map[string]interface{}{
			"userId":     userID,
			"chatRoomId": roomID,
			"joinedAt":   time.Now(),
		},
		Timestamp: time.Now(),
	}
	cc.websocketService.BroadcastToRoom(roomID, wsMessage)

	c.JSON(http.StatusOK, gin.H{"message": "加入聊天室成功"})
}

// GetOnlineUsers 獲取在線用戶
// @Summary 獲取在線用戶
// @Description 獲取當前在線的用戶列表
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/chat/online-users [get]
func (cc *ChatController) GetOnlineUsers(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	onlineUsers := cc.websocketService.GetOnlineUsers()
	c.JSON(http.StatusOK, gin.H{
		"onlineUsers": onlineUsers,
		"count":       len(onlineUsers),
	})
}

// GetRoomUsers 獲取聊天室在線用戶
// @Summary 獲取聊天室在線用戶
// @Description 獲取指定聊天室的在線用戶列表
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "聊天室ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/chat/rooms/{roomId}/online-users [get]
func (cc *ChatController) GetRoomUsers(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授權"})
		return
	}

	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "聊天室ID不能為空"})
		return
	}

	roomUsers := cc.websocketService.GetRoomUsers(roomID)
	c.JSON(http.StatusOK, gin.H{
		"roomUsers": roomUsers,
		"count":     len(roomUsers),
	})
}
