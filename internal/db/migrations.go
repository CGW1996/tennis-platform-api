package db

import (
	"fmt"
	"log"
	"time"

	"tennis-platform/backend/internal/models"

	"gorm.io/gorm"
)

// Migration 遷移結構
type Migration struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// MigrationManager 遷移管理器
type MigrationManager struct {
	db *gorm.DB
}

// NewMigrationManager 創建遷移管理器
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// InitMigrationTable 初始化遷移表
func (m *MigrationManager) InitMigrationTable() error {
	return m.db.AutoMigrate(&Migration{})
}

// RunMigrations 執行所有遷移
func (m *MigrationManager) RunMigrations() error {
	// 初始化遷移表
	if err := m.InitMigrationTable(); err != nil {
		return fmt.Errorf("failed to init migration table: %w", err)
	}

	// 定義遷移列表
	migrations := []struct {
		version     string
		description string
		up          func(*gorm.DB) error
	}{
		{
			version:     "001_initial_schema",
			description: "Create initial database schema",
			up:          m.migration001InitialSchema,
		},
		{
			version:     "002_add_indexes",
			description: "Add database indexes for performance",
			up:          m.migration002AddIndexes,
		},
		{
			version:     "003_add_constraints",
			description: "Add database constraints and triggers",
			up:          m.migration003AddConstraints,
		},
		{
			version:     "004_add_privacy_fields",
			description: "Add privacy control fields to user profiles",
			up:          m.migration004AddPrivacyFields,
		},
		{
			version:     "005_add_review_reports",
			description: "Add review reporting and moderation system",
			up:          m.migration005AddReviewReports,
		},
		{
			version:     "006_add_card_matching",
			description: "Add card-based matching system tables",
			up:          m.migration006AddCardMatching,
		},
		{
			version:     "007_add_lesson_types",
			description: "Add lesson types table and update lessons table",
			up:          m.migration007AddLessonTypes,
		},
	}

	// 執行遷移
	for _, migration := range migrations {
		if err := m.runMigration(migration.version, migration.description, migration.up); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.version, err)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// runMigration 執行單個遷移
func (m *MigrationManager) runMigration(version, description string, up func(*gorm.DB) error) error {
	// 檢查遷移是否已經執行
	var count int64
	if err := m.db.Model(&Migration{}).Where("version = ?", version).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Migration %s already applied, skipping", version)
		return nil
	}

	// 執行遷移
	log.Printf("Running migration: %s - %s", version, description)

	// 開始事務
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 執行遷移函數
	if err := up(tx); err != nil {
		tx.Rollback()
		return err
	}

	// 記錄遷移
	migration := Migration{
		Version:     version,
		Description: description,
		AppliedAt:   time.Now(),
	}
	if err := tx.Create(&migration).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事務
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Migration %s completed successfully", version)
	return nil
}

// migration001InitialSchema 初始數據庫架構
func (m *MigrationManager) migration001InitialSchema(tx *gorm.DB) error {
	// 啟用必要的擴展
	requiredExtensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
	}

	optionalExtensions := []string{
		"CREATE EXTENSION IF NOT EXISTS postgis",
		"CREATE EXTENSION IF NOT EXISTS pg_trgm", // 用於模糊搜索
	}

	// 創建必需的擴展
	for _, ext := range requiredExtensions {
		if err := tx.Exec(ext).Error; err != nil {
			return fmt.Errorf("failed to create required extension: %w", err)
		}
	}

	// 嘗試創建可選的擴展
	for _, ext := range optionalExtensions {
		if err := tx.Exec(ext).Error; err != nil {
			log.Printf("Warning: Failed to create optional extension: %s, Error: %v", ext, err)
			// 繼續執行，不中斷遷移
		}
	}

	// 自動遷移所有模型
	if err := tx.AutoMigrate(models.AllModels()...); err != nil {
		return fmt.Errorf("failed to auto migrate models: %w", err)
	}

	return nil
}

