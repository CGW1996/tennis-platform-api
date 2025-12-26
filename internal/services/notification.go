package services

import (
	"fmt"
	"tennis-platform/backend/internal/models"
	"time"
)

// NotificationService 通知服務接口
type NotificationService interface {
	SendBookingConfirmation(booking *models.Booking) error
	SendBookingReminder(booking *models.Booking) error
	SendBookingCancellation(booking *models.Booking) error
	SendBookingStatusUpdate(booking *models.Booking, oldStatus string) error
	SendMatchNotification(user *models.User, notification *models.MatchNotification) error
}

// EmailNotificationService 郵件通知服務實現
type EmailNotificationService struct {
	emailService *EmailService
}

// NewEmailNotificationService 創建新的郵件通知服務
func NewEmailNotificationService(emailService *EmailService) *EmailNotificationService {
	return &EmailNotificationService{
		emailService: emailService,
	}
}

// SendBookingConfirmation 發送預訂確認通知
func (ns *EmailNotificationService) SendBookingConfirmation(booking *models.Booking) error {
	if booking.User == nil || booking.Court == nil {
		return fmt.Errorf("booking user or court information is missing")
	}

	subject := "預訂確認 - " + booking.Court.Name

	body := fmt.Sprintf(`
親愛的用戶，

您的場地預訂已確認！

預訂詳情：
- 場地：%s
- 地址：%s
- 時間：%s 至 %s
- 總價：%.2f %s
- 預訂編號：%s

如有任何問題，請聯繫我們。

祝您打球愉快！
網球平台團隊
	`,
		booking.Court.Name,
		booking.Court.Address,
		booking.StartTime.Format("2006-01-02 15:04"),
		booking.EndTime.Format("2006-01-02 15:04"),
		booking.TotalPrice,
		"TWD", // 可以從 court 獲取
		booking.ID,
	)

	return ns.emailService.SendEmail(booking.User.Email, subject, body)
}

// SendBookingReminder 發送預訂提醒通知
func (ns *EmailNotificationService) SendBookingReminder(booking *models.Booking) error {
	if booking.User == nil || booking.Court == nil {
		return fmt.Errorf("booking user or court information is missing")
	}

	subject := "預訂提醒 - " + booking.Court.Name

	body := fmt.Sprintf(`
親愛的用戶，

提醒您即將到來的場地預訂：

預訂詳情：
- 場地：%s
- 地址：%s
- 時間：%s 至 %s
- 預訂編號：%s

請準時到達場地。

網球平台團隊
	`,
		booking.Court.Name,
		booking.Court.Address,
		booking.StartTime.Format("2006-01-02 15:04"),
		booking.EndTime.Format("2006-01-02 15:04"),
		booking.ID,
	)

	return ns.emailService.SendEmail(booking.User.Email, subject, body)
}

// SendBookingCancellation 發送預訂取消通知
func (ns *EmailNotificationService) SendBookingCancellation(booking *models.Booking) error {
	if booking.User == nil || booking.Court == nil {
		return fmt.Errorf("booking user or court information is missing")
	}

	subject := "預訂取消確認 - " + booking.Court.Name

	body := fmt.Sprintf(`
親愛的用戶，

您的場地預訂已成功取消。

取消的預訂詳情：
- 場地：%s
- 原定時間：%s 至 %s
- 預訂編號：%s

如有疑問，請聯繫我們。

網球平台團隊
	`,
		booking.Court.Name,
		booking.StartTime.Format("2006-01-02 15:04"),
		booking.EndTime.Format("2006-01-02 15:04"),
		booking.ID,
	)

	return ns.emailService.SendEmail(booking.User.Email, subject, body)
}

