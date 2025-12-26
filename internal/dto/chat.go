package dto

import "time"

// ===== 聊天室相關 =====

// CreateChatRoomRequest 創建聊天室請求
type CreateChatRoomRequest struct {
	MatchID        *string  `json:"matchId,omitempty"`
	Type           string   `json:"type"` // match, group, direct
	Name           *string  `json:"name,omitempty"`
	ParticipantIDs []string `json:"participantIds"`
}

// SendMessageRequest 發送訊息請求
type SendMessageRequest struct {
	ChatRoomID  string `json:"chatRoomId" binding:"required"`
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"messageType,omitempty"` // text, image, file
}

// GetMessagesRequest 獲取訊息請求
type GetMessagesRequest struct {
	ChatRoomID string     `json:"chatRoomId" binding:"required"`
	Page       int        `json:"page,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Before     *time.Time `json:"before,omitempty"` // 獲取此時間之前的訊息
}

// ===== 聊天響應相關 =====

// ChatRoomResponse 聊天室響應
type ChatRoomResponse struct {
	ID           string                    `json:"id"`
	MatchID      *string                   `json:"matchId,omitempty"`
	Type         string                    `json:"type"`
	Name         *string                   `json:"name,omitempty"`
	IsActive     bool                      `json:"isActive"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
	Participants []ChatParticipantResponse `json:"participants"`
	LastMessage  *ChatMessageResponse      `json:"lastMessage,omitempty"`
	UnreadCount  int                       `json:"unreadCount"`
}

// ChatMessageResponse 聊天訊息響應
type ChatMessageResponse struct {
	ID          string    `json:"id"`
	ChatRoomID  string    `json:"chatRoomId"`
	SenderID    string    `json:"senderId"`
	Content     string    `json:"content"`
	MessageType string    `json:"messageType"`
	IsRead      bool      `json:"isRead"`
	CreatedAt   time.Time `json:"createdAt"`
	Sender      *UserInfo `json:"sender,omitempty"`
}

// ChatParticipantResponse 聊天室參與者響應
type ChatParticipantResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	JoinedAt   time.Time  `json:"joinedAt"`
	LastReadAt *time.Time `json:"lastReadAt,omitempty"`
	IsActive   bool       `json:"isActive"`
	User       *UserInfo  `json:"user,omitempty"`
}

// UserInfo 用戶基本信息
type UserInfo struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}