// migration002AddIndexes 添加索引
func (m *MigrationManager) migration002AddIndexes(tx *gorm.DB) error {
	indexes := []string{
		// 用戶相關索引
		"CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_phone_active ON users(phone) WHERE deleted_at IS NULL AND phone IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_ntrp_level ON user_profiles(ntrp_level) WHERE ntrp_level IS NOT NULL",

		// 地理位置索引（需要PostGIS）
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_location ON user_profiles USING GIST(ST_Point(longitude, latitude)) WHERE longitude IS NOT NULL AND latitude IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_courts_location ON courts USING GIST(ST_Point(longitude, latitude))",
		"CREATE INDEX IF NOT EXISTS idx_clubs_location ON clubs USING GIST(ST_Point(longitude, latitude))",

		// 場地相關索引
		"CREATE INDEX IF NOT EXISTS idx_courts_price_active ON courts(price_per_hour) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_courts_rating_active ON courts(average_rating) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_court_reviews_rating_date ON court_reviews(rating, created_at) WHERE deleted_at IS NULL",

		// 配對相關索引
		"CREATE INDEX IF NOT EXISTS idx_matches_status_date ON matches(status, scheduled_at) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_match_participants_composite ON match_participants(user_id, match_id, status)",

		// 聊天相關索引
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_room_date ON chat_messages(chat_room_id, created_at) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_chat_participants_user_active ON chat_participants(user_id) WHERE is_active = true",

		// 教練相關索引
		"CREATE INDEX IF NOT EXISTS idx_coaches_rate_rating ON coaches(hourly_rate, average_rating) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_coaches_verified_active ON coaches(is_verified, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_lessons_date_status ON lessons(scheduled_at, status) WHERE deleted_at IS NULL",

		// 球拍相關索引
		"CREATE INDEX IF NOT EXISTS idx_rackets_brand_model_active ON rackets(brand, model) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_rackets_specs ON rackets(head_size, weight, ntrp_level) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_racket_prices_price_available ON racket_prices(price) WHERE deleted_at IS NULL AND is_available = true",

		// 俱樂部相關索引
		"CREATE INDEX IF NOT EXISTS idx_clubs_rating_active ON clubs(average_rating) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_club_events_date_status ON club_events(start_time, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_club_members_status ON club_members(status) WHERE deleted_at IS NULL",

		// 全文搜索索引
		"CREATE INDEX IF NOT EXISTS idx_courts_name_search ON courts USING GIN(to_tsvector('english', name)) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_courts_address_search ON courts USING GIN(to_tsvector('english', address)) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_clubs_name_search ON clubs USING GIN(to_tsvector('english', name)) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_rackets_search ON rackets USING GIN(to_tsvector('english', brand || ' ' || model)) WHERE deleted_at IS NULL",

		// 複合索引
		"CREATE INDEX IF NOT EXISTS idx_bookings_court_time ON bookings(court_id, start_time, end_time) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_reputation_scores_user ON reputation_scores(user_id, overall_score)",
	}

	for _, indexSQL := range indexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %s, Error: %v", indexSQL, err)
			// 繼續執行其他索引，不中斷遷移
		}
	}

	return nil
}

