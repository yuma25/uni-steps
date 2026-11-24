package di

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/labstack/echo/v4"
	"github.com/yuma25/Uni-Steps/gateway/ai"
	"github.com/yuma25/Uni-Steps/gateway/db"
	"github.com/yuma25/Uni-Steps/gateway/line"
	"github.com/yuma25/Uni-Steps/gateway/lms"
	"github.com/yuma25/Uni-Steps/gateway/notification"
	"github.com/yuma25/Uni-Steps/gateway/scheduler"
	"github.com/yuma25/Uni-Steps/gateway/webpush"
	"github.com/yuma25/Uni-Steps/handler"
	"github.com/yuma25/Uni-Steps/infra"
	"github.com/yuma25/Uni-Steps/port"
	"github.com/yuma25/Uni-Steps/usecase"
	"google.golang.org/api/option"
)

// Container はアプリケーションの全レイヤー（Gateway, Usecase, Handler）を保持する DI コンテナである．
type Container struct {
	DB             *sql.DB
	TaskUsecase    *usecase.TaskUsecase
	SyncUsecase    *usecase.SyncUsecase
	GroupUsecase   *usecase.GroupUsecase
	WakeupUsecase  *usecase.WakeupUsecase
	SummaryUsecase *usecase.SummaryUsecase
	UserRepo       port.UserRepository
	AIService      port.AIService
	NotifService   port.NotificationService
}

// NewContainer は全依存関係を注入・初期化して Container を生成する．
func NewContainer() (*Container, error) {
	// 1. データベース初期化
	sqlDB, err := infra.InitDB()
	if err != nil {
		return nil, err
	}

	// 2. Gateway (Repositories)
	taskRepo := db.NewTaskRepository(sqlDB)
	userRepo := db.NewUserRepository(sqlDB)
	groupRepo := db.NewGroupRepository(sqlDB)
	wakeupRepo := db.NewWakeupRepository(sqlDB)
	logRepo := db.NewNotificationLogRepository(sqlDB)

	// 3. Gateway (AI Service)
	genaiClient, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalf("Gemini クライアントの初期化に失敗した: %v", err)
	}
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}
	aiService := ai.NewGeminiService(genaiClient, modelName)

	// 4. Gateway (Notification Services)
	webPushService := webpush.NewWebPushService(userRepo, os.Getenv("VAPID_PUBLIC_KEY"), os.Getenv("VAPID_PRIVATE_KEY"), "mailto:admin@example.com")
	lineService := line.NewLineService(groupRepo)
	compositeNotifService := notification.NewCompositeNotificationService(lineService, webPushService)

	// 5. Gateway (LMS & Scheduler)
	oauthCfg := infra.NewGoogleOAuthConfig()
	lmsService := lms.NewGoogleClassroomService(userRepo, oauthCfg)
	schService := scheduler.NewInMemScheduler(taskRepo, userRepo, groupRepo, aiService, compositeNotifService, logRepo)

	// 6. Usecases
	summaryUsecase := usecase.NewSummaryUsecase(groupRepo, taskRepo, aiService, compositeNotifService, logRepo, schService)
	taskUsecase := usecase.NewTaskUsecase(taskRepo, groupRepo, aiService, schService)
	syncUsecase := usecase.NewSyncUsecase(taskRepo, groupRepo, lmsService, schService)
	groupUsecase := usecase.NewGroupUsecase(groupRepo, userRepo, logRepo, summaryUsecase)
	wakeupUsecase := usecase.NewWakeupUsecase(wakeupRepo, schService)

	return &Container{
		DB:             sqlDB,
		TaskUsecase:    taskUsecase,
		SyncUsecase:    syncUsecase,
		GroupUsecase:   groupUsecase,
		WakeupUsecase:  wakeupUsecase,
		SummaryUsecase: summaryUsecase,
		UserRepo:       userRepo,
		AIService:      aiService,
		NotifService:   compositeNotifService,
	}, nil
}

// RegisterHandlers は Echo インスタンスに全ハンドラーを登録する．
func (c *Container) RegisterHandlers(e *echo.Echo) {
	oauthCfg := infra.NewGoogleOAuthConfig()

	handler.NewTaskHandler(e, c.TaskUsecase, c.SyncUsecase)
	handler.NewNotificationHandler(e, c.UserRepo, c.AIService, c.NotifService)
	handler.NewAuthHandler(e, c.UserRepo, oauthCfg)
	handler.NewGroupHandler(e, c.GroupUsecase)
	handler.NewWakeupHandler(e, c.WakeupUsecase)
	handler.NewUserHandler(e, c.UserRepo)
}
