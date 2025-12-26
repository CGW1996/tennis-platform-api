package usecases

import (
	"errors"
	"fmt"
	"tennis-platform/backend/internal/dto"
	"tennis-platform/backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatUsecase 聊天用例
type ChatUsecase struct {
	db *gorm.DB
}

// NewChatUsecase 創建新的聊天用例
func NewChatUsecase(db *gorm.DB) *ChatUsecase {
	return &ChatUsecase{
		db: db,
	}
}

// UserInfo 用戶基本信息
type UserInfo struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

// CreateChatRoom 創建聊天室
func (uc *ChatUsecase) CreateChatRoom(userID string, req *dto.CreateChatRoomRequest) (*dto.ChatRoomResponse, error) {
	// 驗證請求
	if len(req.ParticipantIDs) == 0 {
		return nil, errors.New("至少需要一個參與者")
	}

	// 檢查是否為直接聊天且已存在
	if req.Type == "direct" && len(req.ParticipantIDs) == 1 {
		existingRoom, err := uc.findDirectChatRoom(userID, req.ParticipantIDs[0])
		if err == nil && existingRoom != nil {
			return uc.getChatRoomResponse(existingRoom, userID)
		}
	}

	// 開始事務
	tx := uc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建聊天室
	chatRoom := &models.ChatRoom{
		ID:       uuid.New().String(),
		MatchID:  req.MatchID,
		Type:     req.Type,
		Name:     req.Name,
		IsActive: true,
	}

	if err := tx.Create(chatRoom).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("創建聊天室失敗: %w", err)
	}

	// 添加創建者為參與者
	allParticipants := append([]string{userID}, req.ParticipantIDs...)
	uniqueParticipants := removeDuplicates(allParticipants)

	for _, participantID := range uniqueParticipants {
		participant := &models.ChatParticipant{
			ID:         uuid.New().String(),
			ChatRoomID: chatRoom.ID,
			UserID:     participantID,
			JoinedAt:   time.Now(),
			IsActive:   true,
		}

		if err := tx.Create(participant).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("添加參與者失敗: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事務失敗: %w", err)
	}

	// 重新查詢完整的聊天室信息
	return uc.GetChatRoom(chatRoom.ID, userID)
}

// SendMessage 發送訊息
func (uc *ChatUsecase) SendMessage(userID string, req *dto.SendMessageRequest) (*dto.ChatMessageResponse, error) {
	// 檢查用戶是否為聊天室參與者
	var participant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		req.ChatRoomID, userID, true).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("您不是此聊天室的參與者")
		}
		return nil, fmt.Errorf("檢查參與者失敗: %w", err)
	}

	// 創建訊息
	message := &models.ChatMessage{
		ID:          uuid.New().String(),
		ChatRoomID:  req.ChatRoomID,
		SenderID:    userID,
		Content:     req.Content,
		MessageType: req.MessageType,
		IsRead:      false,
	}

	if message.MessageType == "" {
		message.MessageType = "text"
	}

	if err := uc.db.Create(message).Error; err != nil {
		return nil, fmt.Errorf("發送訊息失敗: %w", err)
	}

	// 更新聊天室的最後活動時間
	uc.db.Model(&models.ChatRoom{}).Where("id = ?", req.ChatRoomID).Update("updated_at", time.Now())

	// 查詢完整的訊息信息
	return uc.getMessageResponse(message.ID)
}

// GetMessages 獲取聊天室訊息
func (uc *ChatUsecase) GetMessages(userID string, req *dto.GetMessagesRequest) ([]dto.ChatMessageResponse, error) {
	// 檢查用戶是否為聊天室參與者
	var participant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		req.ChatRoomID, userID, true).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("您不是此聊天室的參與者")
		}
		return nil, fmt.Errorf("檢查參與者失敗: %w", err)
	}

	// 設置默認分頁參數
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	// 構建查詢
	query := uc.db.Where("chat_room_id = ?", req.ChatRoomID)

	if req.Before != nil {
		query = query.Where("created_at < ?", *req.Before)
	}

	var messages []models.ChatMessage
	offset := (req.Page - 1) * req.Limit

	if err := query.Preload("Sender").Preload("Sender.Profile").
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("獲取訊息失敗: %w", err)
	}

	// 轉換為響應格式
	responses := make([]dto.ChatMessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = dto.ChatMessageResponse{
			ID:          message.ID,
			ChatRoomID:  message.ChatRoomID,
			SenderID:    message.SenderID,
			Content:     message.Content,
			MessageType: message.MessageType,
			IsRead:      message.IsRead,
			CreatedAt:   message.CreatedAt,
		}

		// 添加發送者信息
		if message.Sender != nil && message.Sender.Profile != nil {
			responses[i].Sender = &dto.UserInfo{
				ID:        message.Sender.ID,
				FirstName: message.Sender.Profile.FirstName,
				LastName:  message.Sender.Profile.LastName,
				AvatarURL: message.Sender.Profile.AvatarURL,
			}
		}
	}

	return responses, nil
}