// migration003AddConstraints 添加約束和觸發器
func (m *MigrationManager) migration003AddConstraints(tx *gorm.DB) error {
	constraints := []string{
		// 添加檢查約束
		"ALTER TABLE user_profiles ADD CONSTRAINT check_ntrp_level_range CHECK (ntrp_level IS NULL OR (ntrp_level >= 1.0 AND ntrp_level <= 7.0))",
		"ALTER TABLE court_reviews ADD CONSTRAINT check_rating_range CHECK (rating >= 1 AND rating <= 5)",
		"ALTER TABLE coach_reviews ADD CONSTRAINT check_coach_rating_range CHECK (rating >= 1 AND rating <= 5)",
		"ALTER TABLE racket_reviews ADD CONSTRAINT check_racket_rating_range CHECK (rating >= 1 AND rating <= 5)",
		"ALTER TABLE club_reviews ADD CONSTRAINT check_club_rating_range CHECK (rating >= 1 AND rating <= 5)",

		// 添加唯一約束
		"ALTER TABLE oauth_accounts ADD CONSTRAINT unique_provider_user UNIQUE (provider, provider_id)",
		"ALTER TABLE club_members ADD CONSTRAINT unique_club_user_active UNIQUE (club_id, user_id) WHERE deleted_at IS NULL",

		// 添加外鍵約束（如果 GORM 沒有正確創建）
		"ALTER TABLE user_profiles ADD CONSTRAINT fk_user_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		"ALTER TABLE oauth_accounts ADD CONSTRAINT fk_oauth_accounts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
		"ALTER TABLE refresh_tokens ADD CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
	}

	for _, constraintSQL := range constraints {
		if err := tx.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to add constraint: %s, Error: %v", constraintSQL, err)
			// 某些約束可能已經存在，繼續執行
		}
	}

	// 創建觸發器函數
	triggerFunctions := []string{
		`CREATE OR REPLACE FUNCTION update_court_rating()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE courts 
			SET average_rating = (
				SELECT COALESCE(AVG(rating::numeric), 0)
				FROM court_reviews 
				WHERE court_id = NEW.court_id AND deleted_at IS NULL
			),
			total_reviews = (
				SELECT COUNT(*)
				FROM court_reviews 
				WHERE court_id = NEW.court_id AND deleted_at IS NULL
			)
			WHERE id = NEW.court_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`CREATE OR REPLACE FUNCTION update_coach_rating()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE coaches 
			SET average_rating = (
				SELECT COALESCE(AVG(rating::numeric), 0)
				FROM coach_reviews 
				WHERE coach_id = NEW.coach_id AND deleted_at IS NULL
			),
			total_reviews = (
				SELECT COUNT(*)
				FROM coach_reviews 
				WHERE coach_id = NEW.coach_id AND deleted_at IS NULL
			)
			WHERE id = NEW.coach_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`CREATE OR REPLACE FUNCTION update_racket_rating()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE rackets 
			SET average_rating = (
				SELECT COALESCE(AVG(rating::numeric), 0)
				FROM racket_reviews 
				WHERE racket_id = NEW.racket_id AND deleted_at IS NULL
			),
			total_reviews = (
				SELECT COUNT(*)
				FROM racket_reviews 
				WHERE racket_id = NEW.racket_id AND deleted_at IS NULL
			)
			WHERE id = NEW.racket_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`CREATE OR REPLACE FUNCTION update_club_rating()
		RETURNS TRIGGER AS $$
		BEGIN
			UPDATE clubs 
			SET average_rating = (
				SELECT COALESCE(AVG(rating::numeric), 0)
				FROM club_reviews 
				WHERE club_id = NEW.club_id AND deleted_at IS NULL
			),
			total_reviews = (
				SELECT COUNT(*)
				FROM club_reviews 
				WHERE club_id = NEW.club_id AND deleted_at IS NULL
			)
			WHERE id = NEW.club_id;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,
	}

	for _, funcSQL := range triggerFunctions {
		if err := tx.Exec(funcSQL).Error; err != nil {
			log.Printf("Warning: Failed to create trigger function: %v", err)
		}
	}

	// 創建觸發器
	triggers := []string{
		"DROP TRIGGER IF EXISTS trigger_update_court_rating ON court_reviews",
		"CREATE TRIGGER trigger_update_court_rating AFTER INSERT OR UPDATE OR DELETE ON court_reviews FOR EACH ROW EXECUTE FUNCTION update_court_rating()",

		"DROP TRIGGER IF EXISTS trigger_update_coach_rating ON coach_reviews",
		"CREATE TRIGGER trigger_update_coach_rating AFTER INSERT OR UPDATE OR DELETE ON coach_reviews FOR EACH ROW EXECUTE FUNCTION update_coach_rating()",

		"DROP TRIGGER IF EXISTS trigger_update_racket_rating ON racket_reviews",
		"CREATE TRIGGER trigger_update_racket_rating AFTER INSERT OR UPDATE OR DELETE ON racket_reviews FOR EACH ROW EXECUTE FUNCTION update_racket_rating()",

		"DROP TRIGGER IF EXISTS trigger_update_club_rating ON club_reviews",
		"CREATE TRIGGER trigger_update_club_rating AFTER INSERT OR UPDATE OR DELETE ON club_reviews FOR EACH ROW EXECUTE FUNCTION update_club_rating()",
	}

	for _, triggerSQL := range triggers {
		if err := tx.Exec(triggerSQL).Error; err != nil {
			log.Printf("Warning: Failed to create trigger: %s, Error: %v", triggerSQL, err)
		}
	}

	return nil
}

