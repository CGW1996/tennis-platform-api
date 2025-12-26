package db

import (
	"encoding/json"
	"log"

	"tennis-platform/backend/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seeder 種子數據管理器
type Seeder struct {
	db *gorm.DB
}

// NewSeeder 創建種子數據管理器
func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

// SeedAll 執行所有種子數據
func (s *Seeder) SeedAll() error {
	log.Println("Starting database seeding...")

	// 檢查是否已經有數據
	var userCount int64
	if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount > 0 {
		log.Println("Database already has data, skipping seeding")
		return nil
	}

	// 執行種子數據
	if err := s.seedUsers(); err != nil {
		return err
	}

	if err := s.seedCourts(); err != nil {
		return err
	}

	if err := s.seedCoaches(); err != nil {
		return err
	}

	if err := s.seedRackets(); err != nil {
		return err
	}

	if err := s.seedClubs(); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// seedUsers 創建測試用戶
func (s *Seeder) seedUsers() error {
	log.Println("Seeding users...")

	// 創建密碼哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []models.User{
		{
			ID:            uuid.New().String(),
			Email:         "admin@tennis-platform.com",
			PasswordHash:  string(passwordHash),
			EmailVerified: true,
			IsActive:      true,
			Profile: &models.UserProfile{
				FirstName:         "Admin",
				LastName:          "User",
				NTRPLevel:         func() *float64 { v := 4.5; return &v }(),
				PlayingStyle:      func() *string { v := "all-court"; return &v }(),
				PreferredHand:     func() *string { v := "right"; return &v }(),
				Latitude:          func() *float64 { v := 25.0330; return &v }(),
				Longitude:         func() *float64 { v := 121.5654; return &v }(),
				Bio:               func() *string { v := "Platform administrator and tennis enthusiast"; return &v }(),
				Gender:            func() *string { v := "male"; return &v }(),
				PlayingFrequency:  func() *string { v := "regular"; return &v }(),
				PreferredTimes:    pq.StringArray{"morning", "evening"},
				MaxTravelDistance: func() *float64 { v := 10.0; return &v }(),
			},
		},
		{
			ID:            uuid.New().String(),
			Email:         "test@example.com",
			PasswordHash:  string(passwordHash),
			EmailVerified: true,
			IsActive:      true,
			Profile: &models.UserProfile{
				FirstName:         "John",
				LastName:          "Doe",
				NTRPLevel:         func() *float64 { v := 3.5; return &v }(),
				PlayingStyle:      func() *string { v := "aggressive"; return &v }(),
				PreferredHand:     func() *string { v := "right"; return &v }(),
				Latitude:          func() *float64 { v := 25.0478; return &v }(),
				Longitude:         func() *float64 { v := 121.5319; return &v }(),
				Bio:               func() *string { v := "Love playing tennis on weekends"; return &v }(),
				Gender:            func() *string { v := "male"; return &v }(),
				PlayingFrequency:  func() *string { v := "casual"; return &v }(),
				PreferredTimes:    pq.StringArray{"weekend", "evening"},
				MaxTravelDistance: func() *float64 { v := 15.0; return &v }(),
			},
		},
		{
			ID:            uuid.New().String(),
			Email:         "jane.smith@example.com",
			PasswordHash:  string(passwordHash),
			EmailVerified: true,
			IsActive:      true,
			Profile: &models.UserProfile{
				FirstName:         "Jane",
				LastName:          "Smith",
				NTRPLevel:         func() *float64 { v := 4.0; return &v }(),
				PlayingStyle:      func() *string { v := "defensive"; return &v }(),
				PreferredHand:     func() *string { v := "left"; return &v }(),
				Latitude:          func() *float64 { v := 25.0173; return &v }(),
				Longitude:         func() *float64 { v := 121.5397; return &v }(),
				Bio:               func() *string { v := "Competitive player seeking challenging matches"; return &v }(),
				Gender:            func() *string { v := "male"; return &v }(),
				PlayingFrequency:  func() *string { v := "competitive"; return &v }(),
				PreferredTimes:    pq.StringArray{"morning", "afternoon"},
				MaxTravelDistance: func() *float64 { v := 20.0; return &v }(),
			},
		},
	}

	for _, user := range users {
		if err := s.db.Create(&user).Error; err != nil {
			return err
		}

		// 創建信譽分數
		reputationScore := models.ReputationScore{
			UserID:           user.ID,
			AttendanceRate:   95.0,
			PunctualityScore: 90.0,
			SkillAccuracy:    85.0,
			BehaviorRating:   4.5,
			TotalMatches:     10,
			CompletedMatches: 9,
			CancelledMatches: 1,
			OverallScore:     88.5,
		}
		if err := s.db.Create(&reputationScore).Error; err != nil {
			return err
		}
	}

	log.Printf("Created %d users", len(users))
	return nil
}

// seedCourts 創建測試場地
func (s *Seeder) seedCourts() error {
	log.Println("Seeding courts...")

	// 創建營業時間 JSON
	operatingHours1, _ := json.Marshal(map[string]string{
		"monday":    "06:00-22:00",
		"tuesday":   "06:00-22:00",
		"wednesday": "06:00-22:00",
		"thursday":  "06:00-22:00",
		"friday":    "06:00-22:00",
		"saturday":  "06:00-22:00",
		"sunday":    "06:00-22:00",
	})

	operatingHours2, _ := json.Marshal(map[string]string{
		"monday":    "07:00-21:00",
		"tuesday":   "07:00-21:00",
		"wednesday": "07:00-21:00",
		"thursday":  "07:00-21:00",
		"friday":    "07:00-21:00",
		"saturday":  "08:00-20:00",
		"sunday":    "08:00-20:00",
	})

	operatingHours3, _ := json.Marshal(map[string]string{
		"monday":    "05:30-22:00",
		"tuesday":   "05:30-22:00",
		"wednesday": "05:30-22:00",
		"thursday":  "05:30-22:00",
		"friday":    "05:30-22:00",
		"saturday":  "05:30-22:00",
		"sunday":    "05:30-22:00",
	})

	courts := []models.Court{
		{
			ID:             uuid.New().String(),
			Name:           "台北網球中心",
			Description:    func() *string { v := "專業網球場地，設備完善，適合各種水平的球員"; return &v }(),
			Address:        "台北市信義區松壽路20號",
			Latitude:       25.0330,
			Longitude:      121.5654,
			Facilities:     []string{"更衣室", "淋浴間", "停車場", "餐廳", "專業照明"},
			CourtType:      "hard",
			PricePerHour:   800.0,
			Currency:       "TWD",
			Images:         []string{"/images/courts/taipei-tennis-center-1.jpg", "/images/courts/taipei-tennis-center-2.jpg"},
			OperatingHours: operatingHours1,
			ContactPhone:   func() *string { v := "+886-2-2345-6789"; return &v }(),
			ContactEmail:   func() *string { v := "info@taipeitenniscenter.com"; return &v }(),
			Website:        func() *string { v := "https://taipeitenniscenter.com"; return &v }(),
			AverageRating:  4.5,
			TotalReviews:   25,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			Name:           "大安森林公園網球場",
			Description:    func() *string { v := "位於市中心的公園網球場，環境優美，價格實惠"; return &v }(),
			Address:        "台北市大安區新生南路二段1號",
			Latitude:       25.0173,
			Longitude:      121.5397,
			Facilities:     []string{"更衣室", "飲水機", "休息區"},
			CourtType:      "hard",
			PricePerHour:   400.0,
			Currency:       "TWD",
			Images:         []string{"/images/courts/daan-park-1.jpg"},
			OperatingHours: operatingHours2,
			ContactPhone:   func() *string { v := "+886-2-2700-1234"; return &v }(),
			AverageRating:  4.0,
			TotalReviews:   18,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			Name:           "天母運動公園網球場",
			Description:    func() *string { v := "天母地區知名網球場，場地寬敞，設施完善"; return &v }(),
			Address:        "台北市士林區忠誠路二段77號",
			Latitude:       25.1175,
			Longitude:      121.5252,
			Facilities:     []string{"更衣室", "淋浴間", "停車場", "商店", "觀眾席"},
			CourtType:      "clay",
			PricePerHour:   600.0,
			Currency:       "TWD",
			Images:         []string{"/images/courts/tianmu-1.jpg", "/images/courts/tianmu-2.jpg"},
			OperatingHours: operatingHours3,
			ContactPhone:   func() *string { v := "+886-2-2876-5432"; return &v }(),
			ContactEmail:   func() *string { v := "tianmu@sportspark.gov.tw"; return &v }(),
			AverageRating:  4.2,
			TotalReviews:   32,
			IsActive:       true,
		},
	}

	for _, court := range courts {
		if err := s.db.Create(&court).Error; err != nil {
			return err
		}
	}

	log.Printf("Created %d courts", len(courts))
	return nil
}

// seedCoaches 創建測試教練
func (s *Seeder) seedCoaches() error {
	log.Println("Seeding coaches...")

	// 先獲取一些用戶作為教練
	var users []models.User
	if err := s.db.Limit(2).Find(&users).Error; err != nil {
		return err
	}

	if len(users) < 2 {
		log.Println("Not enough users to create coaches")
		return nil
	}

	coaches := []models.Coach{
		{
			ID:             uuid.New().String(),
			UserID:         users[0].ID,
			LicenseNumber:  func() *string { v := "TTA-2023-001"; return &v }(),
			Certifications: []string{"ITF Level 2", "PTR Professional", "USPTA Certified"},
			Experience:     8,
			Specialties:    []string{"beginner", "intermediate", "junior"},
			Biography: func() *string {
				v := "專業網球教練，擁有8年教學經驗，擅長基礎技術教學和青少年培訓"
				return &v
			}(),
			HourlyRate:    1500.0,
			Currency:      "TWD",
			Languages:     []string{"中文", "English"},
			AverageRating: 4.8,
			TotalReviews:  45,
			TotalLessons:  200,
			IsVerified:    true,
			IsActive:      true,
			AvailableHours: map[string][]string{
				"monday":    {"09:00-12:00", "14:00-18:00"},
				"tuesday":   {"09:00-12:00", "14:00-18:00"},
				"wednesday": {"09:00-12:00", "14:00-18:00"},
				"thursday":  {"09:00-12:00", "14:00-18:00"},
				"friday":    {"09:00-12:00", "14:00-18:00"},
				"saturday":  {"08:00-12:00", "13:00-17:00"},
				"sunday":    {"08:00-12:00", "13:00-17:00"},
			},
		},
		{
			ID:             uuid.New().String(),
			UserID:         users[1].ID,
			LicenseNumber:  func() *string { v := "TTA-2023-002"; return &v }(),
			Certifications: []string{"ITF Level 3", "RPT Certified"},
			Experience:     12,
			Specialties:    []string{"intermediate", "advanced", "competitive"},
			Biography: func() *string {
				v := "前職業選手轉任教練，專精於競技網球訓練和戰術指導"
				return &v
			}(),
			HourlyRate:    2000.0,
			Currency:      "TWD",
			Languages:     []string{"中文", "English", "日本語"},
			AverageRating: 4.9,
			TotalReviews:  38,
			TotalLessons:  150,
			IsVerified:    true,
			IsActive:      true,
			AvailableHours: map[string][]string{
				"monday":    {"10:00-12:00", "15:00-19:00"},
				"tuesday":   {"10:00-12:00", "15:00-19:00"},
				"wednesday": {"10:00-12:00", "15:00-19:00"},
				"thursday":  {"10:00-12:00", "15:00-19:00"},
				"friday":    {"10:00-12:00", "15:00-19:00"},
				"saturday":  {"09:00-13:00"},
				"sunday":    {"09:00-13:00"},
			},
		},
	}

	for _, coach := range coaches {
		if err := s.db.Create(&coach).Error; err != nil {
			return err
		}
	}

	log.Printf("Created %d coaches", len(coaches))
	return nil
}

// seedRackets 創建測試球拍
func (s *Seeder) seedRackets() error {
	log.Println("Seeding rackets...")

	rackets := []models.Racket{
		{
			ID:             uuid.New().String(),
			Brand:          "Wilson",
			Model:          "Pro Staff 97",
			Year:           func() *int { v := 2023; return &v }(),
			HeadSize:       97,
			Weight:         315,
			Balance:        func() *int { v := 315; return &v }(),
			StringPattern:  "16x19",
			BeamWidth:      func() *float64 { v := 21.5; return &v }(),
			Length:         27,
			Stiffness:      func() *int { v := 68; return &v }(),
			SwingWeight:    func() *int { v := 335; return &v }(),
			PowerLevel:     func() *int { v := 6; return &v }(),
			ControlLevel:   func() *int { v := 9; return &v }(),
			ManeuverLevel:  func() *int { v := 7; return &v }(),
			StabilityLevel: func() *int { v := 8; return &v }(),
			Description: func() *string {
				v := "專業級網球拍，適合中高級球員，提供優異的控制性和穩定性"
				return &v
			}(),
			Images:        pq.StringArray{"/images/rackets/wilson-pro-staff-97-1.jpg", "/images/rackets/wilson-pro-staff-97-2.jpg"},
			MSRP:          func() *float64 { v := 8500.0; return &v }(),
			Currency:      "TWD",
			AverageRating: 4.6,
			TotalReviews:  28,
			IsActive:      true,
		},
		{
			ID:             uuid.New().String(),
			Brand:          "Babolat",
			Model:          "Pure Drive",
			Year:           func() *int { v := 2023; return &v }(),
			HeadSize:       100,
			Weight:         300,
			Balance:        func() *int { v := 320; return &v }(),
			StringPattern:  "16x19",
			BeamWidth:      func() *float64 { v := 23.0; return &v }(),
			Length:         27,
			Stiffness:      func() *int { v := 72; return &v }(),
			SwingWeight:    func() *int { v := 320; return &v }(),
			PowerLevel:     func() *int { v := 8; return &v }(),
			ControlLevel:   func() *int { v := 6; return &v }(),
			ManeuverLevel:  func() *int { v := 8; return &v }(),
			StabilityLevel: func() *int { v := 7; return &v }(),
			Description: func() *string {
				v := "經典力量型球拍，適合中級球員，提供強勁的力量和旋轉"
				return &v
			}(),
			Images:        pq.StringArray{"/images/rackets/babolat-pure-drive-1.jpg"},
			MSRP:          func() *float64 { v := 7200.0; return &v }(),
			Currency:      "TWD",
			AverageRating: 4.4,
			TotalReviews:  35,
			IsActive:      true,
		},
		{
			ID:             uuid.New().String(),
			Brand:          "Head",
			Model:          "Speed MP",
			Year:           func() *int { v := 2023; return &v }(),
			HeadSize:       100,
			Weight:         300,
			Balance:        func() *int { v := 320; return &v }(),
			StringPattern:  "16x19",
			BeamWidth:      func() *float64 { v := 23.0; return &v }(),
			Length:         27,
			Stiffness:      func() *int { v := 62; return &v }(),
			SwingWeight:    func() *int { v := 325; return &v }(),
			PowerLevel:     func() *int { v := 7; return &v }(),
			ControlLevel:   func() *int { v := 7; return &v }(),
			ManeuverLevel:  func() *int { v := 8; return &v }(),
			StabilityLevel: func() *int { v := 7; return &v }(),
			Description: func() *string {
				v := "平衡型球拍，適合各種打法，提供良好的力量和控制平衡"
				return &v
			}(),
			Images:        pq.StringArray{"/images/rackets/head-speed-mp-1.jpg", "/images/rackets/head-speed-mp-2.jpg"},
			MSRP:          func() *float64 { v := 6800.0; return &v }(),
			Currency:      "TWD",
			AverageRating: 4.3,
			TotalReviews:  22,
			IsActive:      true,
		},
	}

	for _, racket := range rackets {
		if err := s.db.Create(&racket).Error; err != nil {
			return err
		}
	}

	log.Printf("Created %d rackets", len(rackets))
	return nil
}

// seedClubs 創建測試俱樂部
func (s *Seeder) seedClubs() error {
	log.Println("Seeding clubs...")

	clubs := []models.Club{
		{
			ID:   uuid.New().String(),
			Name: "台北網球俱樂部",
			Description: func() *string {
				v := "台北市歷史最悠久的網球俱樂部，提供專業的網球訓練和社交活動"
				return &v
			}(),
			Address:      "台北市中正區羅斯福路一段4號",
			Latitude:     25.0418,
			Longitude:    121.5188,
			ContactPhone: func() *string { v := "+886-2-2321-5678"; return &v }(),
			ContactEmail: func() *string { v := "info@taipeitennisclub.com"; return &v }(),
			Website:      func() *string { v := "https://taipeitennisclub.com"; return &v }(),
			Images:       pq.StringArray{"/images/clubs/taipei-tennis-club-1.jpg", "/images/clubs/taipei-tennis-club-2.jpg"},
			Facilities:   pq.StringArray{"8個網球場", "健身房", "游泳池", "餐廳", "會議室", "停車場"},
			MembershipFees: map[string]float64{
				"monthly": 3000.0,
				"yearly":  30000.0,
			},
			Currency:       "TWD",
			MaxMembers:     func() *int { v := 500; return &v }(),
			CurrentMembers: 320,
			AverageRating:  4.5,
			TotalReviews:   42,
			IsActive:       true,
		},
		{
			ID:   uuid.New().String(),
			Name: "信義網球俱樂部",
			Description: func() *string {
				v := "位於信義區的現代化網球俱樂部，設施先進，服務優質"
				return &v
			}(),
			Address:      "台北市信義區基隆路一段200號",
			Latitude:     25.0478,
			Longitude:    121.5319,
			ContactPhone: func() *string { v := "+886-2-2345-9876"; return &v }(),
			ContactEmail: func() *string { v := "contact@xinyitennisclub.com"; return &v }(),
			Website:      func() *string { v := "https://xinyitennisclub.com"; return &v }(),
			Images:       pq.StringArray{"/images/clubs/xinyi-tennis-club-1.jpg"},
			Facilities:   pq.StringArray{"6個網球場", "健身房", "SPA", "咖啡廳", "專業教練", "器材租借"},
			MembershipFees: map[string]float64{
				"monthly": 4000.0,
				"yearly":  40000.0,
			},
			Currency:       "TWD",
			MaxMembers:     func() *int { v := 300; return &v }(),
			CurrentMembers: 180,
			AverageRating:  4.3,
			TotalReviews:   28,
			IsActive:       true,
		},
	}

	for _, club := range clubs {
		if err := s.db.Create(&club).Error; err != nil {
			return err
		}
	}

	log.Printf("Created %d clubs", len(clubs))
	return nil
}

// ClearAll 清除所有數據（僅用於開發環境）
func (s *Seeder) ClearAll() error {
	log.Println("Clearing all data...")

	// 按照依賴關係的逆序刪除
	tables := []interface{}{
		&models.ClubReview{},
		&models.ClubEventParticipant{},
		&models.ClubEvent{},
		&models.ClubMember{},
		&models.Club{},
		&models.RacketRecommendation{},
		&models.RacketPrice{},
		&models.RacketReview{},
		&models.Racket{},
		&models.LessonSchedule{},
		&models.Lesson{},
		&models.CoachReview{},
		&models.Coach{},
		&models.ReputationScore{},
		&models.ChatParticipant{},
		&models.ChatMessage{},
		&models.ChatRoom{},
		&models.MatchResult{},
		&models.MatchParticipant{},
		&models.Match{},
		&models.Booking{},
		&models.CourtReview{},
		&models.Court{},
		&models.RefreshToken{},
		&models.OAuthAccount{},
		&models.UserProfile{},
		&models.User{},
	}

	for _, table := range tables {
		if err := s.db.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			log.Printf("Warning: Failed to clear table %T: %v", table, err)
		}
	}

	log.Println("All data cleared")
	return nil
}