// GetChatRooms 獲取用戶的聊天室列表
func (uc *ChatUsecase) GetChatRooms(userID string) ([]dto.ChatRoomResponse, error) {
	var participants []models.ChatParticipant
	if err := uc.db.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("ChatRoom").
		Find(&participants).Error; err != nil {
		return nil, fmt.Errorf("獲取聊天室列表失敗: %w", err)
	}

	responses := make([]dto.ChatRoomResponse, 0, len(participants))
	for _, participant := range participants {
		if participant.ChatRoom != nil && participant.ChatRoom.IsActive {
			response, err := uc.getChatRoomResponse(participant.ChatRoom, userID)
			if err != nil {
				continue // 跳過錯誤的聊天室
			}
			responses = append(responses, *response)
		}
	}

	return responses, nil
}

// GetChatRoom 獲取單個聊天室信息
func (uc *ChatUsecase) GetChatRoom(chatRoomID, userID string) (*dto.ChatRoomResponse, error) {
	// 檢查用戶是否為聊天室參與者
	var participant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		chatRoomID, userID, true).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("您不是此聊天室的參與者")
		}
		return nil, fmt.Errorf("檢查參與者失敗: %w", err)
	}

	var chatRoom models.ChatRoom
	if err := uc.db.Where("id = ? AND is_active = ?", chatRoomID, true).
		First(&chatRoom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("聊天室不存在")
		}
		return nil, fmt.Errorf("獲取聊天室失敗: %w", err)
	}

	return uc.getChatRoomResponse(&chatRoom, userID)
}

// MarkMessagesAsRead 標記訊息為已讀
func (uc *ChatUsecase) MarkMessagesAsRead(userID, chatRoomID string) error {
	// 檢查用戶是否為聊天室參與者
	var participant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		chatRoomID, userID, true).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("您不是此聊天室的參與者")
		}
		return fmt.Errorf("檢查參與者失敗: %w", err)
	}

	// 更新參與者的最後讀取時間
	now := time.Now()
	if err := uc.db.Model(&participant).Update("last_read_at", now).Error; err != nil {
		return fmt.Errorf("更新讀取時間失敗: %w", err)
	}

	return nil
}

// LeaveChatRoom 離開聊天室
func (uc *ChatUsecase) LeaveChatRoom(userID, chatRoomID string) error {
	// 檢查用戶是否為聊天室參與者
	var participant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		chatRoomID, userID, true).First(&participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("您不是此聊天室的參與者")
		}
		return fmt.Errorf("檢查參與者失敗: %w", err)
	}

	// 標記為非活躍
	if err := uc.db.Model(&participant).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("離開聊天室失敗: %w", err)
	}

	return nil
}

// 輔助方法

// findDirectChatRoom 查找直接聊天室
func (uc *ChatUsecase) findDirectChatRoom(userID1, userID2 string) (*models.ChatRoom, error) {
	var chatRoom models.ChatRoom

	// 查找兩個用戶都參與的直接聊天室
	subQuery := uc.db.Table("chat_participants").
		Select("chat_room_id").
		Where("user_id IN (?, ?) AND is_active = ?", userID1, userID2, true).
		Group("chat_room_id").
		Having("COUNT(DISTINCT user_id) = 2")

	if err := uc.db.Where("id IN (?) AND type = ? AND is_active = ?",
		subQuery, "direct", true).First(&chatRoom).Error; err != nil {
		return nil, err
	}

	return &chatRoom, nil
}