// GetAppliedMigrations 獲取已應用的遷移
func (m *MigrationManager) GetAppliedMigrations() ([]Migration, error) {
	var migrations []Migration
	err := m.db.Order("applied_at").Find(&migrations).Error
	return migrations, err
}

// migration004AddPrivacyFields 添加隱私控制欄位
func (m *MigrationManager) migration004AddPrivacyFields(tx *gorm.DB) error {
	privacyFields := []string{
		// 添加位置隱私控制欄位
		"ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS location_privacy BOOLEAN DEFAULT FALSE",

		// 添加檔案隱私控制欄位
		"ALTER TABLE user_profiles ADD COLUMN IF NOT EXISTS profile_privacy VARCHAR(20) DEFAULT 'public'",

		// 添加檔案隱私控制的約束
		"ALTER TABLE user_profiles ADD CONSTRAINT IF NOT EXISTS check_profile_privacy CHECK (profile_privacy IN ('public', 'friends', 'private'))",

		// 為隱私欄位添加索引以提升查詢性能
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_location_privacy ON user_profiles(location_privacy)",
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_profile_privacy ON user_profiles(profile_privacy)",
	}

	for _, fieldSQL := range privacyFields {
		if err := tx.Exec(fieldSQL).Error; err != nil {
			log.Printf("Warning: Failed to add privacy field: %s, Error: %v", fieldSQL, err)
			// 某些欄位可能已經存在，繼續執行
		}
	}

	// 添加註釋
	comments := []string{
		"COMMENT ON COLUMN user_profiles.location_privacy IS '位置隱私設定：true=隱藏精確位置，false=顯示精確位置'",
		"COMMENT ON COLUMN user_profiles.profile_privacy IS '檔案隱私設定：public=公開，friends=僅朋友，private=私人'",
	}

	for _, commentSQL := range comments {
		if err := tx.Exec(commentSQL).Error; err != nil {
			log.Printf("Warning: Failed to add comment: %s, Error: %v", commentSQL, err)
		}
	}

	return nil
}

// migration005AddReviewReports 添加評價舉報和審核系統
func (m *MigrationManager) migration005AddReviewReports(tx *gorm.DB) error {
	// 添加評價舉報相關欄位到 court_reviews 表
	reviewFields := []string{
		"ALTER TABLE court_reviews ADD COLUMN IF NOT EXISTS is_reported BOOLEAN DEFAULT FALSE",
		"ALTER TABLE court_reviews ADD COLUMN IF NOT EXISTS report_count INTEGER DEFAULT 0",
		"ALTER TABLE court_reviews ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active'",
		"ALTER TABLE court_reviews ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMP",
		"ALTER TABLE court_reviews ADD COLUMN IF NOT EXISTS moderated_by UUID",
	}

	for _, fieldSQL := range reviewFields {
		if err := tx.Exec(fieldSQL).Error; err != nil {
			log.Printf("Warning: Failed to add review field: %s, Error: %v", fieldSQL, err)
		}
	}

	// 添加狀態約束
	statusConstraints := []string{
		"ALTER TABLE court_reviews ADD CONSTRAINT IF NOT EXISTS check_review_status CHECK (status IN ('active', 'hidden', 'deleted'))",
	}

	for _, constraintSQL := range statusConstraints {
		if err := tx.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to add status constraint: %s, Error: %v", constraintSQL, err)
		}
	}

	// 創建評價舉報表（如果不存在）
	if err := tx.AutoMigrate(&models.ReviewReport{}); err != nil {
		return fmt.Errorf("failed to create review_reports table: %w", err)
	}

	// 添加評價舉報相關索引
	reportIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_court_reviews_status ON court_reviews(status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_court_reviews_reported ON court_reviews(is_reported) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_review_reports_status ON review_reports(status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_review_reports_review_user ON review_reports(review_id, user_id) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range reportIndexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create report index: %s, Error: %v", indexSQL, err)
		}
	}

	// 添加註釋
	comments := []string{
		"COMMENT ON COLUMN court_reviews.is_reported IS '是否被舉報'",
		"COMMENT ON COLUMN court_reviews.report_count IS '舉報次數'",
		"COMMENT ON COLUMN court_reviews.status IS '評價狀態：active=正常，hidden=隱藏，deleted=已刪除'",
		"COMMENT ON COLUMN court_reviews.moderated_at IS '審核時間'",
		"COMMENT ON COLUMN court_reviews.moderated_by IS '審核人員ID'",
		"COMMENT ON TABLE review_reports IS '評價舉報記錄表'",
	}

	for _, commentSQL := range comments {
		if err := tx.Exec(commentSQL).Error; err != nil {
			log.Printf("Warning: Failed to add comment: %s, Error: %v", commentSQL, err)
		}
	}

	return nil
}

