package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	SourceManual          = "manual"           // 手動入力による登録
	SourceGoogleClassroom = "google_classroom" // Google Classroom からの取得
)

// RecurrenceSettings は課題の定期的な繰り返し配信設定である．
type RecurrenceSettings struct {
	Type        string      // 繰り返しの種別（none, weekly, biweekly, custom）
	CustomDates []time.Time // Typeがcustomの場合に指定される特定日付のリスト
}

// Task はグループ内で管理される課題を表す純粋なドメインエンティティである．
type Task struct {
	ID               string              // 課題を一意に識別するUUID
	GroupID          string              // 所属するグループのID
	Source           string              // 課題の登録元（manual, google_classroom等）
	ExternalID       string              // 外部LMS側の課題ID（重複登録防止用）
	CreatorID        string              // 課題を登録した作成者のユーザーID
	Title            string              // 課題のタイトル（件名）
	Subject          string              // 課題の科目名（コース名，例: "統計学", "英語コミュニケーション"）
	Deadline         time.Time           // 提出締切日時
	IsLMSDeadlineSet bool                // LMS上で正式に締切日時が設定されているか
	LMSUpdateTime    time.Time           // LMS側での最終更新日時
	Recurrence       RecurrenceSettings  // 課題の繰り返し設定
	UserProgress     []*TaskUserProgress // メンバー全員のこの課題に対する完了状態リスト
}

// TaskUserProgress はある課題に対する特定のユーザーの完了状況である．
type TaskUserProgress struct {
	TaskID      string    // 対象の課題ID
	UserID      string    // 対象のユーザーID
	IsCompleted bool      // 課題を完了したかどうかのフラグ
	UpdatedAt   time.Time // 完了状態が最後に更新された日時
}

// SetupManualDefaults は手動登録時の初期値設定を行う．
func (t *Task) SetupManualDefaults() {
	t.Source = SourceManual
	if t.ID == "" {
		id, err := uuid.NewV7()
		if err == nil {
			t.ID = id.String()
		} else {
			t.ID = uuid.New().String()
		}
	}
	if t.ExternalID == "" {
		t.ExternalID = t.ID
	}
	for _, up := range t.UserProgress {
		up.TaskID = t.ID
		if up.UpdatedAt.IsZero() {
			up.UpdatedAt = time.Now()
		}
	}
}

// Validate は課題のドメインルールに基づくバリデーションを行う．
func (t *Task) Validate() error {
	if t.Title == "" {
		return fmt.Errorf("タイトルは必須である")
	}
	return nil
}
