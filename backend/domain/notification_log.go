package domain

import (
	"time"
)

const (
	NotificationTypeRemind  = "remind"  // 通常の課題リマインド
	NotificationTypeSOS     = "sos"     // 起床見守り失敗時の SOS
	NotificationTypeSummary = "summary" // 朝夕の課題サマリー
)

// NotificationLog は送信された通知の履歴を保持する純粋なドメインエンティティである．
type NotificationLog struct {
	ID        string    // ログレコードを一意に識別するUUID
	GroupID   string    // 送信先のグループID
	UserID    string    // 通知のきっかけとなったユーザーID
	Type      string    // 通知の種別（remind, sos, summary）
	Message   string    // 実際に送信された通知文面
	CreatedAt time.Time // 送信日時
}
