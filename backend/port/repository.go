package port

import (
	"context"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
)

// TaskRepository は課題データの永続化に関するインターフェース．
type TaskRepository interface {
	Save(ctx context.Context, task *domain.Task) error                                     // 課題情報の新規作成または更新（進捗状況も同期）
	FindByID(ctx context.Context, id string) (*domain.Task, error)                         // 課題IDから課題情報を1件取得
	FindByExternalID(ctx context.Context, externalID string) (*domain.Task, error)         // 外部LMSの課題IDから検索（重複登録防止用）
	FindByGroupID(ctx context.Context, groupID string) ([]*domain.Task, error)             // 指定グループに所属する全課題を締切順で取得
	FindApproachingDeadlines(ctx context.Context, until time.Time) ([]*domain.Task, error) // 指定日時までに提出期限が迫っている課題を取得
	Delete(ctx context.Context, id string) error                                           // 課題と紐づく全進捗データを完全に削除
}

// UserRepository はユーザーデータの永続化に関するインターフェース．
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error                         // ユーザー情報の新規作成または更新（Upsert）
	UpdateWebPushToken(ctx context.Context, userID string, token string) error // Web Push通知トークンのみを更新
	FindByID(ctx context.Context, id string) (*domain.User, error)             // ユーザーIDからユーザー情報を取得（所属グループ含む）
	FindByEmail(ctx context.Context, email string) (*domain.User, error)       // メールアドレスからユーザー情報を検索（ログイン認証用）
}

// GroupRepository はグループデータの永続化に関するインターフェース．
type GroupRepository interface {
	Save(ctx context.Context, group *domain.Group) error                      // グループ情報の新規作成または設定更新
	FindByID(ctx context.Context, id string) (*domain.Group, error)           // グループIDからグループ情報（メンバー含む）を取得
	FindAllGroups(ctx context.Context) ([]*domain.Group, error)               // システム内の全グループを取得（定期バッチ・自動同期用）
	FindByInviteCode(ctx context.Context, code string) (*domain.Group, error) // 8桁の招待コードからグループを検索
	FindByUserID(ctx context.Context, userID string) ([]*domain.Group, error) // 指定ユーザーが所属しているグループ一覧を取得
	AddUserToGroup(ctx context.Context, groupID string, userID string) error  // グループにユーザーをメンバーとして参加させる
	RemoveUser(ctx context.Context, groupID string, userID string) error      // グループから指定ユーザーを脱退させる
	Delete(ctx context.Context, id string) error                              // グループ本体とメンバー紐付けを完全に削除
}

// WakeupRepository は起床確認データの永続化に関するインターフェース．
type WakeupRepository interface {
	Save(ctx context.Context, check *domain.WakeupCheck) error                            // 起床見守りレコードの新規保存またはステータス更新
	Delete(ctx context.Context, id string) error                                          // 起床見守りレコードの削除（キャンセル時）
	FindPendingByTime(ctx context.Context, now time.Time) ([]*domain.WakeupCheck, error)  // 目標時刻＋猶予を過ぎても未起床（pending）のレコードを取得
	FindActiveByUser(ctx context.Context, userID string) ([]*domain.WakeupCheck, error)   // 特定ユーザーの現在進行中の起床見守りを取得
	FindActiveByGroup(ctx context.Context, groupID string) ([]*domain.WakeupCheck, error) // 指定グループの本日分の起床宣言・結果一覧を取得
}

// NotificationLogRepository は通知履歴データの永続化に関するインターフェース．
type NotificationLogRepository interface {
	Save(ctx context.Context, log *domain.NotificationLog) error                                     // 配信された新しい通知ログをデータベースに記録
	FindByGroupID(ctx context.Context, groupID string, limit int) ([]*domain.NotificationLog, error) // 指定グループの過去の通知ログを新着順に取得
}