// SendBookingStatusUpdate 發送預訂狀態更新通知
func (ns *EmailNotificationService) SendBookingStatusUpdate(booking *models.Booking, oldStatus string) error {
	if booking.User == nil || booking.Court == nil {
		return fmt.Errorf("booking user or court information is missing")
	}

	statusMap := map[string]string{
		"pending":   "待確認",
		"confirmed": "已確認",
		"cancelled": "已取消",
		"completed": "已完成",
	}

	oldStatusText := statusMap[oldStatus]
	newStatusText := statusMap[booking.Status]

	subject := "預訂狀態更新 - " + booking.Court.Name

	body := fmt.Sprintf(`
親愛的用戶，

您的預訂狀態已更新：

預訂詳情：
- 場地：%s
- 時間：%s 至 %s
- 預訂編號：%s
- 狀態變更：%s → %s

網球平台團隊
	`,
		booking.Court.Name,
		booking.StartTime.Format("2006-01-02 15:04"),
		booking.EndTime.Format("2006-01-02 15:04"),
		booking.ID,
		oldStatusText,
		newStatusText,
	)

	return ns.emailService.SendEmail(booking.User.Email, subject, body)
}

// BookingReminderScheduler 預訂提醒調度器
type BookingReminderScheduler struct {
	notificationService NotificationService
}

// NewBookingReminderScheduler 創建新的預訂提醒調度器
func NewBookingReminderScheduler(notificationService NotificationService) *BookingReminderScheduler {
	return &BookingReminderScheduler{
		notificationService: notificationService,
	}
}

// ScheduleReminder 安排提醒
func (brs *BookingReminderScheduler) ScheduleReminder(booking *models.Booking) {
	// 計算提醒時間（預訂開始前1小時）
	reminderTime := booking.StartTime.Add(-1 * time.Hour)

	// 如果提醒時間已經過了，就不安排提醒
	if reminderTime.Before(time.Now()) {
		return
	}

	// 在實際應用中，這裡應該使用任務隊列或定時任務系統
	// 這裡只是一個簡單的示例
	go func() {
		time.Sleep(time.Until(reminderTime))
		brs.notificationService.SendBookingReminder(booking)
	}()
}

// MockNotificationService 模擬通知服務（用於測試）
type MockNotificationService struct{}

// NewMockNotificationService 創建模擬通知服務
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{}
}

// SendBookingConfirmation 模擬發送預訂確認通知
func (mns *MockNotificationService) SendBookingConfirmation(booking *models.Booking) error {
	fmt.Printf("Mock: Sending booking confirmation for booking %s\n", booking.ID)
	return nil
}

// SendBookingReminder 模擬發送預訂提醒通知
func (mns *MockNotificationService) SendBookingReminder(booking *models.Booking) error {
	fmt.Printf("Mock: Sending booking reminder for booking %s\n", booking.ID)
	return nil
}

// SendBookingCancellation 模擬發送預訂取消通知
func (mns *MockNotificationService) SendBookingCancellation(booking *models.Booking) error {
	fmt.Printf("Mock: Sending booking cancellation for booking %s\n", booking.ID)
	return nil
}

// SendBookingStatusUpdate 模擬發送預訂狀態更新通知
func (mns *MockNotificationService) SendBookingStatusUpdate(booking *models.Booking, oldStatus string) error {
	fmt.Printf("Mock: Sending booking status update for booking %s: %s -> %s\n", booking.ID, oldStatus, booking.Status)
	return nil
}

// SendMatchNotification 發送配對通知
func (ns *EmailNotificationService) SendMatchNotification(user *models.User, notification *models.MatchNotification) error {
	if user == nil {
		return fmt.Errorf("user information is missing")
	}

	subject := notification.Title

	body := fmt.Sprintf(`
親愛的 %s，

%s

這是來自網球平台的配對通知。請登入應用程式查看詳細資訊。

網球平台團隊
	`,
		user.Email, // 可以用真實姓名替換
		notification.Message,
	)

	return ns.emailService.SendEmail(user.Email, subject, body)
}

// SendMatchNotification 模擬發送配對通知
func (mns *MockNotificationService) SendMatchNotification(user *models.User, notification *models.MatchNotification) error {
	fmt.Printf("Mock: Sending match notification to user %s: %s\n", user.ID, notification.Message)
	return nil
}