// getChatRoomResponse 獲取聊天室響應
func (uc *ChatUsecase) getChatRoomResponse(chatRoom *models.ChatRoom, userID string) (*dto.ChatRoomResponse, error) {
	response := &dto.ChatRoomResponse{
		ID:        chatRoom.ID,
		MatchID:   chatRoom.MatchID,
		Type:      chatRoom.Type,
		Name:      chatRoom.Name,
		IsActive:  chatRoom.IsActive,
		CreatedAt: chatRoom.CreatedAt,
		UpdatedAt: chatRoom.UpdatedAt,
	}

	// 獲取參與者
	var participants []models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND is_active = ?", chatRoom.ID, true).
		Preload("User").Preload("User.Profile").
		Find(&participants).Error; err != nil {
		return nil, fmt.Errorf("獲取參與者失敗: %w", err)
	}

	response.Participants = make([]dto.ChatParticipantResponse, len(participants))
	for i, participant := range participants {
		response.Participants[i] = dto.ChatParticipantResponse{
			ID:         participant.ID,
			UserID:     participant.UserID,
			JoinedAt:   participant.JoinedAt,
			LastReadAt: participant.LastReadAt,
			IsActive:   participant.IsActive,
		}

		if participant.User != nil && participant.User.Profile != nil {
			response.Participants[i].User = &dto.UserInfo{
				ID:        participant.User.ID,
				FirstName: participant.User.Profile.FirstName,
				LastName:  participant.User.Profile.LastName,
				AvatarURL: participant.User.Profile.AvatarURL,
			}
		}
	}

	// 獲取最後一條訊息
	var lastMessage models.ChatMessage
	if err := uc.db.Where("chat_room_id = ?", chatRoom.ID).
		Preload("Sender").Preload("Sender.Profile").
		Order("created_at DESC").
		First(&lastMessage).Error; err == nil {

		messageResponse := &dto.ChatMessageResponse{
			ID:          lastMessage.ID,
			ChatRoomID:  lastMessage.ChatRoomID,
			SenderID:    lastMessage.SenderID,
			Content:     lastMessage.Content,
			MessageType: lastMessage.MessageType,
			IsRead:      lastMessage.IsRead,
			CreatedAt:   lastMessage.CreatedAt,
		}

		if lastMessage.Sender != nil && lastMessage.Sender.Profile != nil {
			messageResponse.Sender = &dto.UserInfo{
				ID:        lastMessage.Sender.ID,
				FirstName: lastMessage.Sender.Profile.FirstName,
				LastName:  lastMessage.Sender.Profile.LastName,
				AvatarURL: lastMessage.Sender.Profile.AvatarURL,
			}
		}

		response.LastMessage = messageResponse
	}

	// 計算未讀訊息數量
	var currentParticipant models.ChatParticipant
	if err := uc.db.Where("chat_room_id = ? AND user_id = ? AND is_active = ?",
		chatRoom.ID, userID, true).First(&currentParticipant).Error; err == nil {

		query := uc.db.Model(&models.ChatMessage{}).Where("chat_room_id = ?", chatRoom.ID)
		if currentParticipant.LastReadAt != nil {
			query = query.Where("created_at > ?", *currentParticipant.LastReadAt)
		}

		var unreadCount int64
		query.Count(&unreadCount)
		response.UnreadCount = int(unreadCount)
	}

	return response, nil
}

// getMessageResponse 獲取訊息響應
func (uc *ChatUsecase) getMessageResponse(messageID string) (*dto.ChatMessageResponse, error) {
	var message models.ChatMessage
	if err := uc.db.Where("id = ?", messageID).
		Preload("Sender").Preload("Sender.Profile").
		First(&message).Error; err != nil {
		return nil, fmt.Errorf("獲取訊息失敗: %w", err)
	}

	response := &dto.ChatMessageResponse{
		ID:          message.ID,
		ChatRoomID:  message.ChatRoomID,
		SenderID:    message.SenderID,
		Content:     message.Content,
		MessageType: message.MessageType,
		IsRead:      message.IsRead,
		CreatedAt:   message.CreatedAt,
	}

	if message.Sender != nil && message.Sender.Profile != nil {
		response.Sender = &dto.UserInfo{
			ID:        message.Sender.ID,
			FirstName: message.Sender.Profile.FirstName,
			LastName:  message.Sender.Profile.LastName,
			AvatarURL: message.Sender.Profile.AvatarURL,
		}
	}

	return response, nil
}

// removeDuplicates 移除重複的字符串
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