// migration006AddCardMatching 添加抽卡配對系統表
func (m *MigrationManager) migration006AddCardMatching(tx *gorm.DB) error {
	// 創建抽卡互動表
	if err := tx.AutoMigrate(&models.CardInteraction{}); err != nil {
		return fmt.Errorf("failed to create card_interactions table: %w", err)
	}

	// 創建配對通知表
	if err := tx.AutoMigrate(&models.MatchNotification{}); err != nil {
		return fmt.Errorf("failed to create match_notifications table: %w", err)
	}

	// 添加抽卡互動相關索引
	cardIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_card_interactions_user_action ON card_interactions(user_id, action) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_card_interactions_target_user ON card_interactions(target_user_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_card_interactions_match ON card_interactions(is_match) WHERE deleted_at IS NULL AND is_match = true",
		"CREATE INDEX IF NOT EXISTS idx_card_interactions_created_at ON card_interactions(created_at) WHERE deleted_at IS NULL",

		// 複合索引用於檢查互相喜歡
		"CREATE INDEX IF NOT EXISTS idx_card_interactions_mutual_like ON card_interactions(user_id, target_user_id, action) WHERE deleted_at IS NULL AND action = 'like'",

		// 避免重複互動的唯一索引
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_card_interactions_unique_user_target ON card_interactions(user_id, target_user_id) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range cardIndexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create card interaction index: %s, Error: %v", indexSQL, err)
		}
	}

	// 添加配對通知相關索引
	notificationIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_match_notifications_user_type ON match_notifications(user_id, type) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_match_notifications_unread ON match_notifications(user_id, is_read) WHERE deleted_at IS NULL AND is_read = false",
		"CREATE INDEX IF NOT EXISTS idx_match_notifications_created_at ON match_notifications(created_at) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range notificationIndexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create notification index: %s, Error: %v", indexSQL, err)
		}
	}

	// 添加約束
	constraints := []string{
		"ALTER TABLE card_interactions ADD CONSTRAINT IF NOT EXISTS check_card_action CHECK (action IN ('like', 'dislike', 'skip'))",
		"ALTER TABLE match_notifications ADD CONSTRAINT IF NOT EXISTS check_notification_type CHECK (type IN ('match_success', 'match_request', 'match_cancelled'))",
	}

	for _, constraintSQL := range constraints {
		if err := tx.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to add card matching constraint: %s, Error: %v", constraintSQL, err)
		}
	}

	// 添加註釋
	comments := []string{
		"COMMENT ON TABLE card_interactions IS '抽卡互動記錄表'",
		"COMMENT ON COLUMN card_interactions.action IS '互動動作：like=喜歡，dislike=不喜歡，skip=跳過'",
		"COMMENT ON COLUMN card_interactions.is_match IS '是否配對成功'",
		"COMMENT ON TABLE match_notifications IS '配對通知表'",
		"COMMENT ON COLUMN match_notifications.type IS '通知類型：match_success=配對成功，match_request=配對請求，match_cancelled=配對取消'",
		"COMMENT ON COLUMN match_notifications.data IS '通知額外數據（JSON格式）'",
	}

	for _, commentSQL := range comments {
		if err := tx.Exec(commentSQL).Error; err != nil {
			log.Printf("Warning: Failed to add comment: %s, Error: %v", commentSQL, err)
		}
	}

	return nil
}

