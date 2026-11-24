package domain

import "time"

const (
	AICharacterDefault = "default" // 標準的な通知スタイル
	AICharacterStrict  = "strict"  // 厳しい指導官スタイル
	AICharacterKind    = "kind"    // 心配性な幼馴染スタイル
	AICharacterCool    = "cool"    // 冷徹な執事スタイル
)

const (
	SummaryTypeMorning = "morning" // 朝の課題サマリー配信
	SummaryTypeEvening = "evening" // 夕方の課題サマリー配信
)

// Group は課題や起床見守りを共有するルームを表す純粋なドメインエンティティである．
type Group struct {
	ID                 string    // グループを一意に識別するUUID
	Name               string    // グループ（部屋）の表示名
	OwnerID            string    // グループを作成・管理するオーナーユーザーのID
	LineChannelToken   string    // LINE公式アカウントのチャネルアクセストークン
	LineGroupID        string    // 通知送信先となるLINEグループのID
	LastSyncedAt       time.Time // 外部LMSと最後に同期を行った日時
	LMSLastUpdatedAt   time.Time // LMS側で最後に課題更新があった日時
	InviteCode         string    // 部屋に参加するための招待コード
	RemindIntervals    []int     // 締切前の通知間隔（分単位の配列，例: [1440, 60]）
	AICharacter        string    // 通知AIの性格タイプ（default, strict, kind, cool）
	SummaryMorningTime string    // 朝の課題サマリー送信時刻（例: "08:00"）
	SummaryEveningTime string    // 夕方の課題サマリー送信時刻（例: "20:00"）
	Users              []*User   // 参加しているメンバー一覧
}
