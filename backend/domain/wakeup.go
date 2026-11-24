package domain

import (
	"time"
)

const (
	WakeupStatusPending   = "pending"   // 起床確認待ち
	WakeupStatusConfirmed = "confirmed" // 起床確認済み（成功）
	WakeupStatusAlerted   = "alerted"   // 時間超過・アラート送信済み（失敗）
)

// WakeupCheck は特定の時間の起床を監視・判定するための純粋なドメインエンティティである．
type WakeupCheck struct {
	ID           string    // 起床確認レコードを一意に識別するUUID
	UserID       string    // 起床宣言をした対象ユーザーのID
	GroupID      string    // 起床見守りを行う共有ルームのID
	TargetTime   time.Time // 目標起床日時
	GraceMinutes int       // 目標時刻超過後の猶予分数（例: 15）
	Status       string    // 現在のステータス（pending, confirmed, alerted）
	CreatedAt    time.Time // 起床宣言が登録された日時
}
