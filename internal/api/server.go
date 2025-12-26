package api

import (
	"net/http"
	"tennis-platform/backend/internal/config"
	"tennis-platform/backend/internal/controllers"
	"tennis-platform/backend/internal/db"
	"tennis-platform/backend/internal/middleware"
	"tennis-platform/backend/internal/services"
	"tennis-platform/backend/internal/usecases"

	_ "tennis-platform/backend/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server API 服務器
type Server struct {
	config                    *config.Config
	database                  *db.Database
	router                    *gin.Engine
	jwtService                *services.JWTService
	websocketService          *services.WebSocketService
	authController            *controllers.AuthController
	userController            *controllers.UserController
	courtController           *controllers.CourtController
	coachController           *controllers.CoachController
	matchingController        *controllers.MatchingController
	chatController            *controllers.ChatController
	reputationController      *controllers.ReputationController
	matchStatisticsController *controllers.MatchStatisticsController
	racketController          *controllers.RacketController
}

// NewServer 創建新的 API 服務器
func NewServer(cfg *config.Config, database *db.Database) *Server {
	// 設置 Gin 模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化服務層
	jwtService := services.NewJWTService(cfg)
	uploadService := services.NewUploadService(cfg)
	websocketService := services.NewWebSocketService()

	// 初始化通知服務
	var notificationService services.NotificationService
	if cfg.Env == "production" {
		// 在生產環境中使用真實的郵件通知服務
		emailService := services.NewEmailService(cfg)
		notificationService = services.NewEmailNotificationService(emailService)
	} else {
		// 在開發環境中使用模擬通知服務
		notificationService = services.NewMockNotificationService()
	}

	// 初始化用例層
	authUsecase := usecases.NewAuthUsecase(database.DB, cfg)
	userUsecase := usecases.NewUserUsecase(database.DB)
	courtUsecase := usecases.NewCourtUsecase(database.DB)
	reviewUsecase := usecases.NewReviewUsecase(database.DB, uploadService)
	bookingUsecase := usecases.NewBookingUsecase(database.DB, notificationService)
	coachUsecase := usecases.NewCoachUsecase(database.DB)
	matchingUsecase := usecases.NewMatchingUsecase(database.DB)
	chatUsecase := usecases.NewChatUsecase(database.DB)
	racketUsecase := usecases.NewRacketUsecase(database.DB)
	racketPriceUsecase := usecases.NewRacketPriceUsecase(database.DB)
	racketReviewUsecase := usecases.NewRacketReviewUsecase(database.DB)

	// 初始化控制器層
	authController := controllers.NewAuthController(authUsecase)
	userController := controllers.NewUserController(userUsecase, uploadService)
	courtController := controllers.NewCourtController(courtUsecase, reviewUsecase, bookingUsecase, uploadService)
	coachController := controllers.NewCoachController(coachUsecase)
	matchingController := controllers.NewMatchingController(matchingUsecase)
	chatController := controllers.NewChatController(chatUsecase, websocketService)
	reputationController := controllers.NewReputationController(database.DB)
	matchStatisticsController := controllers.NewMatchStatisticsController(database.DB)
	racketController := controllers.NewRacketController(racketUsecase, racketPriceUsecase, racketReviewUsecase, uploadService)

	server := &Server{
		config:     cfg,
		database:   database,
		router:     gin.Default(),
		jwtService: jwtService,

		websocketService:          websocketService,
		authController:            authController,
		userController:            userController,
		courtController:           courtController,
		coachController:           coachController,
		matchingController:        matchingController,
		chatController:            chatController,
		reputationController:      reputationController,
		matchStatisticsController: matchStatisticsController,
		racketController:          racketController,
	}

	// Disable automatic redirect for trailing slash
	server.router.RedirectTrailingSlash = false

	server.setupRoutes()
	return server
}

// setupRoutes 設置路由
func (s *Server) setupRoutes() {
	// CORS 中間件配置
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000", // Next.js 開發服務器
		"http://127.0.0.1:3000",
		s.config.FrontendURL, // 從配置讀取前端 URL
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"Accept",
		"Cache-Control",
		"X-Requested-With",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true

	s.router.Use(cors.New(config))

	// 靜態文件服務
	s.router.Static("/uploads", s.config.Upload.UploadPath)

	// Swagger 文檔
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 重定向根路徑的 swagger 到 index.html
	s.router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// 健康檢查
	s.router.GET("/health", s.healthCheck)

	// API v1 路由組
	v1 := s.router.Group("/api/v1")
	{
		// 認證相關路由（無需認證）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", s.authController.Register)
			auth.POST("/login", s.authController.Login)
			auth.POST("/refresh", s.authController.RefreshToken)
			auth.POST("/logout", s.authController.Logout)
			auth.POST("/forgot-password", s.authController.ForgotPassword)
			auth.POST("/reset-password", s.authController.ResetPassword)

			// OAuth 相關路由
			oauth := auth.Group("/oauth")
			{
				oauth.GET("/:provider", s.authController.GetOAuthAuthURL)
				oauth.POST("/:provider/callback", s.authController.OAuthCallback)
			}
		}

		// 需要認證的路由
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(s.jwtService))
		{
			// 用戶相關路由
			users := protected.Group("/users")
			{
				users.GET("/profile", s.userController.GetProfile)
				users.POST("/profile", s.userController.CreateProfile)
				users.PUT("/profile", s.userController.UpdateProfile)
				users.PUT("/preferences", s.userController.UpdatePreferences)
				users.PUT("/location", s.userController.UpdateLocation)
				users.POST("/avatar", s.userController.UploadAvatar)
			}

			// OAuth 帳號管理路由（需要認證）
			oauthProtected := protected.Group("/auth/oauth")
			{
				oauthProtected.GET("/accounts", s.authController.GetLinkedOAuthAccounts)
				oauthProtected.POST("/:provider/link", s.authController.LinkOAuthAccount)
				oauthProtected.DELETE("/:provider/unlink", s.authController.UnlinkOAuthAccount)
			}
		}

		// 公開路由（不需要認證）
		public := v1.Group("/")
		{
			// NTRP 等級信息
			public.GET("/users/ntrp-levels", s.userController.GetNTRPLevels)
		}

		// 場地相關路由
		courts := v1.Group("/courts")
		{
			// 公開路由
			courts.GET("", s.courtController.SearchCourts)
			courts.GET("/facilities", s.courtController.GetAvailableFacilities)
			courts.GET("/types", s.courtController.GetCourtTypes)
			courts.GET("/availability", s.courtController.GetCourtAvailability)
			courts.GET("/:id", s.courtController.GetCourt)
			courts.GET("/:id/reviews/statistics", s.courtController.GetReviewStatistics)

			// 需要認證的路由
			courtsProtected := courts.Group("/")
			courtsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				courtsProtected.POST("", s.courtController.CreateCourt)
				courtsProtected.PUT("/:id", s.courtController.UpdateCourt)
				courtsProtected.DELETE("/:id", s.courtController.DeleteCourt)
				courtsProtected.POST("/:id/images", s.courtController.UploadCourtImages)

			}
		}

		// 評價相關路由
		reviews := v1.Group("/reviews")
		{
			// 公開路由
			reviews.GET("", s.courtController.GetReviews)
			reviews.GET("/:id", s.courtController.GetReview)

			// 需要認證的路由
			reviewsProtected := reviews.Group("/")
			reviewsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				reviewsProtected.POST("", s.courtController.CreateReview)
				reviewsProtected.PUT("/:id", s.courtController.UpdateReview)
				reviewsProtected.DELETE("/:id", s.courtController.DeleteReview)
				reviewsProtected.POST("/:id/report", s.courtController.ReportReview)
				reviewsProtected.POST("/:id/helpful", s.courtController.MarkReviewHelpful)
				reviewsProtected.POST("/images", s.courtController.UploadReviewImages)
			}
		}

		// 預訂相關路由
		bookings := v1.Group("/bookings")
		bookings.Use(middleware.AuthMiddleware(s.jwtService))
		{
			bookings.POST("", s.courtController.CreateBooking)
			bookings.GET("", s.courtController.GetBookings)
			bookings.GET("/:id", s.courtController.GetBooking)
			bookings.PUT("/:id", s.courtController.UpdateBooking)
			bookings.POST("/:id/cancel", s.courtController.CancelBooking)
		}

		// 教練相關路由
		coaches := v1.Group("/coaches")
		{
			// 公開路由
			coaches.GET("", s.coachController.SearchCoaches)
			coaches.GET("/specialties", s.coachController.GetCoachSpecialties)
			coaches.GET("/certifications", s.coachController.GetCoachCertifications)
			coaches.GET("/languages", s.coachController.GetAvailableLanguages)
			coaches.GET("/currencies", s.coachController.GetAvailableCurrencies)
			coaches.GET("/:id", s.coachController.GetCoach)
			coaches.GET("/:id/statistics", s.coachController.GetCoachStatistics)
			coaches.GET("/:id/lesson-types", s.coachController.GetLessonTypes)
			coaches.GET("/:id/availability", s.coachController.GetCoachAvailability)
			coaches.GET("/:id/schedule", s.coachController.GetCoachSchedule)
			coaches.GET("/:id/review-statistics", s.coachController.GetCoachReviewStatistics)

			// 需要認證的路由
			coachesProtected := coaches.Group("")
			coachesProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				coachesProtected.POST("", s.coachController.CreateCoachProfile)
				coachesProtected.GET("/my-profile", s.coachController.GetMyCoachProfile)
				coachesProtected.PUT("/:id", s.coachController.UpdateCoachProfile)
				coachesProtected.POST("/verify", s.coachController.VerifyCoach)

				// 課程類型管理
				coachesProtected.POST("/lesson-types", s.coachController.CreateLessonType)
				coachesProtected.PUT("/schedule", s.coachController.UpdateCoachSchedule)
			}
		}

		// 課程類型相關路由
		lessonTypes := v1.Group("/lesson-types")
		lessonTypes.Use(middleware.AuthMiddleware(s.jwtService))
		{
			lessonTypes.PUT("/:id", s.coachController.UpdateLessonType)
			lessonTypes.DELETE("/:id", s.coachController.DeleteLessonType)
		}

		// 課程相關路由
		lessons := v1.Group("/lessons")
		{
			// 需要認證的路由
			lessonsProtected := lessons.Group("/")
			lessonsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				lessonsProtected.POST("", s.coachController.CreateLesson)
				lessonsProtected.GET("", s.coachController.GetLessons)
				lessonsProtected.GET("/:id", s.coachController.GetLesson)
				lessonsProtected.PUT("/:id", s.coachController.UpdateLesson)
				lessonsProtected.POST("/:id/cancel", s.coachController.CancelLesson)
			}
		}

		// 教練評價相關路由
		coachReviews := v1.Group("/coach-reviews")
		{
			// 公開路由
			coachReviews.GET("", s.coachController.GetCoachReviews)
			coachReviews.GET("/:id", s.coachController.GetCoachReview)
			coachReviews.GET("/available-tags", s.coachController.GetAvailableReviewTags)

			// 需要認證的路由
			coachReviewsProtected := coachReviews.Group("/")
			coachReviewsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				coachReviewsProtected.POST("", s.coachController.CreateCoachReview)
				coachReviewsProtected.PUT("/:id", s.coachController.UpdateCoachReview)
				coachReviewsProtected.DELETE("/:id", s.coachController.DeleteCoachReview)
				coachReviewsProtected.POST("/mark-helpful", s.coachController.MarkReviewHelpful)
				coachReviewsProtected.GET("/can-review", s.coachController.CheckCanReviewCoach)
			}
		}

		// 智能排課相關路由
		intelligentScheduling := v1.Group("/intelligent-scheduling")
		{
			// 公開路由
			intelligentScheduling.GET("/options", s.coachController.GetIntelligentSchedulingOptions)

			// 需要認證的路由
			intelligentSchedulingProtected := intelligentScheduling.Group("/")
			intelligentSchedulingProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				intelligentSchedulingProtected.POST("/recommendations", s.coachController.GetIntelligentRecommendations)
				intelligentSchedulingProtected.POST("/optimal-time", s.coachController.FindOptimalLessonTime)
				intelligentSchedulingProtected.POST("/detect-conflicts", s.coachController.DetectSchedulingConflicts)
				intelligentSchedulingProtected.POST("/resolve-conflict", s.coachController.ResolveSchedulingConflict)
				intelligentSchedulingProtected.POST("/coaches/:coachId/factors", s.coachController.GetCoachRecommendationFactors)
			}
		}

		// 配對相關路由
		matching := v1.Group("/matching")
		matching.Use(middleware.AuthMiddleware(s.jwtService))
		{
			matching.POST("/find", s.matchingController.FindMatches)
			matching.GET("/random", s.matchingController.FindRandomMatches)
			matching.GET("/reputation", s.matchingController.GetReputationScore)
			matching.GET("/history", s.matchingController.GetMatchingHistory)
			matching.POST("/create", s.matchingController.CreateMatch)
			matching.GET("/statistics", s.matchingController.GetMatchingStatistics)
			matching.PUT("/reputation/:userID", s.matchingController.UpdateReputation)

			// 抽卡配對相關路由
			matching.POST("/card-action", s.matchingController.ProcessCardAction)
			matching.GET("/card-history", s.matchingController.GetCardInteractionHistory)
			matching.GET("/notifications", s.matchingController.GetMatchNotifications)
			matching.PUT("/notifications/:notificationID/read", s.matchingController.MarkNotificationAsRead)
		}

		// 俱樂部相關路由
		clubs := v1.Group("/clubs")
		{
			clubs.GET("", s.getClubs)
			clubs.GET("/:id", s.getClub)
			clubs.POST("", s.createClub)
		}

		// 球拍相關路由
		rackets := v1.Group("/rackets")
		{
			// 公開路由
			rackets.GET("", s.racketController.SearchRackets)
			rackets.GET("/brands", s.racketController.GetAvailableBrands)
			rackets.GET("/specifications", s.racketController.GetRacketSpecifications)
			rackets.GET("/:id", s.racketController.GetRacket)
			rackets.GET("/:id/prices", s.racketController.GetRacketPrices)
			rackets.GET("/:id/reviews", s.racketController.GetRacketReviews)
			rackets.GET("/:id/reviews/statistics", s.racketController.GetRacketReviewStatistics)

			// 需要認證的路由
			racketsProtected := rackets.Group("/")
			racketsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				racketsProtected.POST("", s.racketController.CreateRacket)
				racketsProtected.PUT("/:id", s.racketController.UpdateRacket)
				racketsProtected.DELETE("/:id", s.racketController.DeleteRacket)
				racketsProtected.POST("/images", s.racketController.UploadRacketImages)
				racketsProtected.POST("/:id/prices", s.racketController.CreateRacketPrice)
				racketsProtected.POST("/:id/reviews", s.racketController.CreateRacketReview)
			}
		}

		// 球拍價格相關路由
		racketPrices := v1.Group("/racket-prices")
		racketPrices.Use(middleware.AuthMiddleware(s.jwtService))
		{
			racketPrices.PUT("/:priceId", s.racketController.UpdateRacketPrice)
			racketPrices.DELETE("/:priceId", s.racketController.DeleteRacketPrice)
			racketPrices.PUT("/:priceId/availability", s.racketController.UpdatePriceAvailability)
		}

		// 球拍評價相關路由
		racketReviews := v1.Group("/racket-reviews")
		racketReviews.Use(middleware.AuthMiddleware(s.jwtService))
		{
			racketReviews.POST("/:reviewId/helpful", s.racketController.MarkRacketReviewHelpful)
		}

		// 聊天相關路由
		chat := v1.Group("/chat")
		chat.Use(middleware.AuthMiddleware(s.jwtService))
		{
			// WebSocket 連接
			chat.GET("/ws", s.chatController.HandleWebSocket)

			// 聊天室管理
			chat.POST("/rooms", s.chatController.CreateChatRoom)
			chat.GET("/rooms", s.chatController.GetChatRooms)
			chat.GET("/rooms/:roomId", s.chatController.GetChatRoom)
			chat.POST("/rooms/:roomId/join", s.chatController.JoinChatRoom)
			chat.POST("/rooms/:roomId/leave", s.chatController.LeaveChatRoom)
			chat.POST("/rooms/:roomId/read", s.chatController.MarkMessagesAsRead)

			// 訊息管理
			chat.POST("/messages", s.chatController.SendMessage)
			chat.GET("/rooms/:roomId/messages", s.chatController.GetMessages)

			// 在線用戶
			chat.GET("/online-users", s.chatController.GetOnlineUsers)
			chat.GET("/rooms/:roomId/online-users", s.chatController.GetRoomUsers)
		}

		// 信譽評分相關路由
		reputation := v1.Group("/reputation")
		{
			// 公開路由
			reputation.GET("/leaderboard", s.reputationController.GetReputationLeaderboard)
			reputation.GET("/stats", s.reputationController.GetReputationStats)
			reputation.GET("/users/:userId/score", s.reputationController.GetUserReputationScore)

			// 需要認證的路由
			reputationProtected := reputation.Group("/")
			reputationProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				reputationProtected.GET("/users/:userId/history", s.reputationController.GetUserReputationHistory)
				reputationProtected.POST("/users/:userId/attendance", s.reputationController.RecordMatchAttendance)
				reputationProtected.POST("/users/:userId/punctuality", s.reputationController.RecordMatchPunctuality)
				reputationProtected.POST("/users/:userId/skill-accuracy", s.reputationController.RecordSkillLevelAccuracy)
				reputationProtected.POST("/users/:userId/behavior-review", s.reputationController.SubmitBehaviorReview)
				reputationProtected.POST("/users/:userId/update-ntrp", s.reputationController.UpdateUserNTRPLevel)
			}
		}

		// 配對統計相關路由
		matchStats := v1.Group("/match-statistics")
		{
			// 公開路由（部分受隱私設定限制）
			matchStats.GET("/users/:userId", s.matchStatisticsController.GetUserMatchStatistics)
			matchStats.GET("/users/:userId/history", s.matchStatisticsController.GetUserMatchHistory)
			matchStats.GET("/users/:userId/skill-progression", s.matchStatisticsController.GetSkillLevelProgression)
			matchStats.GET("/users/:userId/reputation", s.matchStatisticsController.GetReputationScoreWithPrivacy)

			// 需要認證的路由
			matchStatsProtected := matchStats.Group("/")
			matchStatsProtected.Use(middleware.AuthMiddleware(s.jwtService))
			{
				// 比賽結果記錄
				matchStatsProtected.POST("/matches/:matchId/result", s.matchStatisticsController.RecordMatchResult)
				matchStatsProtected.POST("/results/:resultId/confirm", s.matchStatisticsController.ConfirmMatchResult)
				matchStatsProtected.GET("/pending-confirmations", s.matchStatisticsController.GetMatchResultsForConfirmation)

				// 技術等級調整
				matchStatsProtected.POST("/users/:userId/adjust-skill-level", s.matchStatisticsController.ManuallyAdjustSkillLevel)

				// 隱私設定
				matchStatsProtected.GET("/privacy-settings", s.matchStatisticsController.GetUserPrivacySettings)
				matchStatsProtected.PUT("/privacy-settings", s.matchStatisticsController.UpdateUserPrivacySettings)

				// 統計摘要
				matchStatsProtected.GET("/summary", s.matchStatisticsController.GetMatchStatisticsSummary)
			}
		}
	}
}

// Start 啟動服務器
func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Port)
}

// healthCheck 健康檢查處理器
// @Summary 健康檢查
// @Description 檢查 API 服務器狀態
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Tennis Platform API is running",
		"version": "1.0.0",
	})
}

// 以下是其他功能的佔位符處理器，將在後續任務中實現

func (s *Server) getClubs(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (s *Server) getClub(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}

func (s *Server) createClub(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented yet"})
}