// migration007AddLessonTypes 添加課程類型表和更新課程表
func (m *MigrationManager) migration007AddLessonTypes(tx *gorm.DB) error {
	// 創建課程類型表
	if err := tx.AutoMigrate(&models.LessonType{}); err != nil {
		return fmt.Errorf("failed to create lesson_types table: %w", err)
	}

	// 添加課程類型ID欄位到課程表
	lessonFields := []string{
		"ALTER TABLE lessons ADD COLUMN IF NOT EXISTS lesson_type_id UUID",
		"ALTER TABLE lessons ADD COLUMN IF NOT EXISTS cancel_reason TEXT",
	}

	for _, fieldSQL := range lessonFields {
		if err := tx.Exec(fieldSQL).Error; err != nil {
			log.Printf("Warning: Failed to add lesson field: %s, Error: %v", fieldSQL, err)
		}
	}

	// 添加外鍵約束
	constraints := []string{
		"ALTER TABLE lessons ADD CONSTRAINT IF NOT EXISTS fk_lessons_lesson_type FOREIGN KEY (lesson_type_id) REFERENCES lesson_types(id) ON DELETE SET NULL",
	}

	for _, constraintSQL := range constraints {
		if err := tx.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to add lesson constraint: %s, Error: %v", constraintSQL, err)
		}
	}

	// 添加課程類型相關索引
	lessonTypeIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_lesson_types_coach_active ON lesson_types(coach_id, is_active) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_lesson_types_type_level ON lesson_types(type, level) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_lesson_types_price ON lesson_types(price) WHERE deleted_at IS NULL AND is_active = true",
		"CREATE INDEX IF NOT EXISTS idx_lessons_lesson_type ON lessons(lesson_type_id) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_lessons_coach_date ON lessons(coach_id, scheduled_at) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_lessons_student_date ON lessons(student_id, scheduled_at) WHERE deleted_at IS NULL",
	}

	for _, indexSQL := range lessonTypeIndexes {
		if err := tx.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create lesson type index: %s, Error: %v", indexSQL, err)
		}
	}

	// 添加約束
	lessonTypeConstraints := []string{
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_type CHECK (type IN ('individual', 'group', 'clinic'))",
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_level CHECK (level IS NULL OR level IN ('beginner', 'intermediate', 'advanced'))",
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_duration CHECK (duration >= 30 AND duration <= 480)",
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_price CHECK (price >= 0)",
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_participants CHECK (max_participants IS NULL OR max_participants >= 1)",
		"ALTER TABLE lesson_types ADD CONSTRAINT IF NOT EXISTS check_lesson_type_min_max CHECK (min_participants IS NULL OR max_participants IS NULL OR min_participants <= max_participants)",
	}

	for _, constraintSQL := range lessonTypeConstraints {
		if err := tx.Exec(constraintSQL).Error; err != nil {
			log.Printf("Warning: Failed to add lesson type constraint: %s, Error: %v", constraintSQL, err)
		}
	}

	// 添加註釋
	comments := []string{
		"COMMENT ON TABLE lesson_types IS '課程類型表'",
		"COMMENT ON COLUMN lesson_types.type IS '課程類型：individual=個人課程，group=團體課程，clinic=訓練營'",
		"COMMENT ON COLUMN lesson_types.level IS '課程等級：beginner=初級，intermediate=中級，advanced=高級'",
		"COMMENT ON COLUMN lesson_types.duration IS '課程時長（分鐘）'",
		"COMMENT ON COLUMN lesson_types.max_participants IS '最大參與人數（團體課程用）'",
		"COMMENT ON COLUMN lesson_types.min_participants IS '最小參與人數（團體課程用）'",
		"COMMENT ON COLUMN lesson_types.equipment IS '需要的設備'",
		"COMMENT ON COLUMN lesson_types.prerequisites IS '先決條件'",
		"COMMENT ON COLUMN lessons.lesson_type_id IS '關聯的課程類型ID'",
		"COMMENT ON COLUMN lessons.cancel_reason IS '取消原因'",
	}

	for _, commentSQL := range comments {
		if err := tx.Exec(commentSQL).Error; err != nil {
			log.Printf("Warning: Failed to add comment: %s, Error: %v", commentSQL, err)
		}
	}

	return nil
}

// RollbackMigration 回滾遷移（僅用於開發環境）
func (m *MigrationManager) RollbackMigration(version string) error {
	return m.db.Where("version = ?", version).Delete(&Migration{}).Error
}
