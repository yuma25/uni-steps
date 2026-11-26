package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/uni-steps/domain"
	"github.com/yuma25/uni-steps/port"
)

// SummaryUsecase はグループ全体の状況を要約し，通知するロジックを担当する．
type SummaryUsecase struct {
	groupRepo    port.GroupRepository
	taskRepo     port.TaskRepository
	aiService    port.AIService
	notifService port.NotificationService
	logRepo      port.NotificationLogRepository
	scheduler    port.SchedulerService
	loc          *time.Location // 日本時間 (JST)
}

// NewSummaryUsecase は SummaryUsecase の新しいインスタンスを生成する．
func NewSummaryUsecase(gr port.GroupRepository, tr port.TaskRepository, ai port.AIService, ns port.NotificationService, lr port.NotificationLogRepository, sch port.SchedulerService) *SummaryUsecase {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}

	return &SummaryUsecase{
		groupRepo:    gr,
		taskRepo:     tr,
		aiService:    ai,
		notifService: ns,
		logRepo:      lr,
		scheduler:    sch,
		loc:          loc,
	}
}

// ScheduleAllSummaries は全グループの定期サマリー配信タイマーを起動時に一括予約する．
func (uc *SummaryUsecase) ScheduleAllSummaries(ctx context.Context) error {
	groups, err := uc.groupRepo.FindAllGroups(ctx)
	if err != nil {
		return fmt.Errorf("全グループの取得に失敗した: %w", err)
	}

	for _, group := range groups {
		uc.ScheduleGroupSummary(ctx, group)
	}
	return nil
}

// ScheduleGroupSummary は特定グループの朝・夕サマリーの次回配信タイマーを予約する．
func (uc *SummaryUsecase) ScheduleGroupSummary(ctx context.Context, group *domain.Group) {
	if group.SummaryMorningTime != "" {
		uc.scheduleSingleSummary(ctx, group.ID, domain.SummaryTypeMorning, group.SummaryMorningTime)
	}
	if group.SummaryEveningTime != "" {
		uc.scheduleSingleSummary(ctx, group.ID, domain.SummaryTypeEvening, group.SummaryEveningTime)
	}
}

// scheduleSingleSummary は朝または夕の単一タイマーを予約し，発火時に単一再予約を行う．
func (uc *SummaryUsecase) scheduleSingleSummary(ctx context.Context, groupID string, summaryType string, timeStr string) {
	now := time.Now().In(uc.loc)
	runAt, err := uc.calcNextRunTime(timeStr, now)
	if err != nil {
		log.Printf("[Summary] 時刻パースエラー (GroupID: %s, Type: %s): %v\n", groupID, summaryType, err)
		return
	}

	err = uc.scheduler.ScheduleSummary(ctx, groupID, summaryType, runAt, func() {
		runCtx := context.Background()
		if err := uc.ProcessSingleGroupSummary(runCtx, groupID, summaryType); err != nil {
			log.Printf("[Summary] サマリー配信失敗 (GroupID: %s, Type: %s): %v\n", groupID, summaryType, err)
		}

		// 実行された側のみ、最新のグループ設定を取得して翌日分を単一再予約（二重予約防止）
		latestGroup, err := uc.groupRepo.FindByID(runCtx, groupID)
		if err != nil || latestGroup == nil {
			log.Printf("[Summary] 翌日再予約のためのグループ取得失敗 (GroupID: %s): %v\n", groupID, err)
			return
		}

		var nextTimeStr string
		if summaryType == domain.SummaryTypeMorning {
			nextTimeStr = latestGroup.SummaryMorningTime
		} else {
			nextTimeStr = latestGroup.SummaryEveningTime
		}

		if nextTimeStr != "" {
			uc.scheduleSingleSummary(runCtx, groupID, summaryType, nextTimeStr)
		}
	})
	if err != nil {
		log.Printf("[Summary] タイマー予約登録エラー (GroupID: %s, Type: %s): %v\n", groupID, summaryType, err)
	}
}

// CancelGroupSummary は特定グループのサマリー配信予約を全てキャンセルする．
func (uc *SummaryUsecase) CancelGroupSummary(ctx context.Context, groupID string) {
	if err := uc.scheduler.CancelSummary(ctx, groupID, domain.SummaryTypeMorning); err != nil {
		log.Printf("[Summary] 朝タイマーキャンセル失敗 (GroupID: %s): %v\n", groupID, err)
	}
	if err := uc.scheduler.CancelSummary(ctx, groupID, domain.SummaryTypeEvening); err != nil {
		log.Printf("[Summary] 夕タイマーキャンセル失敗 (GroupID: %s): %v\n", groupID, err)
	}
}

// calcNextRunTime は指定時刻文字列（"15:04" 形式）から，JST 基準で次回到達日時を算出する．
// 既に今日の指定時刻を過ぎている場合は，翌日の同時刻を返す．
func (uc *SummaryUsecase) calcNextRunTime(timeStr string, now time.Time) (time.Time, error) {
	parsed, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("時刻形式が不正です(HH:mm): %w", err)
	}

	nowInLoc := now.In(uc.loc)
	target := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), parsed.Hour(), parsed.Minute(), 0, 0, uc.loc)
	// 既に今現在よりも過去の場合は，翌日の同時刻にセットする
	if !target.After(nowInLoc) {
		target = target.AddDate(0, 0, 1)
	}
	return target, nil
}

