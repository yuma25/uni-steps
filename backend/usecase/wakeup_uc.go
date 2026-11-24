package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuma25/Uni-Steps/domain"
	"github.com/yuma25/Uni-Steps/port"
)

// WakeupUsecase は起床確認の登録や判定に関するビジネスロジックを担当する．
type WakeupUsecase struct {
	wakeupRepo port.WakeupRepository
	scheduler  port.SchedulerService
}

// NewWakeupUsecase は WakeupUsecase の新しいインスタンスを生成する．
func NewWakeupUsecase(wr port.WakeupRepository, sch port.SchedulerService) *WakeupUsecase {
	return &WakeupUsecase{
		wakeupRepo: wr,
		scheduler:  sch,
	}
}

// RequestWakeup は新しい起床見守りを予約し，同時にスケジューラーに SOS 発信を登録する．
func (uc *WakeupUsecase) RequestWakeup(ctx context.Context, userID, groupID string, targetTime time.Time, graceMinutes int) (*domain.WakeupCheck, error) {
	// 既存のアクティブな見守りがあればキャンセルして上書きする
	_ = uc.CancelWakeup(ctx, userID)

	// データの準備: 新規起床宣言のエンティティを作成
	check := &domain.WakeupCheck{
		ID:           uuid.New().String(),
		UserID:       userID,
		GroupID:      groupID,
		TargetTime:   targetTime,
		GraceMinutes: graceMinutes,
		Status:       domain.WakeupStatusPending,
		CreatedAt:    time.Now(),
	}

	// データベースへ保存: 何よりも先にDBに記録を保存する
	if err := uc.wakeupRepo.Save(ctx, check); err != nil {
		return nil, fmt.Errorf("起床確認の保存に失敗した： %w", err)
	}

	// 副作用の実行: DB保存が成功した後にだけ、外部連携を行う（スケジューラーに SOS 送信を予約）
	runAt := targetTime.Add(time.Duration(graceMinutes) * time.Minute)
	if err := uc.scheduler.ScheduleWakeupSOS(ctx, check.ID, userID, groupID, runAt); err != nil {
		return nil, fmt.Errorf("SOS 通知の予約に失敗した： %w", err)
	}

	//  結果の返却
	return check, nil
}

// CancelWakeup は進行中の見守りをキャンセルし，データベースから削除する．
func (uc *WakeupUsecase) CancelWakeup(ctx context.Context, userID string) error {
	checks, err := uc.wakeupRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, check := range checks {
		// スケジューラーの予約をキャンセル
		_ = uc.scheduler.CancelWakeupSOS(ctx, check.ID)
		// データベースから削除
		_ = uc.wakeupRepo.Delete(ctx, check.ID)
	}

	return nil
}

// ConfirmWakeup は本人の起床を確認し，予約されていた SOS 通知をキャンセルする．
func (uc *WakeupUsecase) ConfirmWakeup(ctx context.Context, userID string) error {
	// 現在進行中の（未確認の）起床確認を取得する．
	checks, err := uc.wakeupRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, check := range checks {
		//  データの準備: ステータスを「起床確認済み」に更新
		check.Status = domain.WakeupStatusConfirmed

		//  データベースへ保存: 更新後のステータスをDBに保存
		if err := uc.wakeupRepo.Save(ctx, check); err != nil {
			continue
		}

		//  副作用の実行: DB保存が成功した後にだけ、外部連携を行う（スケジューラーのSOS予約をキャンセル）
		_ = uc.scheduler.CancelWakeupSOS(ctx, check.ID)
	}

	//  結果の返却
	return nil
}

// GetActiveChecks はユーザーの進行中の起床確認を取得する．
func (uc *WakeupUsecase) GetActiveChecks(ctx context.Context, userID string) ([]*domain.WakeupCheck, error) {
	return uc.wakeupRepo.FindActiveByUser(ctx, userID)
}

// GetActiveGroupChecks は指定されたグループの本日分（または進行中）の起床確認一覧を取得する．
func (uc *WakeupUsecase) GetActiveGroupChecks(ctx context.Context, groupID string) ([]*domain.WakeupCheck, error) {
	return uc.wakeupRepo.FindActiveByGroup(ctx, groupID)
}
