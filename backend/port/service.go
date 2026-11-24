package port

import (
	"context"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// AIService は AI による文章生成に関するインターフェースである．
type AIService interface {
	GenerateRemindMessage(ctx context.Context, task *domain.Task, style string) (string, error)            // 課題内容と指定キャラクターの性格に応じた通知文を生成
	GenerateGroupSummaryMessage(ctx context.Context, workloadSummary string, style string) (string, error) // 部屋全体の未完了課題サマリー文をキャラクター口調で生成
}

// LMSService は外部学習管理システム（LMS）連携に関するインターフェースである．
type LMSService interface {
	FetchTasks(ctx context.Context, userID string) ([]*domain.Task, error) // Google Classroom等の外部LMSから最新の課題一覧を取得
	GetProviderName() string                                               // LMSの提供元名（"google_classroom"等）を返却
}

// NotificationService は通知配信（LINE / Web Push）に関するインターフェースである．
type NotificationService interface {
	SendGroupMessage(ctx context.Context, targetID string, message string) error                  // LINEグループ等の共有チャネルへ一斉通知を送信
	SendDirectMessage(ctx context.Context, userID string, message string, targetURL string) error // 特定ユーザーのWeb Pushや個別LINEへ直接通知を送信
}

// SchedulerService は未来の処理予約に関するインターフェースである．
type SchedulerService interface {
	ScheduleTaskRemind(ctx context.Context, task *domain.Task, intervalMinutes int, style string, runAt time.Time) error // 締切前リマインドの自動送信タイマーを登録
	CancelTaskReminds(ctx context.Context, taskID string) error                                                          // 課題完了または削除時に予約済みリマインドタイマーを取消
	ScheduleWakeupSOS(ctx context.Context, wakeupID string, userID string, groupID string, runAt time.Time) error        // 目標時刻＋猶予時間超過時のSOS通知タイマーを登録
	CancelWakeupSOS(ctx context.Context, wakeupID string) error                                                          // 起床確認成功時に予約済みSOSタイマーを取消
	ScheduleSummary(ctx context.Context, groupID string, summaryType string, runAt time.Time, action func()) error       // 指定日時に定期サマリー配信を実行するタイマーを登録
	CancelSummary(ctx context.Context, groupID string, summaryType string) error                                         // 予約済みの定期サマリー配信タイマーを取消
}
