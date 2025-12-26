package models

// AllModels 返回所有需要遷移的模型
func AllModels() []interface{} {
	return []interface{}{
		// 用戶相關
		&User{},
		&UserProfile{},
		&OAuthAccount{},
		&RefreshToken{},

		// 場地相關
		&Court{},
		&CourtReview{},
		&ReviewReport{},
		&Booking{},

		// 配對和聊天相關
		&Match{},
		&MatchParticipant{},
		&MatchResult{},
		&ChatRoom{},
		&ChatMessage{},
		&ChatParticipant{},
		&ReputationScore{},
		&PunctualityRecord{},
		&SkillAccuracyRecord{},
		&BehaviorReview{},
		&CardInteraction{},
		&MatchNotification{},
		&SkillLevelRecord{},
		&UserPrivacySettings{},

		// 教練相關
		&Coach{},
		&CoachReview{},
		&LessonType{},
		&Lesson{},
		&LessonSchedule{},

		// 球拍相關
		&Racket{},
		&RacketReview{},
		&RacketPrice{},
		&RacketRecommendation{},

		// 俱樂部相關
		&Club{},
		&ClubMember{},
		&ClubEvent{},
		&ClubEventParticipant{},
		&ClubReview{},
	}
}