// ProcessSingleGroupSummary は特定のグループに対してサマリーを生成・送信する．
func (uc *SummaryUsecase) ProcessSingleGroupSummary(ctx context.Context, groupID string, summaryType string) error {
	group, err := uc.groupRepo.FindByID(ctx, groupID)
	if err != nil || group == nil {
		return fmt.Errorf("グループが見つからない (ID: %s)", groupID)
	}

	// グループに紐づく全ての課題を取得
	tasks, err := uc.taskRepo.FindByGroupID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("課題一覧の取得に失敗した: %w", err)
	}

	// 「対象」の課題をメンバーごとに集計
	now := time.Now().In(uc.loc)
	tomorrow := now.AddDate(0, 0, 1)

	// 境界値判定の改善: 終了時刻を「翌々日の 00:00:00 JST」とし、半開区間 [targetStart, targetEnd) で判定
	targetStart := now
	targetEnd := time.Date(now.Year(), now.Month(), now.Day()+2, 0, 0, 0, 0, uc.loc)
	tomorrowStart := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, uc.loc)

	var typeLabel string
	if summaryType == domain.SummaryTypeMorning {
		typeLabel = fmt.Sprintf("%d/%d(今日) 〜 %d/%d(明日) が期限の課題", now.Month(), now.Day(), tomorrow.Month(), tomorrow.Day())
	} else {
		typeLabel = fmt.Sprintf("%d/%d(今夜残り) 〜 %d/%d(明日) が期限の課題", now.Month(), now.Day(), tomorrow.Month(), tomorrow.Day())
	}

	summaryItems := []string{}
	for _, user := range group.Users {
		var userTasks []string
		for _, task := range tasks {
			for _, up := range task.UserProgress {
				if up.UserID == user.ID && !up.IsCompleted {
					// 半開区間 [targetStart, targetEnd) による厳密な判定:
					// !deadline.Before(targetStart) && deadline.Before(targetEnd) でミリ秒ズレや境界値漏れを防止
					isTarget := !task.Deadline.IsZero() && !task.Deadline.Before(targetStart) && task.Deadline.Before(targetEnd)

					if isTarget {
						deadlineInLoc := task.Deadline.In(uc.loc)

						// 明日締切か今日締切かを判定
						var dayLabel string
						if deadlineInLoc.Before(tomorrowStart) {
							dayLabel = "今日"
						} else {
							dayLabel = "明日"
						}

						timeStr := fmt.Sprintf(" (%s %s)", dayLabel, deadlineInLoc.Format("15:04"))
						userTasks = append(userTasks, fmt.Sprintf("「%s」%s", task.Title, timeStr))
					}
				}
			}
		}

		if len(userTasks) > 0 {
			var items []string
			items = append(items, fmt.Sprintf("%s の課題:", user.Name))
			for _, task := range userTasks {
				items = append(items, fmt.Sprintf("- %s", task))
			}
			if len(userTasks) > 1 {
				items[len(items)-1] += fmt.Sprintf(" 計 %d 件", len(userTasks))
			}
			summaryItems = append(summaryItems, strings.Join(items, "\n"))
		} else {
			summaryItems = append(summaryItems, fmt.Sprintf("%s の課題: \n- 対象課題なし", user.Name))
		}
	}

	// データの準備: 集計テキストをプロンプト用に整形（時間帯の文脈を明記）
	var timeHeader string
	if summaryType == domain.SummaryTypeMorning {
		timeHeader = "【朝のサマリー】"
	} else {
		timeHeader = "【夜（夕方）のサマリー】"
	}

	statusText := strings.Join(summaryItems, "\n")
	content := fmt.Sprintf("%s %s の状況:\n%s", timeHeader, typeLabel, statusText)

	// 副作用の実行: 外部サービス連携（AI文章生成 ➔ メッセージ送信 ➔ 履歴保存）
	// Gemini AI にサマリー文を生成させる
	msg, err := uc.aiService.GenerateGroupSummaryMessage(ctx, content, group.AICharacter)
	if err != nil {
		log.Printf("[Summary] AI サマリー生成失敗（固定文へ切替，GroupID: %s）: %v\n", group.ID, err)
		msg = fmt.Sprintf("【%sサマリー】\n%s の状況は以下の通りです．\n\n%s\n\n一歩ずつ進めていきましょう！", summaryType, typeLabel, statusText)
	} else {
		msg = fmt.Sprintf("%s\n\n--- 課題詳細 ---\n%s", msg, statusText)
	}

	// LINE グループへ一斉配信（設定されている場合）
	if group.LineChannelToken != "" && group.LineGroupID != "" {
		if err := uc.notifService.SendGroupMessage(ctx, group.ID, msg); err != nil {
			log.Printf("[Summary] LINE グループ送信エラー (GroupID: %s): %v\n", group.ID, err)
		}
	}

	// Web Push でも全員に通知（スマホ・ブラウザ向け）
	for _, member := range group.Users {
		targetURL := fmt.Sprintf("/dashboard?user_id=%s&group_id=%s", member.ID, group.ID)
		if err := uc.notifService.SendDirectMessage(ctx, member.ID, msg, targetURL); err != nil {
			log.Printf("[Summary] Web Push 送信エラー (UserID: %s, GroupID: %s): %v\n", member.ID, group.ID, err)
		}
	}

	// データベースへ保存: 配信した履歴をデータベースに永続化
	if err := uc.logRepo.Save(ctx, &domain.NotificationLog{
		ID:        uuid.New().String(),
		GroupID:   group.ID,
		UserID:    "system",
		Type:      domain.NotificationTypeSummary,
		Message:   msg,
		CreatedAt: time.Now(),
	}); err != nil {
		log.Printf("[Summary] 通知ログ保存エラー (GroupID: %s): %v\n", group.ID, err)
	}

	log.Printf("[Summary] グループ %s への %s サマリー送信が完了した．\n", group.Name, summaryType)
	// ⑤ 結果の返却
	return nil
}
