package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

// TaskUsecase は課題管理に関するビジネスロジックを担当する構造体である．
type TaskUsecase struct {
	taskRepo  port.TaskRepository   // 課題データの永続化を担うリポジトリである．
	groupRepo port.GroupRepository  // グループ設定を取得するためのリポジトリである．
	aiService port.AIService        // AI による文章生成（リマインド等）を担うサービスである．
	scheduler port.SchedulerService // リマインド予約を管理するサービスである．
}

// NewTaskUsecase は TaskUsecase の新しいインスタンスを生成する．
func NewTaskUsecase(tr port.TaskRepository, gr port.GroupRepository, ai port.AIService, sch port.SchedulerService) *TaskUsecase {
	return &TaskUsecase{
		taskRepo:  tr,
		groupRepo: gr,
		aiService: ai,
		scheduler: sch,
	}
}

// RegisterManualTask は UI から直接入力された情報に基づいて課題を登録するユースケースである．
func (uc *TaskUsecase) RegisterManualTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	//　ドメイン層の初期設定とバリデーションを適用する（ダメなら即座に弾く）
	task.SetupManualDefaults()
	if err := task.Validate(); err != nil {
		return nil, err
	}

	// データの準備: 作成者 ID はドメインモデルにセットされた状態で渡される（Handler でセット）

	// データベースへ保存: 何よりも先にDB永続化を確定させる
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("手動タスクの保存に失敗した： %w", err)
	}

	// 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（タイマー予約など）
	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)
	if group != nil && !task.Deadline.IsZero() && task.Deadline.After(time.Now()) {
		hasPending := false
		for _, up := range task.UserProgress {
			if !up.IsCompleted {
				hasPending = true
				break
			}
		}
		if hasPending {
			for _, interval := range group.RemindIntervals {
				_ = uc.scheduler.ScheduleTaskRemind(ctx, task, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
			}
		}
	}

	// 結果の返却: 登録された課題オブジェクトを呼び出し元へ返す
	return task, nil
}

// ListGroupTasks は特定のグループの課題一覧を取得するユースケースである．
func (uc *TaskUsecase) ListGroupTasks(ctx context.Context, groupID string) ([]*domain.Task, error) {
	return uc.taskRepo.FindByGroupID(ctx, groupID)
}

// UpdateTask は課題の情報を更新するユースケースである．
func (uc *TaskUsecase) UpdateTask(ctx context.Context, taskID string, input *domain.Task, operatorID string) (*domain.Task, error) {
	existing, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("更新対象の課題が見つからない")
	}

	// 期限を過ぎた手動課題の編集を禁止する．
	if existing.Source == domain.SourceManual && !existing.Deadline.IsZero() && existing.Deadline.Before(time.Now()) {
		return nil, fmt.Errorf("期限を過ぎた手動課題は編集できない")
	}

	// 部屋の設定を取得
	group, _ := uc.groupRepo.FindByID(ctx, existing.GroupID)
	if group == nil {
		return nil, fmt.Errorf("所属グループが見つからない")
	}

	// 権限チェック
	// タイトルや期限の変更は，作成者 または グループオーナーのみ可能である．
	isCreator := existing.CreatorID == operatorID
	isOwner := group.OwnerID == operatorID

	// 秒単位での期限比較（タイムゾーンやミリ秒のズレによる誤判定を防ぐ）
	isDeadlineChanged := existing.Deadline.Unix() != input.Deadline.Unix()
	isSubjectChanged := existing.Subject != input.Subject
	isRecurrenceChanged := existing.Recurrence.Type != input.Recurrence.Type

	if input.Title != existing.Title || isDeadlineChanged || isSubjectChanged || isRecurrenceChanged {
		if existing.Source == domain.SourceManual {
			if !isCreator && !isOwner {
				return nil, fmt.Errorf("タイトルや期限、科目名、繰り返し設定の変更は，作成者またはオーナーのみ可能である")
			}
			existing.Title = input.Title
			existing.Deadline = input.Deadline
			existing.Recurrence = input.Recurrence
		} else {
			// LMS課題の場合、LMSで期限未定の場合のみUni-Steps上で期限を設定・変更可能
			if !existing.IsLMSDeadlineSet {
				existing.Deadline = input.Deadline
			}
		}
		// 科目名（授業名）は手動課題・LMS課題を問わず設定・更新可能
		existing.Subject = input.Subject
	}

	// 進捗状況（該当者）の更新は，グループメンバーであれば誰でも可能とする（仕様）．
	if input.UserProgress != nil {
		inputUserIDs := make(map[string]*domain.TaskUserProgress)
		for _, up := range input.UserProgress {
			inputUserIDs[up.UserID] = up
		}
		newProgress := []*domain.TaskUserProgress{}
		for _, ep := range existing.UserProgress {
			if _, ok := inputUserIDs[ep.UserID]; ok {
				newProgress = append(newProgress, ep)
				delete(inputUserIDs, ep.UserID)
			}
		}
		for _, up := range inputUserIDs {
			up.TaskID = taskID
			if up.UpdatedAt.IsZero() {
				up.UpdatedAt = time.Now()
			}
			newProgress = append(newProgress, up)
		}
		existing.UserProgress = newProgress
	}

	// 担当者が一人もいなくなった場合は，課題を自動的に削除する．
	if len(existing.UserProgress) == 0 {
		_ = uc.taskRepo.Delete(ctx, taskID)
		return nil, nil // 削除されたことを示すため nil を返す．
	}

	// データベースへ保存
	if err := uc.taskRepo.Save(ctx, existing); err != nil {
		return nil, err
	}

	// 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（リマインドの再予約・キャンセル）
	if !existing.Deadline.IsZero() {
		hasPending := false
		for _, up := range existing.UserProgress {
			if !up.IsCompleted {
				hasPending = true
				break
			}
		}
		if hasPending {
			for _, interval := range group.RemindIntervals {
				_ = uc.scheduler.ScheduleTaskRemind(ctx, existing, interval, group.AICharacter, existing.Deadline.Add(-time.Duration(interval)*time.Minute))
			}
		} else {
			_ = uc.scheduler.CancelTaskReminds(ctx, taskID)
		}
	} else {
		_ = uc.scheduler.CancelTaskReminds(ctx, taskID)
	}

	// 結果の返却
	return existing, nil
}

