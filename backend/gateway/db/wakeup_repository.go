package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yuma25/uni-steps/domain"
	"github.com/yuma25/uni-steps/port"
)

type wakeupRepository struct {
	db *sql.DB
}

// NewWakeupRepository は WakeupRepository の新しいインスタンスを生成する．
func NewWakeupRepository(db *sql.DB) port.WakeupRepository {
	return &wakeupRepository{
		db: db,
	}
}

func (r *wakeupRepository) Save(ctx context.Context, check *domain.WakeupCheck) error {
	query := `
		INSERT INTO wakeup_checks (
			id, user_id, group_id, target_time, grace_minutes, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			group_id = EXCLUDED.group_id,
			target_time = EXCLUDED.target_time,
			grace_minutes = EXCLUDED.grace_minutes,
			status = EXCLUDED.status,
			created_at = EXCLUDED.created_at;
	`
	_, err := r.db.ExecContext(ctx, query,
		check.ID,
		check.UserID,
		check.GroupID,
		check.TargetTime,
		check.GraceMinutes,
		check.Status,
		check.CreatedAt,
	)
	return err
}

func (r *wakeupRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM wakeup_checks WHERE id = $1;`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *wakeupRepository) FindByID(ctx context.Context, id string) (*domain.WakeupCheck, error) {
	query := `
		SELECT id, user_id, group_id, target_time, grace_minutes, status, created_at
		FROM wakeup_checks
		WHERE id = $1;
	`
	var check domain.WakeupCheck
	var targetTime, createdAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&check.ID,
		&check.UserID,
		&check.GroupID,
		&targetTime,
		&check.GraceMinutes,
		&check.Status,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if targetTime.Valid {
		check.TargetTime = targetTime.Time
	}
	if createdAt.Valid {
		check.CreatedAt = createdAt.Time
	}

	return &check, nil
}

func (r *wakeupRepository) FindPendingByTime(ctx context.Context, now time.Time) ([]*domain.WakeupCheck, error) {
	query := `
		SELECT id, user_id, group_id, target_time, grace_minutes, status, created_at
		FROM wakeup_checks
		WHERE status = $1 AND target_time <= $2;
	`
	return r.queryWakeupChecks(ctx, query, domain.WakeupStatusPending, now)
}

func (r *wakeupRepository) FindActiveByUser(ctx context.Context, userID string) ([]*domain.WakeupCheck, error) {
	query := `
		SELECT id, user_id, group_id, target_time, grace_minutes, status, created_at
		FROM wakeup_checks
		WHERE user_id = $1 AND status = $2;
	`
	return r.queryWakeupChecks(ctx, query, userID, domain.WakeupStatusPending)
}

func (r *wakeupRepository) FindActiveByGroup(ctx context.Context, groupID string) ([]*domain.WakeupCheck, error) {
	now := time.Now().Local()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	query := `
		SELECT id, user_id, group_id, target_time, grace_minutes, status, created_at
		FROM wakeup_checks
		WHERE group_id = $1 AND target_time >= $2;
	`
	return r.queryWakeupChecks(ctx, query, groupID, todayStart)
}

func (r *wakeupRepository) queryWakeupChecks(ctx context.Context, query string, args ...any) ([]*domain.WakeupCheck, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*domain.WakeupCheck
	for rows.Next() {
		var check domain.WakeupCheck
		var targetTime, createdAt sql.NullTime

		if err := rows.Scan(
			&check.ID,
			&check.UserID,
			&check.GroupID,
			&targetTime,
			&check.GraceMinutes,
			&check.Status,
			&createdAt,
		); err != nil {
			return nil, err
		}
		if targetTime.Valid {
			check.TargetTime = targetTime.Time
		}
		if createdAt.Valid {
			check.CreatedAt = createdAt.Time
		}
		checks = append(checks, &check)
	}

	return checks, rows.Err()
}
