package domain

import (
	"time"
)

// User はシステム内のユーザーを表す純粋なドメインエンティティである．
type User struct {
	ID                 string    // ユーザーを一意に識別するユニークID
	Name               string    // ユーザーの表示名
	Email              string    // 連絡先メールアドレス
	WebPushToken       string    // Web Push通知送信用サブスクリプション情報
	GoogleAccessToken  string    // Google Classroom API連携用OAuthアクセストークン
	GoogleRefreshToken string    // OAuthリフレッシュトークン
	GoogleTokenExpiry  time.Time // OAuthアクセストークン有効期限
	Groups             []*Group  // 所属しているグループ一覧
}