// DeleteTask は課題を削除する．作成者またはオーナーのみ可能である．
func (uc *TaskUsecase) DeleteTask(ctx context.Context, taskID string, operatorID string) error {
	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("課題が見つからない")
	}

	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)
	if group == nil {
		return fmt.Errorf("所属グループが見つからない")
	}

	if task.CreatorID != operatorID && group.OwnerID != operatorID {
		return fmt.Errorf("課題の削除は作成者またはオーナーのみ可能である")
	}

	// 全員の通知をキャンセル
	_ = uc.scheduler.CancelTaskReminds(ctx, taskID)

	return uc.taskRepo.Delete(ctx, taskID)
}

// ToggleUserCompletion は特定のユーザーの課題完了状態を反転させる．
func (uc *TaskUsecase) ToggleUserCompletion(ctx context.Context, taskID, userID string) error {
	task, err := uc.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("課題が見つからない")
	}

	// 期限を過ぎた手動課題の完了トグル操作を禁止する．
	if task.Source == domain.SourceManual && !task.Deadline.IsZero() && task.Deadline.Before(time.Now()) {
		return fmt.Errorf("期限を過ぎた手動課題の完了状態は変更できない")
	}

	group, _ := uc.groupRepo.FindByID(ctx, task.GroupID)

	found := false
	for i, up := range task.UserProgress {
		if up.UserID == userID {
			task.UserProgress[i].IsCompleted = !task.UserProgress[i].IsCompleted
			task.UserProgress[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		task.UserProgress = append(task.UserProgress, &domain.TaskUserProgress{
			TaskID:      taskID,
			UserID:      userID,
			IsCompleted: true,
			UpdatedAt:   time.Now(),
		})
	}

	// 全員が完了したか確認
	allCompleted := true
	for _, up := range task.UserProgress {
		if !up.IsCompleted {
			allCompleted = false
			break
		}
	}

	// データベースへ保存（更新された完了状態を保存）
	if err := uc.taskRepo.Save(ctx, task); err != nil {
		return err
	}

	// 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（タイマー解除または再設定）
	if allCompleted {
		// 全員完了した場合はタイマーをキャンセル
		_ = uc.scheduler.CancelTaskReminds(ctx, taskID)
	} else if !allCompleted && group != nil && !task.Deadline.IsZero() {
		// 未完了者がいる状態で完了状態が外れた（未完了に戻った）場合、リマインドを再設定
		for _, interval := range group.RemindIntervals {
			_ = uc.scheduler.ScheduleTaskRemind(ctx, task, interval, group.AICharacter, task.Deadline.Add(-time.Duration(interval)*time.Minute))
		}
	}

	// 結果の返却
	return nil
}
