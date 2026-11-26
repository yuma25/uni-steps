package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yuma25/uni-steps/di"
)

func main() {
	// 0. タイムゾーンを日本時間 (JST) に設定する．
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("タイムゾーンの読み込みに失敗した（デフォルトを使用する）： %v\n", err)
	} else {
		time.Local = loc
	}

	// 1. 環境変数の読み込み (.env から一元管理)
	_ = godotenv.Load()

	// 2. DI コンテナによる全レイヤーの初期化
	container, err := di.NewContainer()
	if err != nil {
		log.Fatalf("DI コンテナの初期化に失敗した: %v", err)
	}

	// 3. 起動時に全グループのサマリー配信タイマーを初期予約する（常駐タイマー方式）
	if err := container.SummaryUsecase.ScheduleAllSummaries(context.Background()); err != nil {
		log.Printf("[Summary] サマリータイマー初期予約エラー: %v\n", err)
	}

	// 4. 6時間ごとに全グループの課題を自動同期するバックグラウンド処理
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			log.Println("[AutoSync] 自動一括同期処理を開始中...")
			if err := container.SyncUsecase.SyncAllGroups(ctx); err != nil {
				log.Printf("[AutoSync] 自動一括同期エラー: %v\n", err)
			} else {
				log.Println("[AutoSync] 自動一括同期処理が正常に完了した")
			}
		}
	}()

	// 5. Echo サーバーの初期化
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "{\n  \"time\": \"${time_rfc3339}\",\n  \"method\": \"${method}\",\n  \"uri\": \"${uri}\",\n  \"status\": ${status},\n  \"latency\": \"${latency_human}\"\n}\n",
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "https://uni-steps.vercel.app"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Uni-Steps API is running")
	})

	// 6. ハンドラーの登録
	container.RegisterHandlers(e)

	// 7. サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
